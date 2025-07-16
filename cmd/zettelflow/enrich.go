package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	openai "github.com/sashabaranov/go-openai"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var enrichCmd = &cobra.Command{
	Use:   "enrich [directory]",
	Short: "Enrich notes from the split stage with LLM-generated content.",
	Long:  `Processes all notes in the split directory, calls an LLM for each, and saves the results.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		checkAPIKey()
		pterm.DefaultBox.WithTitle("Enrich Stage").Println("Starting enrichment process...")

		// Print settings
		pterm.DefaultSection.Println("Using Enrich Settings")
		leveledList := pterm.LeveledList{
			{Level: 0, Text: fmt.Sprintf("Model: %s", viper.GetString("enrich.model"))},
			{Level: 0, Text: fmt.Sprintf("Temperature: %f", viper.GetFloat64("enrich.temperature"))},
			{Level: 0, Text: fmt.Sprintf("Max Tokens: %d", viper.GetInt("enrich.max_completion_tokens"))},
			{Level: 0, Text: fmt.Sprintf("Parallel Workers: %d", viper.GetInt("enrich.parallel"))},
		}
		pterm.DefaultTree.WithRoot(pterm.NewTreeFromLeveledList(leveledList)).Render()
		pterm.Println() // for spacing
		splitPath := viper.GetString("paths.split")
		if splitPath[0] == '~' {
			home, err := os.UserHomeDir()
			cobra.CheckErr(err)
			splitPath = filepath.Join(home, splitPath[1:])
		}

		enrichPath := viper.GetString("paths.enrich")
		if enrichPath[0] == '~' {
			home, err := os.UserHomeDir()
			cobra.CheckErr(err)
			enrichPath = filepath.Join(home, enrichPath[1:])
		}

		files, err := ioutil.ReadDir(splitPath)
		cobra.CheckErr(err)

		filesToProcess := []os.FileInfo{}
		for _, file := range files {
			if !file.IsDir() {
				filesToProcess = append(filesToProcess, file)
			}
		}

		if len(filesToProcess) == 0 {
			pterm.Info.Println("No notes to enrich in the split directory.")
			os.Exit(0)
		}

		// Load the enrich prompt
		promptPath := viper.GetString("paths.prompts")
		if promptPath[0] == '~' {
			home, err := os.UserHomeDir()
			cobra.CheckErr(err)
			promptPath = filepath.Join(home, promptPath[1:])
		}
		promptFile := filepath.Join(promptPath, "default_enrich.md")
		promptTemplate, err := ioutil.ReadFile(promptFile)
		cobra.CheckErr(err)

		client := openai.NewClient(viper.GetString("llm.api_key"))
		model := viper.GetString("enrich.model")
		if model == "" {
			pterm.Error.Println("Error: enrich model is not defined in the configuration.")
			os.Exit(1)
		}

		for _, file := range filesToProcess {
			expectedExt := viper.GetString("split.output_extension")
			if filepath.Ext(file.Name()) != expectedExt {
				continue
			}

			pterm.Info.Printf("Processing note: %s\n", file.Name())
			filePath := filepath.Join(splitPath, file.Name())
			originalContent, err := ioutil.ReadFile(filePath)
			cobra.CheckErr(err)

			// 1. Send the ENTIRE original content to the LLM
			finalPrompt := strings.Replace(string(promptTemplate), "{content}", string(originalContent), -1)
			req := openai.ChatCompletionRequest{
				Model:               model,
				Temperature:         float32(viper.GetFloat64("enrich.temperature")),
				MaxCompletionTokens: viper.GetInt("enrich.max_completion_tokens"),
				Messages: []openai.ChatCompletionMessage{
					{
						Role:    openai.ChatMessageRoleUser,
						Content: finalPrompt,
					},
				},
			}
			resp, err := client.CreateChatCompletion(context.Background(), req)
			cobra.CheckErr(err)

			llmResponse := resp.Choices[0].Message.Content

			// 2. ISOLATE the YAML from the LLM's response.
			var llmYAML string
			if start := strings.Index(llmResponse, "---"); start != -1 {
				if end := strings.Index(llmResponse[start+3:], "---"); end != -1 {
					llmYAML = llmResponse[start+3 : start+3+end]
				}
			}
			if llmYAML == "" {
				llmYAML = llmResponse
			}
			llmYAML = strings.TrimSpace(llmYAML)

			// 3. ISOLATE the original body content
			bodyStr := ""
			separator := "\n---\n"
			if strings.HasPrefix(string(originalContent), "---") {
				endOfFrontmatter := strings.Index(string(originalContent)[3:], separator)
				if endOfFrontmatter != -1 {
					endOfFrontmatter += 3
					bodyStr = string(originalContent)[endOfFrontmatter+len(separator):]
				}
			}
			if bodyStr == "" {
				bodyStr = string(originalContent)
			}

			// 4. Combine the new YAML from the LLM with the original body
			finalContent := fmt.Sprintf("---\n%s\n---\n%s", llmYAML, bodyStr)

			// 5. Save the final file
			fileName := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name())) + viper.GetString("split.output_extension")
			outputPath := filepath.Join(enrichPath, fileName)
			err = ioutil.WriteFile(outputPath, []byte(finalContent), 0644)
			cobra.CheckErr(err)
			pterm.Success.Printf("  - Saved enriched note to: %s\n", outputPath)
		}
		pterm.Success.Println("Enrich stage complete.")
		os.Exit(0)
	},
}

func init() {
	rootCmd.AddCommand(enrichCmd)
	enrichCmd.Flags().Int("parallel", 4, "Number of parallel workers")
	enrichCmd.Flags().String("filter", "", "Filter notes to enrich (e.g., tag==todo)")
}


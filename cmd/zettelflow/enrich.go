package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	openai "github.com/sashabaranov/go-openai"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var enrichCmd = &cobra.Command{
	Use:   "enrich [directory]",
	Short: "Enrich notes with LLM-generated content.",
	Long:  `Walks a directory of YAML files, calls an LLM for each, and merges the results.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		checkAPIKey()
		var directory string
		if len(args) > 0 {
			directory = args[0]
		} else {
			directory = viper.GetString("paths.split")
			if directory[0] == '~' {
				home, err := os.UserHomeDir()
				cobra.CheckErr(err)
				directory = filepath.Join(home, directory[1:])
			}
		}

		enrichPath := viper.GetString("paths.enrich")
		if enrichPath[0] == '~' {
			home, err := os.UserHomeDir()
			cobra.CheckErr(err)
			enrichPath = filepath.Join(home, enrichPath[1:])
		}

		fmt.Printf("Enriching notes in directory: %s\n", directory)

		files, err := ioutil.ReadDir(directory)
		cobra.CheckErr(err)

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
		model, _ := cmd.Flags().GetString("model")

		for _, file := range files {
			expectedExt := viper.GetString("split.output_extension")
			if file.IsDir() || filepath.Ext(file.Name()) != expectedExt {
				continue
			}

			fmt.Printf("Processing note: %s\n", file.Name())
			filePath := filepath.Join(directory, file.Name())
			originalContent, err := ioutil.ReadFile(filePath)
			cobra.CheckErr(err)

			// 1. Send the ENTIRE original content to the LLM
			finalPrompt := strings.Replace(string(promptTemplate), "{content}", string(originalContent), -1)
			req := openai.ChatCompletionRequest{
				Model: model,
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
				// Fallback for when the LLM doesn't return separators
				llmYAML = llmResponse
			}
			llmYAML = strings.TrimSpace(llmYAML)


			// 3. ISOLATE the original body content
			bodyStr := ""
			separator := "\n---\n"
			if strings.HasPrefix(string(originalContent), "---") {
				endOfFrontmatter := strings.Index(string(originalContent)[3:], separator)
				if endOfFrontmatter != -1 {
					endOfFrontmatter += 3 // The position is relative, so add back the length of the prefix we skipped
					bodyStr = string(originalContent)[endOfFrontmatter+len(separator):]
				}
			}
			if bodyStr == "" {
				// If we couldn't parse, fall back to using the whole original file as the body
				bodyStr = string(originalContent)
			}


			// 4. Combine the new YAML from the LLM with the original body
			finalContent := fmt.Sprintf("---\n%s\n---\n%s", llmYAML, bodyStr)

			// 5. Save the final file
			fileName := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name())) + viper.GetString("split.output_extension")
			outputPath := filepath.Join(enrichPath, fileName)
			err = ioutil.WriteFile(outputPath, []byte(finalContent), 0644)
			cobra.CheckErr(err)
			fmt.Printf("  - Saved enriched note to: %s\n", outputPath)
		}
	},
}

func init() {
	rootCmd.AddCommand(enrichCmd)
	enrichCmd.Flags().Int("parallel", 4, "Number of parallel workers")
	enrichCmd.Flags().String("filter", "", "Filter notes to enrich (e.g., tag==todo)")
	enrichCmd.Flags().StringP("model", "m", "gpt-4o-mini", "The model to use for enrichment")
}


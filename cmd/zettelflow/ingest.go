package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	openai "github.com/sashabaranov/go-openai"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var ingestCmd = &cobra.Command{
	Use:   "ingest [path]",
	Short: "Ingest a file or directory and process it with an LLM.",
	Long:  `Reads input from a file, a directory, or stdin, injects it into a prompt, and calls an OpenAI-compatible API.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		checkAPIKey()
		pterm.DefaultBox.WithTitle("Ingest Stage").Println("Starting ingestion process...")

		// Print settings
		pterm.DefaultSection.Println("Using Ingest Settings")
		leveledList := pterm.LeveledList{
			{Level: 0, Text: fmt.Sprintf("Model: %s", viper.GetString("ingest.model"))},
			{Level: 0, Text: fmt.Sprintf("Temperature: %f", viper.GetFloat64("ingest.temperature"))},
			{Level: 0, Text: fmt.Sprintf("Max Tokens: %d", viper.GetInt("ingest.max_completion_tokens"))},
		}
		pterm.DefaultTree.WithRoot(pterm.NewTreeFromLeveledList(leveledList)).Render()
		pterm.Println() // for spacing

		// Case 1: Path argument is provided (file or directory)
		if len(args) > 0 {
			path := args[0]
			info, err := os.Stat(path)
			cobra.CheckErr(err)

			if info.IsDir() {
				// Process a directory
				pterm.Info.Printf("Ingesting all files in directory: %s\n", path)
				files, err := ioutil.ReadDir(path)
				cobra.CheckErr(err)
				for _, file := range files {
					if !file.IsDir() {
						filePath := filepath.Join(path, file.Name())
						pterm.Debug.Printf("  - Processing file: %s\n", filePath)
						content, err := ioutil.ReadFile(filePath)
						cobra.CheckErr(err)
						processAndSave(string(content), cmd)
					}
				}
			} else {
				// Process a single file
				pterm.Info.Printf("Ingesting file: %s\n", path)
				content, err := ioutil.ReadFile(path)
				cobra.CheckErr(err)
				processAndSave(string(content), cmd)
			}
		} else {
			// Case 2 & 3: No path, check for piped input or start interactive mode
			stat, _ := os.Stdin.Stat()
			var inputText string
			if (stat.Mode() & os.ModeCharDevice) == 0 {
				// Piped input
				pterm.Info.Println("Ingesting from stdin...")
				scanner := bufio.NewScanner(os.Stdin)
				var builder strings.Builder
				for scanner.Scan() {
					builder.WriteString(scanner.Text())
					builder.WriteString("\n")
				}
				inputText = builder.String()
			} else {
				// Interactive prompt
				pterm.Info.Println("Enter text to ingest. Press Enter for a new line.")
				pterm.Info.Println("When you are finished, press Ctrl+D on a new, empty line.")
				scanner := bufio.NewScanner(os.Stdin)
				var builder strings.Builder
				for scanner.Scan() {
					builder.WriteString(scanner.Text())
					builder.WriteString("\n")
				}
				inputText = builder.String()
				if len(strings.TrimSpace(inputText)) == 0 {
					pterm.Warning.Println("No input provided. Exiting.")
					os.Exit(0)
				}
			}
			processAndSave(inputText, cmd)
		}
		pterm.Success.Println("Ingest stage complete.")
		os.Exit(0)
	},
}

// processAndSave contains the core logic for taking text, calling the LLM, and saving the result.
func processAndSave(inputText string, cmd *cobra.Command) {
	model := viper.GetString("ingest.model")
	if model == "" {
		pterm.Error.Println("Error: ingest model is not defined in the configuration.")
		os.Exit(1)
	}

	// If no prompt file is specified, use the default.
	promptFile, _ := cmd.Flags().GetString("prompt")
	if promptFile == "" {
		promptFile = viper.GetString("paths.prompts")
		// Expand ~
		if promptFile[0] == '~' {
			home, err := os.UserHomeDir()
			cobra.CheckErr(err)
			promptFile = filepath.Join(home, promptFile[1:])
		}
		promptFile = filepath.Join(promptFile, "default_ingest.md")
	}

	promptTemplate, err := ioutil.ReadFile(promptFile)
	cobra.CheckErr(err)

	finalPrompt := strings.Replace(string(promptTemplate), "{input_text}", inputText, -1)

	client := openai.NewClient(viper.GetString("llm.api_key"))
	req := openai.ChatCompletionRequest{
		Model:             model,
		Temperature:       float32(viper.GetFloat64("ingest.temperature")),
		MaxCompletionTokens: viper.GetInt("ingest.max_completion_tokens"),
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: finalPrompt,
			},
		},
		Stream: true,
	}

	pterm.Info.Println("Sending request to LLM...")
	stream, err := client.CreateChatCompletionStream(context.Background(), req)
	cobra.CheckErr(err)
	defer stream.Close()

	pterm.Println() // Add a newline for better formatting
	pterm.DefaultSection.Println("LLM Response")
	var responseBuilder strings.Builder
	for {
		response, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			pterm.Error.Printf("\nStream error: %v\n", err)
			os.Exit(1)
		}
		if len(response.Choices) > 0 {
			chunk := response.Choices[0].Delta.Content
			fmt.Print(pterm.LightCyan(chunk))
			responseBuilder.WriteString(chunk)
		}
	}
	pterm.Println() // Add a newline for better formatting
	pterm.DefaultSection.Println("End of Response")


	// Save the response
	ingestPath := viper.GetString("paths.ingest")
	// Expand ~
	if ingestPath[0] == '~' {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)
		ingestPath = filepath.Join(home, ingestPath[1:])
	}

	ts := time.Now().Format("20060102150405")
	outputFile := filepath.Join(ingestPath, fmt.Sprintf("ingest_%s.txt", ts))
	err = ioutil.WriteFile(outputFile, []byte(responseBuilder.String()), 0644)
	cobra.CheckErr(err)
	pterm.Success.Printf("Saved ingested text to: %s\n", outputFile)
}

func init() {
	rootCmd.AddCommand(ingestCmd)
	ingestCmd.Flags().StringP("prompt", "p", "", "Path to a custom prompt file")
}

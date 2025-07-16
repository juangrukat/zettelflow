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
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var ingestCmd = &cobra.Command{
	Use:   "ingest [file]",
	Short: "Ingest a file and process it with an LLM.",
	Long:  `Reads input from a file or stdin, injects it into a prompt, and calls an OpenAI-compatible API.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		checkAPIKey()

		var inputText string

		stat, _ := os.Stdin.Stat()
		if len(args) > 0 {
			// Read from file
			fmt.Printf("Ingesting file: %s\n", args[0])
			content, err := ioutil.ReadFile(args[0])
			cobra.CheckErr(err)
			inputText = string(content)
		} else if (stat.Mode() & os.ModeCharDevice) == 0 {
			// Read from stdin pipe
			fmt.Println("Ingesting from stdin...")
			scanner := bufio.NewScanner(os.Stdin)
			var builder strings.Builder
			for scanner.Scan() {
				builder.WriteString(scanner.Text())
				builder.WriteString("\n")
			}
			inputText = builder.String()
		} else {
			// Interactive prompt
			fmt.Println("Enter your text (press Ctrl+D when finished):")
			scanner := bufio.NewScanner(os.Stdin)
			var builder strings.Builder
			for scanner.Scan() {
				builder.WriteString(scanner.Text())
				builder.WriteString("\n")
			}
			inputText = builder.String()
			if len(strings.TrimSpace(inputText)) == 0 {
				fmt.Println("No input provided. Exiting.")
				return
			}
		}

		promptFile, _ := cmd.Flags().GetString("prompt")
		model, _ := cmd.Flags().GetString("model")

		// If no prompt file is specified, use the default.
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
			Model:     model,
			MaxTokens: 2000,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: finalPrompt,
				},
			},
			Stream: true,
		}
		stream, err := client.CreateChatCompletionStream(context.Background(), req)
		cobra.CheckErr(err)
		defer stream.Close()

		fmt.Println("--- LLM Response ---")
        var responseBuilder strings.Builder
        for {
            response, err := stream.Recv()
            if err == io.EOF {
                break
            }
            if err != nil {
                fmt.Printf("\nStream error: %v\n", err)
                os.Exit(1)
            }
            if len(response.Choices) > 0 {
                chunk := response.Choices[0].Delta.Content
                fmt.Printf(chunk)
                responseBuilder.WriteString(chunk)
            }
        }
        fmt.Println("\n--------------------")

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
		fmt.Println("Saved ingested text to:", outputFile)
	},
}

func init() {
	rootCmd.AddCommand(ingestCmd)
	ingestCmd.Flags().StringP("prompt", "p", "", "Path to a custom prompt file")
	ingestCmd.Flags().StringP("model", "m", "gpt-4o-mini", "The model to use for ingestion")
}

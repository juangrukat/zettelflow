package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var splitCmd = &cobra.Command{
	Use:   "split [file]",
	Short: "Split a file into multiple notes.",
	Long:  `Loads a file, splits it into chunks based on a delimiter, and generates YAML-fronted notes.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var inputFile string
		if len(args) > 0 {
			inputFile = args[0]
		} else {
			// Find the latest file in the ingest directory
			ingestPath := viper.GetString("paths.ingest")
			if ingestPath[0] == '~' {
				home, err := os.UserHomeDir()
				cobra.CheckErr(err)
				ingestPath = filepath.Join(home, ingestPath[1:])
			}
			files, err := ioutil.ReadDir(ingestPath)
			cobra.CheckErr(err)

			var latestFile os.FileInfo
			for _, file := range files {
				if latestFile == nil || file.ModTime().After(latestFile.ModTime()) {
					latestFile = file
				}
			}
			if latestFile == nil {
				fmt.Println("No files found in ingest directory. Run the 'ingest' command first.")
				return
			}
			inputFile = filepath.Join(ingestPath, latestFile.Name())
		}

		delimiter, _ := cmd.Flags().GetString("delimiter")
		preview, _ := cmd.Flags().GetBool("preview")

		fmt.Printf("Splitting file: %s by delimiter: '%s'\n", inputFile, delimiter)

		content, err := ioutil.ReadFile(inputFile)
		cobra.CheckErr(err)

		chunks := strings.Split(string(content), delimiter)
		fmt.Printf("Found %d chunks.\n", len(chunks))

		// Load and parse the YAML template
		templatePath := viper.GetString("paths.templates")
		if templatePath[0] == '~' {
			home, err := os.UserHomeDir()
			cobra.CheckErr(err)
			templatePath = filepath.Join(home, templatePath[1:])
		}
		templatePath = filepath.Join(templatePath, "note_header.yml")
		templateBytes, err := ioutil.ReadFile(templatePath)
		cobra.CheckErr(err)

		funcMap := template.FuncMap{
			// The template pipeline passes the last result as the LAST argument.
			// So, `{{ .Tags | join ", " }}` calls `join(", ", .Tags)`.
			// We need a function that takes the separator first, then the slice.
			"join": func(sep string, a []string) string {
				return strings.Join(a, sep)
			},
		}

		tmpl, err := template.New("note").Funcs(funcMap).Parse(string(templateBytes))
		cobra.CheckErr(err)

		splitPath := viper.GetString("paths.split")
		if splitPath[0] == '~' {
			home, err := os.UserHomeDir()
			cobra.CheckErr(err)
			splitPath = filepath.Join(home, splitPath[1:])
		}

		type NoteData struct {
			Content string
			Date    string
			Title   string
			Tags    []string
		}

		for i, chunk := range chunks {
			chunk = strings.TrimSpace(chunk)
			if chunk == "" {
				continue
			}

			data := NoteData{
				Content: chunk,
				Date:    time.Now().Format("2006-01-02"),
				// Title and Tags are left blank for the enrich stage
			}

			var buf bytes.Buffer
			err := tmpl.Execute(&buf, data)
			cobra.CheckErr(err)

			if preview {
				fmt.Printf("--- Chunk %d ---\n", i+1)
				fmt.Println(buf.String())
			} else {
				ts := time.Now().Format("20060102150405")
				outputFile := filepath.Join(splitPath, fmt.Sprintf("note_%s_%d%s", ts, i+1, viper.GetString("split.output_extension")))
				err = ioutil.WriteFile(outputFile, buf.Bytes(), 0644)
				cobra.CheckErr(err)
				fmt.Println("Created note:", outputFile)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(splitCmd)
	splitCmd.Flags().StringP("delimiter", "d", "###", "The delimiter to split the file by")
	splitCmd.Flags().Bool("preview", false, "Preview the split without writing files")
}


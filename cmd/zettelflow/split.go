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

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var splitCmd = &cobra.Command{
	Use:   "split",
	Short: "Split files from the ingest stage into multiple notes.",
	Long:  `Processes all files in the ingest directory, splits them into chunks, and generates YAML-fronted notes.`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		pterm.DefaultBox.WithTitle("Split Stage").Println("Starting split process...")
		ingestPath := viper.GetString("paths.ingest")
		if ingestPath[0] == '~' {
			home, err := os.UserHomeDir()
			cobra.CheckErr(err)
			ingestPath = filepath.Join(home, ingestPath[1:])
		}
		processedPath := filepath.Join(ingestPath, "processed")
		cobra.CheckErr(os.MkdirAll(processedPath, 0755))

		files, err := ioutil.ReadDir(ingestPath)
		cobra.CheckErr(err)

		filesToProcess := []os.FileInfo{}
		for _, file := range files {
			if !file.IsDir() {
				filesToProcess = append(filesToProcess, file)
			}
		}

		if len(filesToProcess) == 0 {
			pterm.Info.Println("No files to split in the ingest directory.")
			os.Exit(0)
		}

		for _, file := range filesToProcess {
			inputFile := filepath.Join(ingestPath, file.Name())
			delimiter, _ := cmd.Flags().GetString("delimiter")
			preview, _ := cmd.Flags().GetBool("preview")

			pterm.Info.Printf("Splitting file: %s by delimiter: '%s'\n", file.Name(), delimiter)

			content, err := ioutil.ReadFile(inputFile)
			cobra.CheckErr(err)

			chunks := strings.Split(string(content), delimiter)
			pterm.Debug.Printf("Found %d chunks.\n", len(chunks))

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
				}

				var buf bytes.Buffer
				err := tmpl.Execute(&buf, data)
				cobra.CheckErr(err)

				if preview {
                    pterm.NewStyle(pterm.FgLightCyan, pterm.Bold).Printf("--- Chunk %d Preview ---\n", i+1)
                    pterm.Println(buf.String())
                } else {
					ts := time.Now().Format("20060102150405")
					outputFile := filepath.Join(splitPath, fmt.Sprintf("note_%s_%d%s", ts, i+1, viper.GetString("split.output_extension")))
					err = ioutil.WriteFile(outputFile, buf.Bytes(), 0644)
					cobra.CheckErr(err)
					pterm.Success.Printf("  - Created note: %s\n", outputFile)
				}
			}

			// Move the processed file
			if !preview {
				destPath := filepath.Join(processedPath, file.Name())
				pterm.Debug.Printf("  - Moving processed file to: %s\n", destPath)
				err = os.Rename(inputFile, destPath)
				cobra.CheckErr(err)
			}
		}
		pterm.Success.Println("Split stage complete.")
		os.Exit(0)
	},
}

func init() {
	rootCmd.AddCommand(splitCmd)
	splitCmd.Flags().StringP("delimiter", "d", "###", "The delimiter to split the file by")
	splitCmd.Flags().Bool("preview", false, "Preview the split without writing files")
}


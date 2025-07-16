package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var dryRun bool

// cleanCmd represents the base for all clean subcommands
var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean processed files from one or all stages.",
	Long:  `Removes files from the data directories of specified processing stages.`,
}

var cleanIngestCmd = &cobra.Command{
	Use:   "ingest",
	Short: "Clean the ingest stage directory",
	Run: func(cmd *cobra.Command, args []string) {
		cleanStage("ingest")
	},
}

var cleanSplitCmd = &cobra.Command{
	Use:   "split",
	Short: "Clean the split stage directory",
	Run: func(cmd *cobra.Command, args []string) {
		cleanStage("split")
	},
}

var cleanEnrichCmd = &cobra.Command{
	Use:   "enrich",
	Short: "Clean the enrich stage directory",
	Run: func(cmd *cobra.Command, args []string) {
		cleanStage("enrich")
	},
}

var cleanAllCmd = &cobra.Command{
	Use:   "all",
	Short: "Clean all stage directories",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Cleaning all stages...")
		cleanStage("ingest")
		cleanStage("split")
		cleanStage("enrich")
	},
}

func cleanStage(stageName string) {
	stagePath := viper.GetString("paths." + stageName)
	if stagePath == "" {
		fmt.Fprintf(os.Stderr, "Error: stage '%s' not found in configuration.\n", stageName)
		os.Exit(1)
	}

	if stagePath[0] == '~' {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)
		stagePath = filepath.Join(home, stagePath[1:])
	}

	fmt.Printf("Cleaning stage '%s' at path: %s\n", stageName, stagePath)
	files, err := ioutil.ReadDir(stagePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading directory %s: %v\n", stagePath, err)
		os.Exit(1)
	}

	if len(files) == 0 {
		fmt.Println("  - Directory is already empty.")
		return
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		filePath := filepath.Join(stagePath, file.Name())
		if dryRun {
			fmt.Printf("  - [Dry Run] Would delete file: %s\n", filePath)
		} else {
			fmt.Printf("  - Deleting file: %s\n", filePath)
			err := os.Remove(filePath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "    - Error deleting file: %v\n", err)
			}
		}
	}
	fmt.Printf("  - Stage '%s' cleaned successfully.\n", stageName)
}

func init() {
	rootCmd.AddCommand(cleanCmd)
	cleanCmd.AddCommand(cleanIngestCmd)
	cleanCmd.AddCommand(cleanSplitCmd)
	cleanCmd.AddCommand(cleanEnrichCmd)
	cleanCmd.AddCommand(cleanAllCmd)

	// Add the dry-run flag to the parent 'clean' command so all subcommands inherit it.
	cleanCmd.PersistentFlags().BoolVarP(&dryRun, "dry-run", "d", false, "Perform a dry run without deleting files")
}

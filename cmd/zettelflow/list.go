package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// listCmd represents the base for all list subcommands
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List processed files from one or all stages.",
	Long:  `Lists the files in the data directories of specified processing stages.`,
}

var listIngestCmd = &cobra.Command{
	Use:   "ingest",
	Short: "List files in the ingest stage directory",
	Run: func(cmd *cobra.Command, args []string) {
		listStage("ingest")
	},
}

var listSplitCmd = &cobra.Command{
	Use:   "split",
	Short: "List files in the split stage directory",
	Run: func(cmd *cobra.Command, args []string) {
		listStage("split")
	},
}

var listEnrichCmd = &cobra.Command{
	Use:   "enrich",
	Short: "List files in the enrich stage directory",
	Run: func(cmd *cobra.Command, args []string) {
		listStage("enrich")
	},
}

var listAllCmd = &cobra.Command{
	Use:   "all",
	Short: "List files in all stage directories",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Listing files in all stages...")
		listStage("ingest")
		listStage("split")
		listStage("enrich")
	},
}

func listStage(stageName string) {
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

	fmt.Printf("--- Stage: %s ---\n", stageName)
	files, err := ioutil.ReadDir(stagePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading directory %s: %v\n", stagePath, err)
		os.Exit(1)
	}

	if len(files) == 0 {
		fmt.Println("  (No files found)")
		return
	}

	for _, file := range files {
		if !file.IsDir() {
			fmt.Printf("  - %s\n", file.Name())
		}
	}
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.AddCommand(listIngestCmd)
	listCmd.AddCommand(listSplitCmd)
	listCmd.AddCommand(listEnrichCmd)
	listCmd.AddCommand(listAllCmd)
}

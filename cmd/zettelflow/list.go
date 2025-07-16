package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List processed files in a stage.",
	Long:  `Lists the files that have been processed in a specific stage (ingest, split, enrich).`,
	Run: func(cmd *cobra.Command, args []string) {
		stage, _ := cmd.Flags().GetString("stage")
		output, _ := cmd.Flags().GetString("output")
		fmt.Printf("Listing files for stage: %s in %s format\n", stage, output)
		// Placeholder for list logic
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().String("stage", "enrich", "The processing stage to list files from (ingest, split, enrich)")
	listCmd.Flags().String("json", "", "Output in JSON format")
}

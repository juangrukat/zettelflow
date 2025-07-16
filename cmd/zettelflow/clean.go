package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean processed files from a stage.",
	Long:  `Removes files from a specified processing stage.`,
	Run: func(cmd *cobra.Command, args []string) {
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			fmt.Println("Dry run: Cleaning files (not actually deleting)")
		} else {
			fmt.Println("Cleaning files")
		}
		// Placeholder for clean logic
	},
}

func init() {
	rootCmd.AddCommand(cleanCmd)
	cleanCmd.Flags().Bool("dry-run", false, "Perform a dry run without deleting files")
}

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage zettelflow configuration",
	Long:  `Provides utilities to interact with the zettelflow configuration files.`,
}

var pathCmd = &cobra.Command{
	Use:   "path",
	Short: "Prints the path to the configuration directory",
	Long:  `Prints the absolute path to the zettelflow configuration directory.`,
	Run: func(cmd *cobra.Command, args []string) {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)
		configPath := filepath.Join(home, ".config", "zettelflow")
		fmt.Println(configPath)
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(pathCmd)
}

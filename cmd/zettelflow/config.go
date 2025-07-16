package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/user/zettelflow"
)

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// First run? Create default config files.
	zettelflow.FirstRunInit()

	// Find home directory.
	home, err := os.UserHomeDir()
	cobra.CheckErr(err)

	// Search config in home directory with name ".zettelflow" (without extension).
	configPath := filepath.Join(home, ".config", "zettelflow")
	viper.AddConfigPath(configPath)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

// checkAPIKey ensures that the OpenAI API key is set, prompting the user if it's not.
func checkAPIKey() {
	if !viper.IsSet("llm.api_key") || viper.GetString("llm.api_key") == "" {
		fmt.Println("OpenAI API key not found.")
		fmt.Print("Please enter your API key: ")

		reader := bufio.NewReader(os.Stdin)
		apiKey, err := reader.ReadString('\n')
		cobra.CheckErr(err)
		apiKey = strings.TrimSpace(apiKey)

		viper.Set("llm.api_key", apiKey)

		configPath := viper.ConfigFileUsed()
		if configPath == "" {
			home, err := os.UserHomeDir()
			cobra.CheckErr(err)
			configPath = filepath.Join(home, ".config", "zettelflow", "config.yaml")
		}

		// Create the directory if it doesn't exist
		configDir := filepath.Dir(configPath)
		if _, err := os.Stat(configDir); os.IsNotExist(err) {
			err = os.MkdirAll(configDir, 0755)
			cobra.CheckErr(err)
		}

		if err := viper.WriteConfigAs(configPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing config file: %s\n", err)
			os.Exit(1)
		}
		fmt.Println("API key saved to", configPath)
	}
}

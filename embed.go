package zettelflow

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

//go:embed all:assets
var assets embed.FS

func FirstRunInit() {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get home directory: %v\n", err)
		os.Exit(1)
	}

	configDir := filepath.Join(home, ".config", "zettelflow")
	promptsDir := filepath.Join(configDir, "prompts")
	configFilePath := filepath.Join(configDir, "config.yaml")

	// If the prompts directory doesn't exist, we assume it's a first run or a broken install.
	if _, err := os.Stat(promptsDir); os.IsNotExist(err) {
		fmt.Println("First run or incomplete configuration detected. Initializing...")

		// --- Create config directories ---
		templateDir := filepath.Join(configDir, "templates")
		if err := os.MkdirAll(promptsDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create prompts directory: %v\n", err)
			os.Exit(1)
		}
		if err := os.MkdirAll(templateDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create templates directory: %v\n", err)
			os.Exit(1)
		}

		// --- Write default config.yaml ---
		writeFileFromEmbed("assets/config.yaml", configFilePath)

		// --- Write default prompts ---
		writeFileFromEmbed("assets/prompts/default_ingest.md", filepath.Join(promptsDir, "default_ingest.md"))
		writeFileFromEmbed("assets/prompts/default_enrich.md", filepath.Join(promptsDir, "default_enrich.md"))

		// --- Write default template ---
		writeFileFromEmbed("assets/yaml_templates/note_header.yml", filepath.Join(templateDir, "note_header.yml"))

		// --- Create data directories ---
		fmt.Println("Reading configuration to create data directories...")
		viper.SetConfigFile(configFilePath)
		if err := viper.ReadInConfig(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read new config file: %v\n", err)
			os.Exit(1)
		}

		dataDirs := []string{
			viper.GetString("paths.ingest"),
			viper.GetString("paths.split"),
			viper.GetString("paths.enrich"),
			viper.GetString("paths.logs"),
		}
		for _, dir := range dataDirs {
			// Expand ~
			if dir[0] == '~' {
				dir = filepath.Join(home, dir[1:])
			}
			if err := os.MkdirAll(dir, 0755); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to create data directory %s: %v\n", dir, err)
				os.Exit(1)
			}
			fmt.Println("Created data directory:", dir)
		}
		fmt.Println("Initialization complete.")
	}
}

// writeFileFromEmbed reads a file from the embedded assets and writes it to the destination path.
func writeFileFromEmbed(sourcePath, destPath string) {
	content, err := assets.ReadFile(sourcePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read embedded file %s: %v\n", sourcePath, err)
		os.Exit(1)
	}
	if err := os.WriteFile(destPath, content, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write file to %s: %v\n", destPath, err)
		os.Exit(1)
	}
	fmt.Println("Created default file:", destPath)
}
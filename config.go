package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	Port         string `json:"port"`
	DBFile       string `json:"db_file"`
	BackupPath   string `json:"backup_path"`
	GitSync      bool   `json:"git_sync"`
	GitRemoteURL string `json:"git_remote_url"`
}

var GlobalConfig Config

func LoadConfig() {
	configPath := "config.json"

	// Set Defaults
	home, _ := os.UserHomeDir()
	GlobalConfig = Config{
		Port:       "8080",
		DBFile:     "recipes.db",
		BackupPath: filepath.Join(home, "morris_recipe_vault"),
		GitSync:    true,
	}

	// Load from file if it exists
	if _, err := os.Stat(configPath); err == nil {
		file, _ := os.ReadFile(configPath)
		if err := json.Unmarshal(file, &GlobalConfig); err != nil {
			fmt.Printf(" [Config] Warning: Could not parse config.json: %v\n", err)
		} else {
			fmt.Println(" [Config]: Loaded settings from config.json")
		}
	} else {
		// Create default config file if it doesn't exist
		SaveConfig()
		fmt.Println(" [Config]: Created default config.json")
	}

	// Allow Environment Variables to Override
	if envPort := os.Getenv("RECIPE_PORT"); envPort != "" {
		GlobalConfig.Port = envPort
	}
	if envDB := os.Getenv("RECIPE_DB_FILE"); envDB != "" {
		GlobalConfig.DBFile = envDB
	}
	if envPath := os.Getenv("RECIPE_BACKUP_PATH"); envPath != "" {
		GlobalConfig.BackupPath = envPath
	}
	if envSync := os.Getenv("RECIPE_GIT_SYNC"); envSync != "" {
		GlobalConfig.GitSync = envSync != "false"
	}
	if envRemote := os.Getenv("RECIPE_GIT_REMOTE"); envRemote != "" {
		GlobalConfig.GitRemoteURL = envRemote
	}
}

func SaveConfig() {
	configPath := "config.json"
	file, _ := json.MarshalIndent(GlobalConfig, "", "  ")
	os.WriteFile(configPath, file, 0644)
}

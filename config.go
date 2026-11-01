package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Config holds server settings. Loaded from config.json (or a test path)
// and overridable via environment variables: RECIPE_PORT, RECIPE_DB_FILE,
// RECIPE_BACKUP_PATH, RECIPE_GIT_SYNC, RECIPE_GIT_REMOTE.
type Config struct {
	Port         string `json:"port"`
	DBFile       string `json:"db_file"`
	BackupPath   string `json:"backup_path"`
	GitSync      bool   `json:"git_sync"`
	GitRemoteURL string `json:"git_remote_url"`
}

// GlobalConfig is the process-wide configuration, populated by LoadConfig.
var GlobalConfig Config

const configFileName = "config.json"

// LoadConfig loads config.json (creating a default one on first run), applies
// environment-variable overrides, and validates the result. It returns an
// error if the file is unreadable, malformed, or fails validation.
func LoadConfig() error {
	return loadConfig(configFileName)
}

// loadConfig is LoadConfig against an explicit path, so tests can exercise it
// without touching the working directory's real config.json.
func loadConfig(path string) error {
	home, _ := os.UserHomeDir()
	GlobalConfig = Config{
		Port:       "8080",
		DBFile:     "recipes.db",
		BackupPath: filepath.Join(home, "morris_recipe_vault"),
		GitSync:    false,
	}

	data, err := os.ReadFile(path)
	switch {
	case err == nil:
		if err := json.Unmarshal(data, &GlobalConfig); err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}
		// Normalize permissions so a git remote URL (which may carry a token)
		// is never world-readable.
		if err := os.Chmod(path, 0600); err != nil {
			return fmt.Errorf("chmod %s: %w", path, err)
		}
		fmt.Println(" [Config]: Loaded settings from", path)
	case os.IsNotExist(err):
		if err := saveConfig(path); err != nil {
			return fmt.Errorf("create default %s: %w", path, err)
		}
		fmt.Println(" [Config]: Created default", path)
	default:
		return fmt.Errorf("read %s: %w", path, err)
	}

	applyEnvOverrides()
	if err := GlobalConfig.Validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}
	return nil
}

// SaveConfig writes GlobalConfig to config.json with owner-only permissions.
// It returns an error if the marshal or write fails.
func SaveConfig() error {
	return saveConfig(configFileName)
}

// saveConfig is SaveConfig against an explicit path.
func saveConfig(path string) error {
	data, err := json.MarshalIndent(GlobalConfig, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

// applyEnvOverrides layers environment variables on top of the file values.
func applyEnvOverrides() {
	if v := os.Getenv("RECIPE_PORT"); v != "" {
		GlobalConfig.Port = v
	}
	if v := os.Getenv("RECIPE_DB_FILE"); v != "" {
		GlobalConfig.DBFile = v
	}
	if v := os.Getenv("RECIPE_BACKUP_PATH"); v != "" {
		GlobalConfig.BackupPath = v
	}
	if v := os.Getenv("RECIPE_GIT_SYNC"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			GlobalConfig.GitSync = b
		} else {
			fmt.Printf(" [Config] Warning: RECIPE_GIT_SYNC=%q is not a bool (use true/false)\n", v)
		}
	}
	if v := os.Getenv("RECIPE_GIT_REMOTE"); v != "" {
		GlobalConfig.GitRemoteURL = v
	}
}

// Validate checks the config fields for basic sanity. It returns the first
// problem found, or nil when the config is usable.
func (c Config) Validate() error {
	if c.Port == "" {
		return errors.New("port is empty")
	}
	if p, err := strconv.Atoi(c.Port); err != nil || p < 1 || p > 65535 {
		return fmt.Errorf("port %q is not a valid TCP port", c.Port)
	}
	if strings.TrimSpace(c.DBFile) == "" {
		return errors.New("db_file is empty")
	}
	if c.GitSync && strings.TrimSpace(c.BackupPath) == "" {
		return errors.New("backup_path is required when git_sync is enabled")
	}
	if strings.ContainsAny(c.GitRemoteURL, " \t\r\n") {
		return errors.New("git_remote_url must not contain whitespace")
	}
	return nil
}

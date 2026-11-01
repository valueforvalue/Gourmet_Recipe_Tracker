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
	// AutoImport, when true, imports the vault snapshot on startup if the DB
	// is empty. Caveat: a deliberately emptied DB is re-populated from
	// recipes.json on the next start — delete the snapshot too to start fresh.
	AutoImport bool `json:"auto_import"`
	// TestMode, when true, forces git sync against a disposable test repo and
	// a separate test vault directory so sync testing never touches real
	// backup data.
	TestMode bool `json:"test_mode"`
}

// GlobalConfig is the process-wide configuration, populated by LoadConfig.
var GlobalConfig Config

const configFileName = "config.json"

// testBackupRemote is the disposable remote TESTMODE pushes to. It keeps sync
// tests off any real backup repository. Change it when the test repo moves.
const testBackupRemote = "https://github.com/valueforvalue/recipetracker_backup-test.git"

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
		logInfo("config: loaded settings from %s", path)
	case os.IsNotExist(err):
		// Fresh installs default to TESTMODE: sync runs against the disposable
		// test repo and a <backup_path>_test vault until flipped to production.
		GlobalConfig.TestMode = true
		if err := saveConfig(path); err != nil {
			return fmt.Errorf("create default %s: %w", path, err)
		}
		logInfo("config: created default %s", path)
	default:
		return fmt.Errorf("read %s: %w", path, err)
	}

	applyEnvOverrides()
	applyTestMode()
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
			logWarn("config: RECIPE_GIT_SYNC=%q is not a bool (use true/false)", v)
		}
	}
	if v := os.Getenv("RECIPE_GIT_REMOTE"); v != "" {
		GlobalConfig.GitRemoteURL = v
	}
	if v := os.Getenv("RECIPE_AUTO_IMPORT"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			GlobalConfig.AutoImport = b
		} else {
			logWarn("config: RECIPE_AUTO_IMPORT=%q is not a bool (use true/false)", v)
		}
	}
	if v := os.Getenv("RECIPE_TEST_MODE"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			GlobalConfig.TestMode = b
		} else {
			logWarn("config: RECIPE_TEST_MODE=%q is not a bool (use true/false)", v)
		}
	}
}

// applyTestMode redirects git sync to a disposable test repo and a separate
// test vault directory, and forces git_sync on. It runs last so TESTMODE wins
// over both config.json and environment overrides.
func applyTestMode() {
	if !GlobalConfig.TestMode {
		return
	}
	GlobalConfig.GitSync = true
	GlobalConfig.GitRemoteURL = testBackupRemote
	if GlobalConfig.BackupPath == "" {
		home, _ := os.UserHomeDir()
		GlobalConfig.BackupPath = filepath.Join(home, "morris_recipe_vault")
	}
	if !strings.HasSuffix(GlobalConfig.BackupPath, "_test") {
		GlobalConfig.BackupPath += "_test"
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

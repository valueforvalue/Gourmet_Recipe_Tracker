package main

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Set environment variables to override defaults
	os.Setenv("RECIPE_PORT", "9090")
	os.Setenv("RECIPE_DB_FILE", "custom_recipes.db")
	os.Setenv("RECIPE_GIT_SYNC", "false")

	LoadConfig()

	if GlobalConfig.Port != "9090" {
		t.Errorf("Expected Port '9090', got '%s'", GlobalConfig.Port)
	}

	if GlobalConfig.DBFile != "custom_recipes.db" {
		t.Errorf("Expected DBFile 'custom_recipes.db', got '%s'", GlobalConfig.DBFile)
	}

	if GlobalConfig.GitSync != false {
		t.Errorf("Expected GitSync false, got %v", GlobalConfig.GitSync)
	}

	// Clean up environment variables
	os.Unsetenv("RECIPE_PORT")
	os.Unsetenv("RECIPE_DB_FILE")
	os.Unsetenv("RECIPE_GIT_SYNC")
}

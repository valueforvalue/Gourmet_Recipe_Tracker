package main

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestConfigValidate(t *testing.T) {
	cases := []struct {
		name    string
		c       Config
		wantErr bool
	}{
		{"ok", Config{Port: "8080", DBFile: "recipes.db", BackupPath: "/tmp/vault", GitSync: true}, false},
		{"empty port", Config{Port: "", DBFile: "recipes.db"}, true},
		{"non-numeric port", Config{Port: "abc", DBFile: "recipes.db"}, true},
		{"port out of range", Config{Port: "70000", DBFile: "recipes.db"}, true},
		{"empty db file", Config{Port: "8080", DBFile: "   "}, true},
		{"git sync requires backup path", Config{Port: "8080", DBFile: "recipes.db", GitSync: true}, true},
		{"remote with whitespace", Config{Port: "8080", DBFile: "recipes.db", GitRemoteURL: "https://x y"}, true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := c.c.Validate()
			if (err != nil) != c.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, c.wantErr)
			}
		})
	}
}

func TestLoadConfigCreatesDefaultWithGitSyncOff(t *testing.T) {
	os.Unsetenv("RECIPE_GIT_SYNC")
	path := filepath.Join(t.TempDir(), "config.json")

	if err := loadConfig(path); err != nil {
		t.Fatalf("loadConfig: %v", err)
	}
	if GlobalConfig.GitSync {
		t.Error("default git_sync should be false")
	}
	if GlobalConfig.Port != "8080" {
		t.Errorf("default port = %q, want 8080", GlobalConfig.Port)
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("default config file not created: %v", err)
	}
}

func TestLoadConfigEnvOverridesFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	content := `{"port":"8080","db_file":"recipes.db","git_sync":true,"backup_path":"/tmp/vault"}`
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	os.Setenv("RECIPE_PORT", "9090")
	os.Setenv("RECIPE_GIT_SYNC", "false")
	defer os.Unsetenv("RECIPE_PORT")
	defer os.Unsetenv("RECIPE_GIT_SYNC")

	if err := loadConfig(path); err != nil {
		t.Fatalf("loadConfig: %v", err)
	}
	if GlobalConfig.Port != "9090" {
		t.Errorf("env port = %q, want 9090 (env must override file)", GlobalConfig.Port)
	}
	if GlobalConfig.GitSync {
		t.Error("env RECIPE_GIT_SYNC=false did not override file git_sync=true")
	}
}

func TestLoadConfigStrictGitSyncParsing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(`{"port":"8080","db_file":"recipes.db"}`), 0600); err != nil {
		t.Fatal(err)
	}

	os.Setenv("RECIPE_GIT_SYNC", "0")
	defer os.Unsetenv("RECIPE_GIT_SYNC")

	if err := loadConfig(path); err != nil {
		t.Fatalf("loadConfig: %v", err)
	}
	if GlobalConfig.GitSync {
		t.Error(`RECIPE_GIT_SYNC=0 must parse as false (strict ParseBool)`)
	}
}

func TestLoadConfigRejectsMalformedFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte("{not json"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := loadConfig(path); err == nil {
		t.Fatal("loadConfig should fail on malformed config.json")
	}
}

func TestSaveConfigPermissions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	GlobalConfig = Config{Port: "8080", DBFile: "recipes.db"}
	if err := saveConfig(path); err != nil {
		t.Fatalf("saveConfig: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	// Windows does not store POSIX permission bits: os.WriteFile's 0600 is
	// honored on Unix but NTFS reports a default mask. Assert only where the
	// filesystem actually enforces the mode.
	if runtime.GOOS != "windows" && info.Mode().Perm() != 0600 {
		t.Errorf("config perms = %v, want 0600 (owner-only)", info.Mode().Perm())
	}
}

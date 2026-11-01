package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// snapshotFileName is the JSON snapshot written to the vault after every
// mutation. It is the primary import path for crash recovery.
const snapshotFileName = "recipes.json"

// Snapshot is the on-disk shape of recipes.json: the full active library plus
// provenance (app version, schema version, export time) so a backup can be
// traced back to the build that produced it.
type Snapshot struct {
	AppVersion    string   `json:"app_version"`
	SchemaVersion int      `json:"schema_version"`
	ExportedAt    string   `json:"exported_at"`
	Recipes       []Recipe `json:"recipes"`
}

// snapshotMu serializes snapshot writes so concurrent request handlers cannot
// interleave a temp-file write with its rename.
var snapshotMu sync.Mutex

// writeSnapshot writes the full active recipe list to dir/recipes.json,
// atomically (temp file + rename).
func writeSnapshot(dir string, db *sql.DB) error {
	recipes, err := GetAllRecipes(db)
	if err != nil {
		return err
	}
	snap := Snapshot{
		AppVersion:    Version,
		SchemaVersion: SchemaVersion,
		ExportedAt:    time.Now().Format(time.RFC3339),
		Recipes:       recipes,
	}
	data, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	tmp := filepath.Join(dir, snapshotFileName+".tmp")
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, filepath.Join(dir, snapshotFileName))
}

// SyncSnapshot refreshes the JSON snapshot after a DB mutation.
func SyncSnapshot(db *sql.DB) error {
	snapshotMu.Lock()
	defer snapshotMu.Unlock()
	return writeSnapshot(vaultDir(), db)
}

// enqueueCommit asks the sync worker to commit+push the current vault state.
// Used by mutations that change the snapshot without writing a .txt sidecar
// (e.g. purge).
func enqueueCommit() {
	if !GlobalConfig.GitSync {
		return
	}
	select {
	case syncQueue <- SyncTask{Op: "commit", Title: "snapshot"}:
	default:
		logError("sync queue full, snapshot commit skipped")
	}
}

// ImportFromJSON rebuilds the database from a recipes.json snapshot, upserting
// every recipe. It returns the number of recipes imported.
func ImportFromJSON(path string, db *sql.DB) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}

	var recipes []Recipe
	if trimmed := bytes.TrimSpace(data); len(trimmed) > 0 && trimmed[0] == '[' {
		// Legacy snapshot: a bare recipe array with no provenance fields.
		if err := json.Unmarshal(data, &recipes); err != nil {
			return 0, fmt.Errorf("parse %s: %w", path, err)
		}
	} else {
		var snap Snapshot
		if err := json.Unmarshal(data, &snap); err != nil {
			return 0, fmt.Errorf("parse %s: %w", path, err)
		}
		recipes = snap.Recipes
	}

	for _, r := range recipes {
		if err := saveRecipeDB(db, r); err != nil {
			return 0, fmt.Errorf("import %q: %w", r.Title, err)
		}
	}
	return len(recipes), nil
}

// isRecipesEmpty reports whether the recipes table has no rows at all.
func isRecipesEmpty(db *sql.DB) (bool, error) {
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM recipes").Scan(&count); err != nil {
		return false, err
	}
	return count == 0, nil
}

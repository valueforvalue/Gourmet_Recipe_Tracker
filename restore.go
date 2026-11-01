// File: restore.go
package main

import (
	"database/sql"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// RestoreFromVault rebuilds the database from the vault's .txt files.  It
// uses manifest.json when present and also picks up any legacy .txt files the
// manifest does not list, so a partial or pre-manifest vault still restores
// completely.  Returns the number of recipes restored.
func RestoreFromVault(dir string, db *sql.DB) (int, error) {
	files, err := vaultFiles(dir)
	if err != nil {
		return 0, err
	}

	restored := 0
	for _, f := range files {
		r, err := ParseRecipeFile(filepath.Join(dir, f))
		if err != nil {
			logWarn("restore: skip %s: %v", f, err)
			continue
		}
		if r.Title == "" {
			logWarn("restore: skip %s: missing title", f)
			continue
		}
		if err := saveRecipeDB(db, r); err != nil {
			logError("restore: failed %s: %v", f, err)
			continue
		}
		restored++
	}
	return restored, nil
}

// vaultFiles returns the union of manifest-listed files and every *.txt in the
// directory, deduplicated and sorted.
func vaultFiles(dir string) ([]string, error) {
	seen := map[string]bool{}
	m := loadManifest(dir)
	for _, e := range m.Recipes {
		if e.File != "" {
			seen[e.File] = true
		}
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.EqualFold(filepath.Ext(name), ".txt") {
			seen[name] = true
		}
	}

	files := make([]string, 0, len(seen))
	for f := range seen {
		files = append(files, f)
	}
	sort.Strings(files)
	return files, nil
}

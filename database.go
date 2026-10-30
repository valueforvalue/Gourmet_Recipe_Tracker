// File: database.go
package main

import (
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

// --- CONFIGURATION ---

// InitConfig loads settings from GlobalConfig and handles Git syncing logic.
func InitConfig() {
	// Automatic Git Initialization & Linking
	if GlobalConfig.GitSync {
		if _, err := os.Stat(GlobalConfig.BackupPath); os.IsNotExist(err) {
			os.MkdirAll(GlobalConfig.BackupPath, 0755)
		}

		// Check if it's already a git repo
		dotGit := filepath.Join(GlobalConfig.BackupPath, ".git")
		if _, err := os.Stat(dotGit); os.IsNotExist(err) {
			fmt.Printf(" [Git]: Initializing new repository at %s\n", GlobalConfig.BackupPath)
			exec.Command("git", "-C", GlobalConfig.BackupPath, "init").Run()
		}

		// Automatically link remote if provided and not already linked
		if GlobalConfig.GitRemoteURL != "" {
			// Check if origin already exists
			out, _ := exec.Command("git", "-C", GlobalConfig.BackupPath, "remote").Output()
			if !strings.Contains(string(out), "origin") {
				fmt.Printf(" [Git]: Linking to remote %s\n", GlobalConfig.GitRemoteURL)
				exec.Command("git", "-C", GlobalConfig.BackupPath, "remote", "add", "origin", GlobalConfig.GitRemoteURL).Run()
				// Initial pull to sync with existing remote data if any
				exec.Command("git", "-C", GlobalConfig.BackupPath, "pull", "origin", "master").Run()
			}
		}
	}
}

// InitDB sets up the SQLite connection and schema.
func InitDB(filepath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", filepath)
	if err != nil {
		return nil, err
	}

	// SQLite allows a single writer at a time.  With the Flutter app, the web
	// UI, and possibly multiple devices all hitting the server concurrently,
	// the default settings cause "database is locked" errors.  Serialize writes
	// through one connection and enable WAL so readers never block the writer.
	db.SetMaxOpenConns(1)
	for _, stmt := range []string{
		"PRAGMA journal_mode = WAL",
		"PRAGMA busy_timeout = 5000",
		"PRAGMA synchronous = NORMAL",
	} {
		if _, err := db.Exec(stmt); err != nil {
			db.Close()
			return nil, fmt.Errorf("apply %q: %w", stmt, err)
		}
	}

	query := `
	CREATE TABLE IF NOT EXISTS recipes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT UNIQUE,
		ingredients TEXT,
		instructions TEXT,
		tags TEXT,
		notes TEXT,
		is_deleted BOOLEAN DEFAULT 0
	);`

	_, err = db.Exec(query)
	if err != nil {
		return nil, err
	}

	// Normalize existing titles to fix legacy data issues
	rows, err := db.Query("SELECT title FROM recipes")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var oldTitle string
			if err := rows.Scan(&oldTitle); err == nil {
				newTitle := strings.Join(strings.Fields(oldTitle), " ")
				if oldTitle != newTitle {
					db.Exec("UPDATE recipes SET title = ? WHERE title = ?", newTitle, oldTitle)
				}
			}
		}
	}

	return db, nil
}

// SaveRecipe persists a recipe and mirrors it to the vault.
func SaveRecipe(db *sql.DB, r Recipe) error {
	if err := saveRecipeDB(db, r); err != nil {
		fmt.Printf(" [DB Error]: %v\n", err)
		return err
	}
	if err := SyncSidecar(r); err != nil {
		fmt.Printf(" [Sync Error]: %v\n", err)
	}
	return nil
}

// saveRecipeDB performs the database UPSERT only (no vault sync).  It is used
// by both the save handler and the restore tool.
func saveRecipeDB(db *sql.DB, r Recipe) error {
	r.Title = strings.Join(strings.Fields(r.Title), " ")
	ingStr := strings.Join(r.Ingredients, "|")
	insStr := strings.Join(r.Instructions, "|")

	// Clean tags before saving
	var cleanTags []string
	for _, t := range r.Tags {
		trimmed := strings.TrimSpace(t)
		if trimmed != "" {
			cleanTags = append(cleanTags, trimmed)
		}
	}
	tagStr := strings.Join(cleanTags, ",")

	query := `
	INSERT INTO recipes (title, ingredients, instructions, tags, notes, is_deleted)
	VALUES (?, ?, ?, ?, ?, 0)
	ON CONFLICT(title) DO UPDATE SET
		ingredients=excluded.ingredients,
		instructions=excluded.instructions,
		tags=excluded.tags,
		notes=excluded.notes,
		is_deleted=0;`

	_, err := db.Exec(query, r.Title, ingStr, insStr, tagStr, r.Notes)
	return err
}

// safeSplit prevents empty strings from becoming blank array items
func safeSplit(data string, separator string) []string {
	if data == "" {
		return []string{}
	}
	var result []string
	for _, item := range strings.Split(data, separator) {
		trimmed := strings.TrimSpace(item)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// sanitizeFilename converts a recipe title into a filesystem-safe base name.
// It strips path separators and directory components (blocking path traversal)
// and replaces Windows-reserved characters, so it is safe for both vault
// sidecar files and generated PDFs.
func sanitizeFilename(name string) string {
	safe := strings.NewReplacer(
		"/", "-", "\\", "-", ":", "-", "*", "-", "?", "-",
		"\"", "-", "<", "-", ">", "-", "|", "-",
	).Replace(name)
	safe = filepath.Base(safe)
	safe = strings.TrimSpace(safe)
	safe = strings.Trim(safe, ".")
	if safe == "" || safe == "." {
		safe = "unnamed"
	}
	return safe
}

// GetAllRecipes retrieves active recipes for the web view.
func GetAllRecipes(db *sql.DB) ([]Recipe, error) {
	query := `SELECT title, ingredients, instructions, tags, notes FROM recipes WHERE is_deleted = 0 ORDER BY title ASC`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recipes []Recipe
	for rows.Next() {
		var r Recipe
		var ingStr, insStr, tagStr string
		err = rows.Scan(&r.Title, &ingStr, &insStr, &tagStr, &r.Notes)
		if err == nil {
			r.Ingredients = safeSplit(ingStr, "|")
			r.Instructions = safeSplit(insStr, "|")
			r.Tags = safeSplit(tagStr, ",")
			recipes = append(recipes, r)
		}
	}
	return recipes, nil
}

// GetRecipeByTitle retrieves a specific recipe for PDF generation.
func GetRecipeByTitle(db *sql.DB, title string) (Recipe, error) {
	var r Recipe
	var ingStr, insStr, tagStr string
	query := `SELECT title, ingredients, instructions, tags, notes FROM recipes WHERE title = ? AND is_deleted = 0`

	err := db.QueryRow(query, title).Scan(&r.Title, &ingStr, &insStr, &tagStr, &r.Notes)
	if err != nil {
		return r, err
	}

	r.Ingredients = safeSplit(ingStr, "|")
	r.Instructions = safeSplit(insStr, "|")
	r.Tags = safeSplit(tagStr, ",")
	return r, nil
}

// DeleteRecipe flags a recipe as deleted (Soft Delete) and removes its vault
// file so the vault stays in sync with the active library.
func DeleteRecipe(db *sql.DB, title string) error {
	if _, err := db.Exec("UPDATE recipes SET is_deleted = 1 WHERE title = ?", title); err != nil {
		return err
	}
	SyncVaultDelete(title)
	return nil
}

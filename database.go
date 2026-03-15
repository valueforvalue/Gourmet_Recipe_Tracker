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
// Set enableGitSync to true to auto-push to GitHub.
// Ensure backupRepoPath is a folder initialized with 'git init' and a remote.
var enableGitSync = true
var backupRepoPath = `C:\Development\morris_recipe_backups`

// InitDB sets up the SQLite connection and schema.
func InitDB(filepath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", filepath)
	if err != nil {
		return nil, err
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
	return db, err
}

// SaveRecipe handles the database UPSERT and triggers the file sync.
func SaveRecipe(db *sql.DB, r Recipe) error {
	ingStr := strings.Join(r.Ingredients, "|")
	insStr := strings.Join(r.Instructions, "|")

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
	if err != nil {
		fmt.Printf(" [DB Error]: %v\n", err)
		return err
	}

	return SyncSidecar(r)
}

// SyncSidecar handles the "Daily Bread" text backup and GitHub sync.
func SyncSidecar(r Recipe) error {
	// Create local program backup folder if missing
	_ = os.Mkdir("backups", 0755)

	// Sanitize title for filename
	safeTitle := strings.ReplaceAll(r.Title, "/", "-")
	safeTitle = strings.ReplaceAll(safeTitle, "\\", "-")
	fileName := safeTitle + ".txt"

	// Prepare text content
	content := fmt.Sprintf("RECIPE: %s\nTAGS: %s\n\nINGREDIENTS\n- %s\n\nINSTRUCTIONS\n%s\n\nNOTES: %s",
		r.Title,
		strings.Join(r.Tags, ", "),
		strings.Join(r.Ingredients, "\n- "),
		strings.Join(r.Instructions, "\n"),
		r.Notes)

	// Save to the local program backups folder
	os.WriteFile(filepath.Join("backups", fileName), []byte(content), 0644)

	if enableGitSync {
		// Run Git operations in a background thread
		go func(title string, fileContent string) {
			// 1. Teleport file to the Vault folder
			destPath := filepath.Join(backupRepoPath, fileName)
			err := os.WriteFile(destPath, []byte(fileContent), 0644)
			if err != nil {
				fmt.Printf(" [Sync Error]: Could not write to backup path: %v\n", err)
				return
			}

			// 2. Execute Git commands using the -C flag (Target Directory)
			// Add
			exec.Command("git", "-C", backupRepoPath, "add", ".").Run()

			// Commit
			commitMsg := fmt.Sprintf("Update %s", title)
			exec.Command("git", "-C", backupRepoPath, "commit", "-m", commitMsg).Run()

			// Push to origin master
			out, err := exec.Command("git", "-C", backupRepoPath, "push", "origin", "master").CombinedOutput()

			if err != nil {
				fmt.Printf(" [Git Error]: %s\n", string(out))
			} else {
				fmt.Printf(" [Git Success]: %s successfully vaulted to GitHub.\n", title)
			}
		}(r.Title, content)
	}

	return nil
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
			r.Ingredients = strings.Split(ingStr, "|")
			r.Instructions = strings.Split(insStr, "|")
			r.Tags = strings.Split(tagStr, ",")
			recipes = append(recipes, r)
		}
	}
	return recipes, nil
}

// GetRecipeByTitle retrieves a specific recipe for PDF generation.
func GetRecipeByTitle(db *sql.DB, title string) (Recipe, error) {
	var r Recipe
	var ingStr, insStr, tagStr string
	query := `SELECT title, ingredients, instructions, tags, notes FROM recipes WHERE title = ?`

	err := db.QueryRow(query, title).Scan(&r.Title, &ingStr, &insStr, &tagStr, &r.Notes)
	if err != nil {
		return r, err
	}

	r.Ingredients = strings.Split(ingStr, "|")
	r.Instructions = strings.Split(insStr, "|")
	r.Tags = strings.Split(tagStr, ",")
	return r, nil
}

// DeleteRecipe flags a recipe as deleted (Soft Delete).
func DeleteRecipe(db *sql.DB, title string) error {
	_, err := db.Exec("UPDATE recipes SET is_deleted = 1 WHERE title = ?", title)
	return err
}

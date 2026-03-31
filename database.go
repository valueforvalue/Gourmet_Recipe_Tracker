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

type SyncTask struct {
	Title    string
	FileName string
	Content  string
}

var syncQueue = make(chan SyncTask, 100)

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

// StartSyncWorker runs in the background and processes Git syncs one at a time.
func StartSyncWorker() {
	for task := range syncQueue {
		// 1. Write file to the Vault folder
		destPath := filepath.Join(GlobalConfig.BackupPath, task.FileName)
		err := os.WriteFile(destPath, []byte(task.Content), 0644)
		if err != nil {
			fmt.Printf(" [Sync Error]: Could not write to backup path: %v\n", err)
			continue
		}

		// 2. Execute Git commands sequentially
		exec.Command("git", "-C", GlobalConfig.BackupPath, "add", ".").Run()

		commitMsg := fmt.Sprintf("Update %s", task.Title)
		exec.Command("git", "-C", GlobalConfig.BackupPath, "commit", "-m", commitMsg).Run()

		out, err := exec.Command("git", "-C", GlobalConfig.BackupPath, "push", "origin", "master").CombinedOutput()

		if err != nil {
			fmt.Printf(" [Git Error]: %s\n", string(out))
		} else {
			fmt.Printf(" [Git Success]: %s successfully vaulted to GitHub.\n", task.Title)
		}
	}
}

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

// SaveRecipe handles the database UPSERT and triggers the file sync.
func SaveRecipe(db *sql.DB, r Recipe) error {
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

	if GlobalConfig.GitSync {
		syncQueue <- SyncTask{
			Title:    r.Title,
			FileName: fileName,
			Content:  content,
		}
	}

	return nil
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

// DeleteRecipe flags a recipe as deleted (Soft Delete).
func DeleteRecipe(db *sql.DB, title string) error {
	_, err := db.Exec("UPDATE recipes SET is_deleted = 1 WHERE title = ?", title)
	return err
}

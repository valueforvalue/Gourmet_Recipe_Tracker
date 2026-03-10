package main

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	_ "modernc.org/sqlite"
)

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

func SaveRecipe(db *sql.DB, r Recipe) error {
	ingStr := strings.Join(r.Ingredients, "|")
	insStr := strings.Join(r.Instructions, "|")
	tagStr := strings.Join(r.Tags, ",")

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
		return err
	}
	return SyncSidecar(r)
}

func GetRecipeByTitle(db *sql.DB, title string) (Recipe, error) {
	var r Recipe
	var ingStr, insStr, tagStr string
	query := `SELECT title, ingredients, instructions, tags, notes FROM recipes WHERE title = ? AND is_deleted = 0`
	err := db.QueryRow(query, title).Scan(&r.Title, &ingStr, &insStr, &tagStr, &r.Notes)
	if err != nil {
		return r, err
	}
	r.Ingredients = strings.Split(ingStr, "|")
	r.Instructions = strings.Split(insStr, "|")
	r.Tags = strings.Split(tagStr, ",")
	return r, nil
}

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
		if err := rows.Scan(&r.Title, &ingStr, &insStr, &tagStr, &r.Notes); err == nil {
			r.Ingredients = strings.Split(ingStr, "|")
			r.Instructions = strings.Split(insStr, "|")
			r.Tags = strings.Split(tagStr, ",")
			recipes = append(recipes, r)
		}
	}
	return recipes, nil
}

func DeleteRecipe(db *sql.DB, title string) error {
	_, err := db.Exec("UPDATE recipes SET is_deleted = 1 WHERE title = ?", title)
	return err
}

func SyncSidecar(r Recipe) error {
	_ = os.Mkdir("backups", 0755)
	safeTitle := strings.ReplaceAll(r.Title, "/", "-")
	path := fmt.Sprintf("backups/%s.txt", safeTitle)
	content := fmt.Sprintf("RECIPE: %s\nTAGS: %s\n\nINGREDIENTS\n- %s\n\nINSTRUCTIONS\n%s\n\nNOTES: %s",
		r.Title, strings.Join(r.Tags, ", "), strings.Join(r.Ingredients, "\n- "), strings.Join(r.Instructions, "\n"), r.Notes)
	return os.WriteFile(path, []byte(content), 0644)
}

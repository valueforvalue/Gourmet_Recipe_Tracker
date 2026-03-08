package main

import (
	"database/sql"
	"strings"

	_ "modernc.org/sqlite"
)

func InitDB(filepath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", filepath)
	if err != nil {
		return nil, err
	}

	// We make source_file the UNIQUE key now
	query := `
	CREATE TABLE IF NOT EXISTS recipes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		source_file TEXT UNIQUE, 
		title TEXT,
		tags TEXT,
		ingredients TEXT,
		instructions TEXT,
		notes TEXT
	);`

	_, err = db.Exec(query)
	return db, err
}

func SaveRecipe(db *sql.DB, r Recipe) error {
	ingreds := strings.Join(r.Ingredients, "|")
	instruc := strings.Join(r.Instructions, "|")
	tags := strings.Join(r.Tags, ",")

	// This logic says: "If the filename is the same, just update the details."
	query := `
	INSERT INTO recipes (source_file, title, tags, ingredients, instructions, notes) 
	VALUES (?, ?, ?, ?, ?, ?)
	ON CONFLICT(source_file) DO UPDATE SET
		title=excluded.title,
		tags=excluded.tags,
		ingredients=excluded.ingredients,
		instructions=excluded.instructions,
		notes=excluded.notes;`

	_, err := db.Exec(query, r.SourceFile, r.Title, tags, ingreds, instruc, r.Notes)
	return err
}

func GetRecipeCount(db *sql.DB) int {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM recipes").Scan(&count)
	if err != nil {
		return 0
	}
	return count
}

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

	query := `
	CREATE TABLE IF NOT EXISTS recipes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT UNIQUE,
		ingredients TEXT,
		instructions TEXT,
		tags TEXT,
		notes TEXT
	);`

	_, err = db.Exec(query)
	return db, err
}

func SaveRecipe(db *sql.DB, r Recipe) error {
	ingStr := strings.Join(r.Ingredients, "|")
	insStr := strings.Join(r.Instructions, "|")
	tagStr := strings.Join(r.Tags, ",")

	query := `
	INSERT INTO recipes (title, ingredients, instructions, tags, notes)
	VALUES (?, ?, ?, ?, ?)
	ON CONFLICT(title) DO UPDATE SET
		ingredients=excluded.ingredients,
		instructions=excluded.instructions,
		tags=excluded.tags,
		notes=excluded.notes;`

	_, err := db.Exec(query, r.Title, ingStr, insStr, tagStr, r.Notes)
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

func GetAllRecipes(db *sql.DB) ([]Recipe, error) {
	rows, err := db.Query("SELECT title, ingredients, instructions, tags, notes FROM recipes ORDER BY title ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recipes []Recipe
	for rows.Next() {
		var r Recipe
		var ingStr, insStr, tagStr string
		err := rows.Scan(&r.Title, &ingStr, &insStr, &tagStr, &r.Notes)
		if err == nil {
			r.Ingredients = strings.Split(ingStr, "|")
			r.Instructions = strings.Split(insStr, "|")
			r.Tags = strings.Split(tagStr, ",")
			recipes = append(recipes, r)
			recipes[len(recipes)-1].Title = strings.TrimSpace(r.Title)
		}
	}
	return recipes, nil
}

// Added for the Selection Config logic
func GetRecipeByTitle(db *sql.DB, title string) (Recipe, error) {
	var r Recipe
	var ingStr, insStr, tagStr string
	query := "SELECT title, ingredients, instructions, tags, notes FROM recipes WHERE title = ?"
	err := db.QueryRow(query, title).Scan(&r.Title, &ingStr, &insStr, &tagStr, &r.Notes)
	if err != nil {
		return r, err
	}
	r.Ingredients = strings.Split(ingStr, "|")
	r.Instructions = strings.Split(insStr, "|")
	r.Tags = strings.Split(tagStr, ",")
	return r, nil
}

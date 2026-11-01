package main

import (
	"database/sql"
	"path/filepath"
	"testing"
)

// TestInitDBNormalizesLegacyTitlesWithoutDeadlock seeds a legacy double-spaced
// title and runs InitDB against it. The title-normalization path previously
// held the single pooled connection open (via *Rows) while issuing an UPDATE
// inside the same loop, which deadlocked under SetMaxOpenConns(1).
func TestInitDBNormalizesLegacyTitlesWithoutDeadlock(t *testing.T) {
	path := filepath.Join(t.TempDir(), "legacy.db")

	seed, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := seed.Exec(`CREATE TABLE recipes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT UNIQUE,
		ingredients TEXT,
		instructions TEXT,
		tags TEXT,
		notes TEXT,
		is_deleted BOOLEAN DEFAULT 0
	)`); err != nil {
		t.Fatal(err)
	}
	if _, err := seed.Exec(`INSERT INTO recipes (title, ingredients, instructions, tags, notes, is_deleted)
		VALUES ('My  Recipe', '', '', '', '', 0)`); err != nil {
		t.Fatal(err)
	}
	seed.Close()

	db, err := InitDB(path)
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	defer db.Close()

	var title string
	if err := db.QueryRow("SELECT title FROM recipes").Scan(&title); err != nil {
		t.Fatal(err)
	}
	if title != "My Recipe" {
		t.Errorf("title = %q, want %q (legacy title not normalized)", title, "My Recipe")
	}
}

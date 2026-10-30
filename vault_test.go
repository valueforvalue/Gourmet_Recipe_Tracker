package main

import (
	"path/filepath"
	"testing"
)

func TestVaultWriteManifestRestoreRoundTrip(t *testing.T) {
	dir := t.TempDir()
	dbFile := filepath.Join(dir, "restore.db")
	db, err := InitDB(dbFile)
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	defer db.Close()

	recipe := Recipe{
		Title:        "Round Trip Cake",
		Tags:         []string{"Dessert", "Easy"},
		Ingredients:  []string{"1 cup Sugar", "2 cups Flour"},
		Instructions: []string{"Mix.", "Bake."},
		Notes:        "Warm.",
	}

	fileName, err := writeVaultEntry(dir, recipe)
	if err != nil {
		t.Fatalf("writeVaultEntry: %v", err)
	}
	if fileName != "Round Trip Cake.txt" {
		t.Errorf("unexpected filename %q", fileName)
	}

	m := loadManifest(dir)
	if len(m.Recipes) != 1 || m.Recipes[0].Title != recipe.Title {
		t.Errorf("manifest mismatch: %+v", m.Recipes)
	}

	n, err := RestoreFromVault(dir, db)
	if err != nil {
		t.Fatalf("RestoreFromVault: %v", err)
	}
	if n != 1 {
		t.Errorf("expected 1 restored, got %d", n)
	}

	got, err := GetRecipeByTitle(db, recipe.Title)
	if err != nil || got.Title != recipe.Title || len(got.Ingredients) != 2 {
		t.Errorf("restore mismatch: %+v err=%v", got, err)
	}

	if _, err := removeVaultEntry(dir, recipe.Title); err != nil {
		t.Fatalf("removeVaultEntry: %v", err)
	}
	if len(loadManifest(dir).Recipes) != 0 {
		t.Errorf("expected empty manifest after remove, got %+v", loadManifest(dir).Recipes)
	}
}

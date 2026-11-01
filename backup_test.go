package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteSnapshotRoundTrip(t *testing.T) {
	GlobalConfig.GitSync = false
	db := newSoftDeleteTestDB(t)

	for _, title := range []string{"Cake", "Pie"} {
		if err := saveRecipeDB(db, Recipe{Title: title}); err != nil {
			t.Fatal(err)
		}
	}

	dir := t.TempDir()
	if err := writeSnapshot(dir, db); err != nil {
		t.Fatalf("writeSnapshot: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, snapshotFileName))
	if err != nil {
		t.Fatal(err)
	}
	var got []Recipe
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Errorf("snapshot has %d recipes, want 2", len(got))
	}
}

func TestImportFromJSON(t *testing.T) {
	GlobalConfig.GitSync = false
	db := newSoftDeleteTestDB(t)

	snap := []Recipe{
		{Title: "Imported A", Ingredients: []string{"x"}},
		{Title: "Imported B"},
	}
	data, err := json.Marshal(snap)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "recipes.json")
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}

	n, err := ImportFromJSON(path, db)
	if err != nil {
		t.Fatalf("ImportFromJSON: %v", err)
	}
	if n != 2 {
		t.Errorf("imported %d, want 2", n)
	}
	all, err := GetAllRecipes(db)
	if err != nil || len(all) != 2 {
		t.Errorf("GetAllRecipes after import = %d (err=%v), want 2", len(all), err)
	}
}

func TestImportFromJSONMalformed(t *testing.T) {
	db := newSoftDeleteTestDB(t)
	path := filepath.Join(t.TempDir(), "bad.json")
	if err := os.WriteFile(path, []byte("{not json"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := ImportFromJSON(path, db); err == nil {
		t.Error("ImportFromJSON should fail on malformed JSON")
	}
}

func TestIsRecipesEmpty(t *testing.T) {
	db := newSoftDeleteTestDB(t)

	empty, err := isRecipesEmpty(db)
	if err != nil || !empty {
		t.Fatalf("expected empty, got %v err=%v", empty, err)
	}

	if err := saveRecipeDB(db, Recipe{Title: "X"}); err != nil {
		t.Fatal(err)
	}
	empty, err = isRecipesEmpty(db)
	if err != nil || empty {
		t.Errorf("expected non-empty after save, got %v err=%v", empty, err)
	}
}

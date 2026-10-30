package main

import (
	"os"
	"testing"
)

func TestDatabaseOperations(t *testing.T) {
	dbFile := "test_recipes.db"
	db, err := InitDB(dbFile)
	if err != nil {
		t.Fatalf("Failed to init DB: %v", err)
	}
	defer db.Close()
	defer os.Remove(dbFile)

	// Test Save
	recipe := Recipe{
		Title:        "Test Recipe",
		Tags:         []string{"Tag1", "Tag2"},
		Ingredients:  []string{"Ing1", "Ing2"},
		Instructions: []string{"Inst1", "Inst2"},
		Notes:        "Some notes",
	}

	// Disable git sync for testing
	GlobalConfig.GitSync = false

	err = SaveRecipe(db, recipe)
	if err != nil {
		t.Fatalf("Failed to save recipe: %v", err)
	}

	// Test Get
	fetched, err := GetRecipeByTitle(db, "Test Recipe")
	if err != nil {
		t.Fatalf("Failed to fetch recipe: %v", err)
	}

	if fetched.Title != recipe.Title {
		t.Errorf("Expected title %s, got %s", recipe.Title, fetched.Title)
	}

	// Test Delete
	err = DeleteRecipe(db, "Test Recipe")
	if err != nil {
		t.Fatalf("Failed to delete recipe: %v", err)
	}

	_, err = GetRecipeByTitle(db, "Test Recipe")
	if err == nil {
		t.Error("Expected error fetching deleted recipe, got nil")
	}

	// Test GetAll
	recipes, err := GetAllRecipes(db)
	if err != nil {
		t.Fatalf("Failed to get all recipes: %v", err)
	}
	if len(recipes) != 0 {
		t.Errorf("Expected 0 recipes, got %d", len(recipes))
	}
}

func TestSanitizeFilename(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"Simple Recipe", "Simple Recipe"},
		// Path separators are replaced with dashes; the result stays within the
		// target directory (no traversal) because no slashes remain.
		{"../../../etc/passwd", "..-..-..-etc-passwd"},
		{"recipe/with/slashes", "recipe-with-slashes"},
		{"recipe\\with\\backslashes", "recipe-with-backslashes"},
		{"  spaces  ", "spaces"},
		{"", "unnamed"},
		{".", "unnamed"},
	}

	for _, c := range cases {
		got := sanitizeFilename(c.input)
		if got != c.expected {
			t.Errorf("sanitizeFilename(%q) = %q, want %q", c.input, got, c.expected)
		}
	}
}

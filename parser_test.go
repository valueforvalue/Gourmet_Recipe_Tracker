package main

import (
	"os"
	"testing"
)

func TestParseRecipeFile(t *testing.T) {
	// Create a dummy recipe file
	content := `RECIPE: Test Cake
TAGS: Dessert, Easy

INGREDIENTS
- 1 cup Sugar
- 2 cups Flour

INSTRUCTIONS
1. Mix ingredients.
2. Bake at 350F.

NOTES: Best served warm.`

	tmpFile := "test_recipe.txt"
	os.WriteFile(tmpFile, []byte(content), 0644)
	defer os.Remove(tmpFile)

	recipe, err := ParseRecipeFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to parse recipe: %v", err)
	}

	if recipe.Title != "Test Cake" {
		t.Errorf("Expected Title 'Test Cake', got '%s'", recipe.Title)
	}

	if len(recipe.Tags) != 2 || recipe.Tags[0] != "Dessert" || recipe.Tags[1] != "Easy" {
		t.Errorf("Expected Tags [Dessert, Easy], got %v", recipe.Tags)
	}

	if len(recipe.Ingredients) != 2 || recipe.Ingredients[0] != "1 cup Sugar" {
		t.Errorf("Expected first ingredient '1 cup Sugar', got '%s'", recipe.Ingredients[0])
	}

	if len(recipe.Instructions) != 2 || recipe.Instructions[0] != "1. Mix ingredients." {
		t.Errorf("Expected first instruction '1. Mix ingredients.', got '%s'", recipe.Instructions[0])
	}

	if recipe.Notes != "Best served warm." {
		t.Errorf("Expected Notes 'Best served warm.', got '%s'", recipe.Notes)
	}
}

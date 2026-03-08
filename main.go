package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	// 1. Configuration
	inputDir := "recipes_to_import"
	outputDir := "Printables"
	dbFile := "recipes.db"

	// 2. Ensure folders exist
	ensureDirExists(inputDir)
	ensureDirExists(outputDir)

	// 3. Connect to the Database
	db, err := InitDB(dbFile)
	if err != nil {
		log.Fatalf("Critical: Could not initialize database: %v", err)
	}
	defer db.Close()

	// 4. Read the input folder
	files, err := os.ReadDir(inputDir)
	if err != nil {
		log.Fatalf("Critical: Could not read folder %s: %v", inputDir, err)
	}

	fmt.Println("========================================")
	fmt.Println("   RECIPE TRACKER: ACTIVE SYNC          ")
	fmt.Println("========================================")

	successCount := 0
	errorCount := 0

	for _, file := range files {
		// Only process .txt files
		if !file.IsDir() && strings.HasSuffix(strings.ToLower(file.Name()), ".txt") {

			// --- THE IGNORE RULE ---
			// If the filename is Template.txt (case-insensitive), we skip it
			if strings.EqualFold(file.Name(), "Template.txt") {
				continue
			}

			fullPath := filepath.Join(inputDir, file.Name())

			recipe, err := ParseFile(fullPath)
			if err != nil {
				fmt.Printf("[FAIL] Could not read %s: %v\n", file.Name(), err)
				errorCount++
				continue
			}

			err = SaveRecipe(db, recipe)
			if err != nil {
				fmt.Printf("[FAIL] Database error for %s: %v\n", recipe.Title, err)
				errorCount++
				continue
			}

			err = ExportToPDF(recipe)
			if err != nil {
				fmt.Printf("[FAIL] PDF error for %s: %v\n", recipe.Title, err)
				errorCount++
				continue
			}

			fmt.Printf("[OK] Synced: %s\n", recipe.Title)
			successCount++
		}
	}

	fmt.Println("========================================")
	fmt.Printf("Sync Complete: %d Updated, %d Failed.\n", successCount, errorCount)
	fmt.Println("========================================")
}

func ensureDirExists(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.Mkdir(path, 0755)
		if err != nil {
			log.Fatalf("Could not create folder %s: %v", path, err)
		}
	}
}

package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func main() {
	inputDir := "recipes_to_import"
	outputDir := "Printables"
	dbFile := "recipes.db"

	// These functions must be defined below or in another file
	ensureDirExists(inputDir)
	ensureDirExists(outputDir)

	db, err := InitDB(dbFile)
	if err != nil {
		log.Fatalf("Critical: %v", err)
	}
	defer db.Close()

	reader := bufio.NewReader(os.Stdin)

	for {
		clearScreen()
		count := GetRecipeCount(db)

		fmt.Println("########################################################")
		fmt.Println("#             GOURMET RECIPE TRACKER v2.1              #")
		fmt.Printf("#          Library Size: %d Recipes Saved               #\n", count)
		fmt.Println("########################################################")
		fmt.Println("#                                                      #")
		fmt.Println("#  [1] SYNC: Standard (Full Page)                      #")
		fmt.Println("#  [2] SYNC: Booklet (Half-Page / Book Size)           #")
		fmt.Println("#  [3] OPEN: View 'Printables' Folder                  #")
		fmt.Println("#  [4] CLEAN: Re-create Template.txt                   #")
		fmt.Println("#                                                      #")
		fmt.Println("#  [Q] QUIT                                            #")
		fmt.Println("########################################################")
		fmt.Print("\n > Selection: ")

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToUpper(input))

		switch input {
		case "1":
			runFullSync(db, inputDir, false)
			pause(reader)
		case "2":
			runFullSync(db, inputDir, true)
			pause(reader)
		case "3":
			openFolder(outputDir)
		case "4":
			setupTemplateFile(filepath.Join(inputDir, "Template.txt"))
			fmt.Println("\nTemplate refreshed!")
			pause(reader)
		case "Q":
			return
		}
	}
}

func runFullSync(db *sql.DB, inputDir string, isBooklet bool) {
	fmt.Printf("\nGenerating PDFs (Booklet Mode: %v)...\n", isBooklet)
	files, _ := os.ReadDir(inputDir)

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(strings.ToLower(file.Name()), ".txt") {
			if strings.EqualFold(file.Name(), "Template.txt") {
				continue
			}
			recipe, _ := ParseFile(filepath.Join(inputDir, file.Name()))
			SaveRecipe(db, recipe)
			ExportToPDF(recipe, isBooklet)
			fmt.Printf("[OK] %s\n", recipe.Title)
		}
	}
}

// FIX: Added the missing ensureDirExists function
func ensureDirExists(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.Mkdir(path, 0755)
		if err != nil {
			log.Fatalf("Could not create folder %s: %v", path, err)
		}
	}
}

func pause(reader *bufio.Reader) {
	fmt.Println("\nPress Enter to continue...")
	reader.ReadString('\n')
}

func clearScreen() {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cls")
	} else {
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func openFolder(path string) {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("explorer", path)
	} else if runtime.GOOS == "darwin" {
		cmd = exec.Command("open", path)
	}
	if cmd != nil {
		cmd.Run()
	}
}

func setupTemplateFile(path string) {
	content := `RECIPE: 
TAGS: 

--- INGREDIENTS ---
- 

--- INSTRUCTIONS ---
1. 

NOTES: `
	_ = os.WriteFile(path, []byte(content), 0644)
}

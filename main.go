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

var Version = "development"

func main() {
	inputDir := "recipes_to_import"
	outputDir := "Printables"
	dbFile := "recipes.db"
	configFile := "cookbook_config.txt"

	ensureDirExists(inputDir)
	ensureDirExists(outputDir)

	db, err := InitDB(dbFile)
	if err != nil {
		log.Fatalf("Critical Error: %v", err)
	}
	defer db.Close()

	reader := bufio.NewReader(os.Stdin)

	for {
		clearScreen()
		count := GetRecipeCount(db)

		fmt.Println("########################################################")
		fmt.Printf("#             GOURMET RECIPE TRACKER %s              #\n", Version)
		fmt.Printf("#          Library Size: %d Recipes Saved               #\n", count)
		fmt.Println("########################################################")
		fmt.Println("#                                                      #")
		fmt.Println("#  [1] SYNC: Individual Letter PDFs                    #")
		fmt.Println("#  [2] SYNC: Individual Booklet PDFs                   #")
		fmt.Println("#  [3] OPEN: View 'Printables' Folder                  #")
		fmt.Println("#  [4] MASTER: Generate Custom Cookbook (Full/Booklet) #")
		fmt.Println("#  [5] CLEAN: Re-create Template.txt                   #")
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
			handleMasterCookbook(db, configFile, reader)
		case "5":
			setupTemplateFile(filepath.Join(inputDir, "Template.txt"))
			fmt.Println("\nTemplate refreshed!")
			pause(reader)
		case "Q":
			return
		}
	}
}

func handleMasterCookbook(db *sql.DB, configFile string, reader *bufio.Reader) {
	allRecipes, err := GetAllRecipes(db)
	if err != nil || len(allRecipes) == 0 {
		fmt.Println("\nNo recipes found in the database.")
		pause(reader)
		return
	}

	// 1. Generate/Refresh the config file
	var titles []string
	for _, r := range allRecipes {
		titles = append(titles, r.Title)
	}
	os.WriteFile(configFile, []byte(strings.Join(titles, "\n")), 0644)

	fmt.Println("\n--- CUSTOM COOKBOOK GENERATION ---")
	fmt.Printf("I have updated '%s' with all recipe names.\n", configFile)

	fmt.Print("Would you like to open it now to remove recipes? (Y/N): ")
	openChoice, _ := reader.ReadString('\n')
	if strings.TrimSpace(strings.ToUpper(openChoice)) == "Y" {
		var cmd *exec.Cmd
		if runtime.GOOS == "windows" {
			cmd = exec.Command("notepad.exe", configFile)
		} else if runtime.GOOS == "darwin" {
			cmd = exec.Command("open", "-e", configFile)
		}
		if cmd != nil {
			fmt.Println("Opening editor... please save and close when finished.")
			cmd.Run()
		}
	}

	// Ask for format
	fmt.Println("\nWhich format would you like for this cookbook?")
	fmt.Println("[1] Full Letter (Digital/Standard Printing)")
	fmt.Println("[2] Booklet (Half-Page/For Cutting & Binding)")
	fmt.Print("Selection: ")
	formatChoice, _ := reader.ReadString('\n')
	isBooklet := strings.TrimSpace(formatChoice) == "2"

	fmt.Print("\nPress Enter when you are ready to generate the PDF...")
	reader.ReadString('\n')

	// 2. Read back the edited file
	content, _ := os.ReadFile(configFile)
	var selectedRecipes []Recipe
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		title := strings.TrimSpace(line)
		if title == "" {
			continue
		}
		recipe, err := GetRecipeByTitle(db, title)
		if err == nil {
			selectedRecipes = append(selectedRecipes, recipe)
		}
	}

	if len(selectedRecipes) > 0 {
		fmt.Printf("Generating Master Cookbook (Booklet: %v)...\n", isBooklet)
		ExportMasterCookbook(selectedRecipes, isBooklet)
		fmt.Println("Success! Check your 'Printables' folder.")
	} else {
		fmt.Println("No valid recipes found in the config.")
	}
	pause(reader)
}

func runFullSync(db *sql.DB, inputDir string, isBooklet bool) {
	fmt.Printf("\nScanning for updates (Booklet: %v)...\n", isBooklet)
	files, _ := os.ReadDir(inputDir)
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(strings.ToLower(file.Name()), ".txt") {
			if strings.EqualFold(file.Name(), "Template.txt") {
				continue
			}
			recipe, err := ParseFile(filepath.Join(inputDir, file.Name()))
			if err == nil {
				SaveRecipe(db, recipe)
				ExportToPDF(recipe, isBooklet)
				fmt.Printf("[OK] %s\n", recipe.Title)
			}
		}
	}
}

func ensureDirExists(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, 0755)
	}
}

func pause(reader *bufio.Reader) {
	fmt.Println("\nPress Enter to return to menu...")
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
	content := "RECIPE: \nTAGS: \n\n--- INGREDIENTS ---\n- \n\n--- INSTRUCTIONS ---\n1. \n\nNOTES: "
	_ = os.WriteFile(path, []byte(content), 0644)
}

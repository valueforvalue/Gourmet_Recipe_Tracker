package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
)

var Version = "3.0-Web"

//go:embed web/index.html web/js/elm.js
var frontendFiles embed.FS

func main() {
	dbFile := "recipes.db"
	db, err := InitDB(dbFile)
	if err != nil {
		log.Fatalf("Critical Error: %v", err)
	}
	defer db.Close()

	ensureDirExists("Printables")
	ensureDirExists("backups")

	// 1. Serve static frontend
	webFiles, _ := fs.Sub(frontendFiles, "web")
	http.Handle("/", http.FileServer(http.FS(webFiles)))

	// 2. API: Fetch All
	http.HandleFunc("/api/recipes", func(w http.ResponseWriter, r *http.Request) {
		recipes, err := GetAllRecipes(db)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(recipes)
	})

	// 3. API: Save
	http.HandleFunc("/api/save", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var rcp Recipe
		if err := json.NewDecoder(r.Body).Decode(&rcp); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := SaveRecipe(db, rcp); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	fmt.Printf("Gourmet Tracker started at http://localhost:8080\n")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func ensureDirExists(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, 0755)
	}
}

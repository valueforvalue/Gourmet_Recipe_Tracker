package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

var Version = "3.2-Web"

//go:embed web/index.html web/js/elm.js
var frontendFiles embed.FS

func main() {
	dbFile := "recipes.db"
	outputDir := "Printables"

	ensureDirExists(outputDir)
	ensureDirExists("backups")

	db, err := InitDB(dbFile)
	if err != nil {
		log.Fatalf("Critical Error: %v", err)
	}
	defer db.Close()

	// 1. SERVE FRONTEND
	webFiles, _ := fs.Sub(frontendFiles, "web")
	http.Handle("/", http.FileServer(http.FS(webFiles)))

	// 2. API: FETCH ALL
	http.HandleFunc("/api/recipes", func(w http.ResponseWriter, r *http.Request) {
		recipes, err := GetAllRecipes(db)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(recipes)
	})

	// 3. API: SAVE
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

	// 4. API: DELETE
	http.HandleFunc("/api/delete", func(w http.ResponseWriter, r *http.Request) {
		title := r.URL.Query().Get("title")
		if err := DeleteRecipe(db, title); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	// 5. API: EXPORT SINGLE PDF
	http.HandleFunc("/api/export/pdf", func(w http.ResponseWriter, r *http.Request) {
		title := r.URL.Query().Get("title")
		isBooklet := r.URL.Query().Get("booklet") == "true"
		recipe, err := GetRecipeByTitle(db, title)
		if err != nil {
			http.Error(w, "Recipe not found", http.StatusNotFound)
			return
		}
		ExportToPDF(recipe, isBooklet)
		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s.pdf\"", title))
		http.ServeFile(w, r, filepath.Join(outputDir, title+".pdf"))
	})

	// 6. API: MASTER COOKBOOK (Filename Sync Fix)
	http.HandleFunc("/api/export/cookbook", func(w http.ResponseWriter, r *http.Request) {
		isBooklet := r.URL.Query().Get("booklet") == "true"
		recipes, err := GetAllRecipes(db)
		if err != nil {
			http.Error(w, "Could not fetch recipes", http.StatusInternalServerError)
			return
		}
		err = ExportMasterCookbook(recipes, isBooklet)
		if err != nil {
			http.Error(w, "PDF Generation failed: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Sync with the actual filenames used in pdf.go
		fileName := "Master_Cookbook_Full.pdf"
		if isBooklet {
			fileName = "Master_Cookbook_Booklet.pdf"
		}

		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", fileName))
		http.ServeFile(w, r, filepath.Join(outputDir, fileName))
	})

	fmt.Printf("Gourmet Tracker %s started at http://localhost:8080\n", Version)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func ensureDirExists(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, 0755)
	}
}

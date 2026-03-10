package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
)

var Version = "3.4"

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

	webFiles, _ := fs.Sub(frontendFiles, "web")
	http.Handle("/", http.FileServer(http.FS(webFiles)))

	// 1. API: FETCH ALL
	http.HandleFunc("/api/recipes", func(w http.ResponseWriter, r *http.Request) {
		recipes, err := GetAllRecipes(db)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(recipes)
	})

	// 2. API: SAVE
	http.HandleFunc("/api/save", func(w http.ResponseWriter, r *http.Request) {
		var rcp Recipe
		if err := json.NewDecoder(r.Body).Decode(&rcp); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		SaveRecipe(db, rcp)
		w.WriteHeader(http.StatusOK)
	})

	// 3. API: DELETE
	http.HandleFunc("/api/delete", func(w http.ResponseWriter, r *http.Request) {
		title := r.URL.Query().Get("title")
		DeleteRecipe(db, title)
		w.WriteHeader(http.StatusOK)
	})

	// 4. API: MASTER COOKBOOK
	http.HandleFunc("/api/export/cookbook", func(w http.ResponseWriter, r *http.Request) {
		isBooklet := r.URL.Query().Get("booklet") == "true"
		recipes, _ := GetAllRecipes(db)
		ExportMasterCookbook(recipes, isBooklet)

		fileName := "Master_Cookbook_Full.pdf"
		if isBooklet {
			fileName = "Master_Cookbook_Booklet.pdf"
		}

		w.Header().Set("Content-Type", "application/pdf")
		http.ServeFile(w, r, filepath.Join(outputDir, fileName))
	})

	// 5. API: EXPORT SINGLE PDF (Bug Fix: Distinct Filenames)
	http.HandleFunc("/api/export/pdf", func(w http.ResponseWriter, r *http.Request) {
		title := r.URL.Query().Get("title")
		isBooklet := r.URL.Query().Get("booklet") == "true"

		recipe, err := GetRecipeByTitle(db, title)
		if err != nil {
			http.Error(w, "Recipe not found", http.StatusNotFound)
			return
		}

		// Trigger PDF creation
		ExportToPDF(recipe, isBooklet)

		// Determine the filename that pdf.go created
		// Ensure your ExportToPDF function uses this same naming logic!
		suffix := "_Letter.pdf"
		if isBooklet {
			suffix = "_Booklet.pdf"
		}
		fileName := title + suffix

		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", fileName))

		filePath := filepath.Join(outputDir, fileName)
		http.ServeFile(w, r, filePath)
	})

	localIP := getLocalIP()
	port := "8080"

	fmt.Println("-----------------------------------------------")
	fmt.Println("       Morris Family Recipe Tracker v" + Version)
	fmt.Println("-----------------------------------------------")
	fmt.Printf(" [Internal Use]: http://localhost:%s\n", port)
	fmt.Printf(" [Mobile Use]:   http://%s:%s\n", localIP, port)
	fmt.Println("-----------------------------------------------")

	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "127.0.0.1"
}

func ensureDirExists(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, 0755)
	}
}

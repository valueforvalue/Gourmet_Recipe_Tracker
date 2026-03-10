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
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var Version = "3.6.1-Debug"

//go:embed web/index.html web/js/elm.js
var frontendFiles embed.FS

func main() {
	dbFile := "recipes.db"
	outputDir := "Printables"

	ensureDirExists(outputDir)
	ensureDirExists("backups")

	// --- DATABASE DIAGNOSTICS ---
	absPath, _ := filepath.Abs(dbFile)
	fmt.Println("-----------------------------------------------")
	fmt.Printf(" [DB PATH]: %s\n", absPath)

	db, err := InitDB(dbFile)
	if err != nil {
		log.Fatalf("Critical Error opening DB: %v", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("Database connection is dead: %v", err)
	}
	fmt.Println(" [STATUS]: Database connection verified.")
	fmt.Println("-----------------------------------------------")
	defer db.Close()

	// --- FRONTEND HANDLER ---
	webFiles, _ := fs.Sub(frontendFiles, "web")
	http.Handle("/", http.FileServer(http.FS(webFiles)))

	// --- API ROUTES ---

	// 1. Fetch All Recipes
	http.HandleFunc("/api/recipes", func(w http.ResponseWriter, r *http.Request) {
		recipes, err := GetAllRecipes(db)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(recipes)
	})

	// 2. Save/Update Recipe
	http.HandleFunc("/api/save", func(w http.ResponseWriter, r *http.Request) {
		var rcp Recipe
		if err := json.NewDecoder(r.Body).Decode(&rcp); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		err := SaveRecipe(db, rcp)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	// 3. Delete Recipe (Soft Delete)
	http.HandleFunc("/api/delete", func(w http.ResponseWriter, r *http.Request) {
		title := r.URL.Query().Get("title")
		DeleteRecipe(db, title)
		w.WriteHeader(http.StatusOK)
	})

	// 4. Scrape Recipe from URL
	http.HandleFunc("/api/scrape", func(w http.ResponseWriter, r *http.Request) {
		targetURL := r.URL.Query().Get("url")
		if targetURL == "" {
			http.Error(w, "URL required", http.StatusBadRequest)
			return
		}

		resp, err := http.Get(targetURL)
		if err != nil {
			http.Error(w, "Failed to reach site", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			http.Error(w, "Failed to parse HTML", http.StatusInternalServerError)
			return
		}

		var extracted Recipe
		extracted.Title = strings.TrimSpace(doc.Find("h1").First().Text())
		extracted.Notes = "Source: " + targetURL
		extracted.Tags = []string{"Imported"}

		// Heuristic search for ingredients
		doc.Find("li").Each(func(i int, s *goquery.Selection) {
			txt := strings.TrimSpace(s.Text())
			parent, _ := s.Parent().Attr("class")
			pLower := strings.ToLower(parent)
			tLower := strings.ToLower(txt)

			if strings.Contains(pLower, "ingredient") ||
				strings.Contains(tLower, " cup ") ||
				strings.Contains(tLower, " tsp ") ||
				strings.Contains(tLower, " tbsp ") {
				extracted.Ingredients = append(extracted.Ingredients, txt)
			}
		})

		// Heuristic search for instructions
		doc.Find("li").Each(func(i int, s *goquery.Selection) {
			parent, _ := s.Parent().Attr("class")
			pLower := strings.ToLower(parent)
			if strings.Contains(pLower, "instruction") || strings.Contains(pLower, "step") {
				extracted.Instructions = append(extracted.Instructions, strings.TrimSpace(s.Text()))
			}
		})

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(extracted)
	})

	// 5. PDF Exports
	http.HandleFunc("/api/export/pdf", func(w http.ResponseWriter, r *http.Request) {
		title := r.URL.Query().Get("title")
		isBooklet := r.URL.Query().Get("booklet") == "true"
		recipe, _ := GetRecipeByTitle(db, title)
		ExportToPDF(recipe, isBooklet)

		suffix := "_Letter.pdf"
		if isBooklet {
			suffix = "_Booklet.pdf"
		}

		w.Header().Set("Content-Type", "application/pdf")
		http.ServeFile(w, r, filepath.Join(outputDir, title+suffix))
	})

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

	// --- SERVER START ---
	localIP := getLocalIP()
	port := "8080"
	fmt.Println("       Morris Family Recipe Tracker")
	fmt.Printf(" [Internal]: http://localhost:%s\n", port)
	fmt.Printf(" [Mobile]:   http://%s:%s\n", localIP, port)
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

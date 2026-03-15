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

var Version = "3.8-PrettyPDF"

//go:embed web/index.html web/js/elm.js
var frontendFiles embed.FS

func main() {
	dbFile := "recipes.db"
	ensureDirExists("Printables")
	ensureDirExists("backups")

	fmt.Println("-----------------------------------------------")
	fmt.Printf(" Morris Family Recipe Tracker v%s\n", Version)
	if enableGitSync {
		fmt.Printf(" [!] GIT SYNC: ENABLED (Target: %s)\n", backupRepoPath)
	} else {
		fmt.Println(" [ ] GIT SYNC: DISABLED (Local only mode)")
	}

	db, err := InitDB(dbFile)
	if err != nil {
		log.Fatalf("Critical Error: %v", err)
	}
	defer db.Close()

	webFiles, _ := fs.Sub(frontendFiles, "web")
	http.Handle("/", http.FileServer(http.FS(webFiles)))

	// API ROUTES
	http.HandleFunc("/api/recipes", func(w http.ResponseWriter, r *http.Request) {
		recipes, _ := GetAllRecipes(db)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(recipes)
	})

	http.HandleFunc("/api/save", func(w http.ResponseWriter, r *http.Request) {
		var rcp Recipe
		if err := json.NewDecoder(r.Body).Decode(&rcp); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		SaveRecipe(db, rcp)
		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("/api/delete", func(w http.ResponseWriter, r *http.Request) {
		title := r.URL.Query().Get("title")
		DeleteRecipe(db, title)
		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("/api/scrape", func(w http.ResponseWriter, r *http.Request) {
		targetURL := r.URL.Query().Get("url")
		resp, err := http.Get(targetURL)
		if err != nil {
			return
		}
		defer resp.Body.Close()
		doc, _ := goquery.NewDocumentFromReader(resp.Body)

		var extracted Recipe
		extracted.Title = strings.TrimSpace(doc.Find("h1").First().Text())
		extracted.Notes = "Source: " + targetURL
		extracted.Tags = []string{"Imported"}

		doc.Find("li").Each(func(i int, s *goquery.Selection) {
			txt := strings.TrimSpace(s.Text())
			tLower := strings.ToLower(txt)
			if strings.Contains(tLower, " cup ") || strings.Contains(tLower, " tsp ") || strings.Contains(tLower, " tbsp ") {
				extracted.Ingredients = append(extracted.Ingredients, txt)
			}
		})
		doc.Find("li").Each(func(i int, s *goquery.Selection) {
			parent, _ := s.Parent().Attr("class")
			if strings.Contains(strings.ToLower(parent), "instruction") {
				extracted.Instructions = append(extracted.Instructions, strings.TrimSpace(s.Text()))
			}
		})
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(extracted)
	})

	http.HandleFunc("/api/export/pdf", func(w http.ResponseWriter, r *http.Request) {
		title := r.URL.Query().Get("title")
		isBooklet := r.URL.Query().Get("booklet") == "true"
		recipe, _ := GetRecipeByTitle(db, title)

		ExportToPDF(recipe, isBooklet) // Call fixed in signature

		suffix := "_Letter.pdf"
		if isBooklet {
			suffix = "_Booklet.pdf"
		}
		w.Header().Set("Content-Type", "application/pdf")
		http.ServeFile(w, r, filepath.Join("Printables", title+suffix))
	})

	http.HandleFunc("/api/export/cookbook", func(w http.ResponseWriter, r *http.Request) {
		isBooklet := r.URL.Query().Get("booklet") == "true"
		recipes, _ := GetAllRecipes(db)

		ExportMasterCookbook(recipes, isBooklet) // Call fixed in signature

		fileName := "Master_Cookbook_Full.pdf"
		if isBooklet {
			fileName = "Master_Cookbook_Booklet.pdf"
		}
		w.Header().Set("Content-Type", "application/pdf")
		http.ServeFile(w, r, filepath.Join("Printables", fileName))
	})

	localIP := getLocalIP()
	fmt.Printf(" [Mobile Access]: http://%s:8080\n", localIP)
	fmt.Println("-----------------------------------------------")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}

	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			// Check if it's IPv4 AND explicitly ensure it is not a 169.254.x.x address
			if ipnet.IP.To4() != nil && !ipnet.IP.IsLinkLocalUnicast() {
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

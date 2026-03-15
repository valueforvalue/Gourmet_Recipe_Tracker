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

var Version = "4.0-MasterCookbook"

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
			// FIXED: Return a real error to the frontend instead of silently failing
			http.Error(w, "Failed to fetch URL", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			http.Error(w, fmt.Sprintf("Website returned status code: %d", resp.StatusCode), http.StatusBadRequest)
			return
		}

		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			http.Error(w, "Failed to parse HTML", http.StatusInternalServerError)
			return
		}

		var extracted Recipe
		extracted.Title = strings.TrimSpace(doc.Find("h1").First().Text())
		extracted.Notes = "Source: " + targetURL
		extracted.Tags = []string{"Imported"}

		// FIXED INGREDIENT SCRAPER: Target CSS classes instead of specific measurement words
		doc.Find("li").Each(func(i int, s *goquery.Selection) {
			itemClass, _ := s.Attr("class")
			parentClass, _ := s.Parent().Attr("class")
			grandparentClass, _ := s.Parent().Parent().Attr("class")

			combinedClasses := strings.ToLower(itemClass + " " + parentClass + " " + grandparentClass)

			if strings.Contains(combinedClasses, "ingredient") {
				// strings.Fields cleans up messy spacing and newlines from nested HTML tags
				txt := strings.Join(strings.Fields(s.Text()), " ")
				if txt != "" {
					extracted.Ingredients = append(extracted.Ingredients, txt)
				}
			}
		})

		// FIXED INSTRUCTION SCRAPER: Broaden search to include directions and steps
		doc.Find("li").Each(func(i int, s *goquery.Selection) {
			itemClass, _ := s.Attr("class")
			parentClass, _ := s.Parent().Attr("class")
			grandparentClass, _ := s.Parent().Parent().Attr("class")

			combinedClasses := strings.ToLower(itemClass + " " + parentClass + " " + grandparentClass)

			if strings.Contains(combinedClasses, "instruction") || strings.Contains(combinedClasses, "direction") || strings.Contains(combinedClasses, "step") {
				txt := strings.Join(strings.Fields(s.Text()), " ")
				if txt != "" {
					extracted.Instructions = append(extracted.Instructions, txt)
				}
			}
		})

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(extracted)
	})

	http.HandleFunc("/api/export/pdf", func(w http.ResponseWriter, r *http.Request) {
		title := r.URL.Query().Get("title")
		isBooklet := r.URL.Query().Get("booklet") == "true"
		recipe, _ := GetRecipeByTitle(db, title)

		ExportToPDF(recipe, isBooklet)

		suffix := "_Letter.pdf"
		if isBooklet {
			suffix = "_Booklet.pdf"
		}

		fileName := title + suffix

		w.Header().Set("Content-Type", "application/pdf")
		// FIXED: Tell the browser to download as an attachment with the correct filename
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))

		http.ServeFile(w, r, filepath.Join("Printables", fileName))
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
		// FIXED: Tell the browser to download as an attachment with the correct filename
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))

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

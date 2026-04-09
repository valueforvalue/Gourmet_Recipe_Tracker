package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var Version = "4.0-MasterCookbook"

//go:embed web/index.html web/js/elm.js
var frontendFiles embed.FS

func main() {
	// Initialize Logging
	logFile, err := os.OpenFile("tracker.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf(" [!] Warning: Could not create log file: %v\n", err)
	} else {
		defer logFile.Close()
		multi := io.MultiWriter(os.Stdout, logFile)
		log.SetOutput(multi)
		log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	}

	LoadConfig()
	InitConfig()

	dbFile := GlobalConfig.DBFile
	ensureDirExists("Printables")
	ensureDirExists("backups")

	// Start the background Git sync worker
	if GlobalConfig.GitSync {
		go StartSyncWorker()
	}

	log.Println("-----------------------------------------------")
	log.Printf(" Morris Family Recipe Tracker v%s", Version)
	if GlobalConfig.GitSync {
		log.Printf(" [!] GIT SYNC: ENABLED (Target: %s)", GlobalConfig.BackupPath)
	} else {
		log.Println(" [ ] GIT SYNC: DISABLED (Local only mode)")
	}

	db, err := InitDB(dbFile)
	if err != nil {
		log.Fatalf("Critical Error: %v", err)
	}
	defer db.Close()

	webFiles, err := fs.Sub(frontendFiles, "web")
	if err != nil {
		log.Fatalf("Critical Error: Could not load embedded web files: %v", err)
	}
	http.Handle("/", http.FileServer(http.FS(webFiles)))

	// API ROUTES
	http.HandleFunc("/api/recipes", func(w http.ResponseWriter, r *http.Request) {
		log.Println(" [API] GET /api/recipes")
		recipes, err := GetAllRecipes(db)
		if err != nil {
			log.Printf(" [Error] GetAllRecipes: %v", err)
			http.Error(w, "Failed to retrieve recipes", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(recipes)
	})

	http.HandleFunc("/api/save", func(w http.ResponseWriter, r *http.Request) {
		var rcp Recipe
		if err := json.NewDecoder(r.Body).Decode(&rcp); err != nil {
			log.Printf(" [Error] POST /api/save (Decode): %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		log.Printf(" [API] POST /api/save: %s", rcp.Title)
		err := SaveRecipe(db, rcp)
		if err != nil {
			log.Printf(" [Error] SaveRecipe: %v", err)
			http.Error(w, "Failed to save recipe", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("/api/delete", func(w http.ResponseWriter, r *http.Request) {
		title := r.URL.Query().Get("title")
		log.Printf(" [API] POST /api/delete: %s", title)
		err := DeleteRecipe(db, title)
		if err != nil {
			log.Printf(" [Error] DeleteRecipe: %v", err)
			http.Error(w, "Failed to delete recipe", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("/api/scrape", func(w http.ResponseWriter, r *http.Request) {
		targetURL := r.URL.Query().Get("url")
		log.Printf(" [API] GET /api/scrape: %s", targetURL)
		if !isSafeURL(targetURL) {
			log.Printf(" [Error] Scrape blocked unsafe URL: %s", targetURL)
			http.Error(w, "Invalid or disallowed URL", http.StatusBadRequest)
			return
		}
		resp, err := http.Get(targetURL)
		if err != nil {
			log.Printf(" [Error] Scrape Fetch: %v", err)
			http.Error(w, "Failed to fetch URL", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			log.Printf(" [Error] Scrape Status: %d", resp.StatusCode)
			http.Error(w, fmt.Sprintf("Website returned status code: %d", resp.StatusCode), http.StatusBadRequest)
			return
		}

		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			log.Printf(" [Error] Scrape Parse: %v", err)
			http.Error(w, "Failed to parse HTML", http.StatusInternalServerError)
			return
		}

		var extracted Recipe
		extracted.Title = strings.Join(strings.Fields(doc.Find("h1").First().Text()), " ")
		extracted.Notes = "Source: " + targetURL
		extracted.Tags = []string{"Imported"}

		// FIXED INGREDIENT SCRAPER: Target CSS classes instead of specific measurement words
		doc.Find("li").Each(func(i int, s *goquery.Selection) {
			itemClass, _ := s.Attr("class")
			parentClass, _ := s.Parent().Attr("class")
			grandparentClass, _ := s.Parent().Parent().Attr("class")

			combinedClasses := strings.ToLower(itemClass + " " + parentClass + " " + grandparentClass)

			if strings.Contains(combinedClasses, "ingredient") {
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
		log.Printf(" [API] GET /api/export/pdf: %s", title)
		isBooklet := r.URL.Query().Get("booklet") == "true"
		recipe, err := GetRecipeByTitle(db, title)
		if err != nil {
			log.Printf(" [Error] GetRecipeByTitle: %v", err)
			http.Error(w, "Recipe not found", http.StatusNotFound)
			return
		}

		if err := ExportToPDF(recipe, isBooklet); err != nil {
			log.Printf(" [Error] ExportToPDF: %v", err)
			http.Error(w, "Failed to generate PDF", http.StatusInternalServerError)
			return
		}

		suffix := "_Letter.pdf"
		if isBooklet {
			suffix = "_Booklet.pdf"
		}

		fileName := sanitizeFilename(title) + suffix

		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))

		http.ServeFile(w, r, filepath.Join("Printables", fileName))
	})

	http.HandleFunc("/api/export/cookbook", func(w http.ResponseWriter, r *http.Request) {
		log.Println(" [API] GET /api/export/cookbook")
		isBooklet := r.URL.Query().Get("booklet") == "true"
		recipes, err := GetAllRecipes(db)
		if err != nil {
			log.Printf(" [Error] GetAllRecipes (cookbook): %v", err)
			http.Error(w, "Failed to retrieve recipes", http.StatusInternalServerError)
			return
		}

		if err := ExportMasterCookbook(recipes, isBooklet); err != nil {
			log.Printf(" [Error] ExportMasterCookbook: %v", err)
			http.Error(w, "Failed to generate cookbook PDF", http.StatusInternalServerError)
			return
		}

		fileName := "Master_Cookbook_Full.pdf"
		if isBooklet {
			fileName = "Master_Cookbook_Booklet.pdf"
		}

		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))

		http.ServeFile(w, r, filepath.Join("Printables", fileName))
	})

	localIP := getLocalIP()
	log.Printf(" [Mobile Access]: http://%s:%s", localIP, GlobalConfig.Port)
	log.Println("-----------------------------------------------")
	log.Fatal(http.ListenAndServe(":"+GlobalConfig.Port, nil))
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
		if err := os.Mkdir(path, 0755); err != nil {
			log.Printf(" [!] Warning: Could not create directory %s: %v", path, err)
		}
	}
}

// isSafeURL validates that a URL uses http/https and does not target
// loopback, private, or link-local addresses to prevent SSRF attacks.
// It also resolves hostnames to catch DNS rebinding attempts.
func isSafeURL(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return false
	}
	host := u.Hostname()
	if host == "" || host == "localhost" {
		return false
	}
	if ip := net.ParseIP(host); ip != nil {
		return isSafeIP(ip)
	}
	// Resolve hostname and validate all resulting IPs to guard against DNS rebinding.
	addrs, err := net.LookupHost(host)
	if err != nil {
		return false
	}
	for _, addr := range addrs {
		if ip := net.ParseIP(addr); ip != nil && !isSafeIP(ip) {
			return false
		}
	}
	return true
}

// isSafeIP returns true when ip is a globally routable unicast address.
func isSafeIP(ip net.IP) bool {
	return !ip.IsLoopback() && !ip.IsPrivate() && !ip.IsLinkLocalUnicast() && !ip.IsUnspecified()
}

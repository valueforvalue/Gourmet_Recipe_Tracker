package main

import (
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// Version is the build version reported by -version and the startup banner.
var Version = "4.1-SystemOverhaul"

// purgeConfirmWord is the exact phrase a client must echo to authorize a
// permanent (hard) delete.
const purgeConfirmWord = "yesdeary"

//go:embed web/index.html web/js/elm.js
var frontendFiles embed.FS

// scrapeClient bounds how long a recipe scrape may take so a slow site can't
// tie up a server goroutine forever. Its redirect policy re-validates every
// hop so a safe URL cannot bounce to an internal address.
var scrapeClient = &http.Client{
	Timeout:       30 * time.Second,
	CheckRedirect: checkScrapeRedirect,
}

// corsMiddleware wraps a handler and adds CORS headers so the Flutter
// Android app (and any other cross-origin client) can reach the API.
// The wildcard origin is intentional for a home-server deployment where
// the server is only reachable inside the local network.  If you expose
// this server to the internet, add a reverse-proxy (e.g. Caddy/nginx)
// with HTTPS and authentication in front of it.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// recoveryMiddleware prevents a panicking handler from crashing the whole
// server.  It logs the panic and returns a 500 to the client.
func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				logError("panic %s %s: %v", r.Method, r.URL.Path, rec)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// loggingMiddleware logs every request with its duration.
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		logInfo("api %s %s (%s)", r.Method, r.URL.Path, time.Since(start))
	})
}

func main() {
	versionFlag := flag.Bool("version", false, "print version and exit")
	restoreFlag := flag.Bool("restore", false, "restore the database from the vault and exit")
	restoreDirFlag := flag.String("restore-dir", "", "vault directory to restore from (defaults to configured vault)")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("Morris Family Recipe Tracker v%s\n", Version)
		return
	}

	// Initialize Logging
	logFile, err := setupLogging()
	if err != nil {
		fmt.Printf(" [!] Warning: Could not create log file: %v\n", err)
		log.SetOutput(os.Stdout)
		log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	} else {
		defer logFile.Close()
	}

	if err := LoadConfig(); err != nil {
		log.Fatalf("Critical Error: %v", err)
	}
	InitConfig()

	dbFile := GlobalConfig.DBFile
	ensureDirExists("Printables")
	ensureDirExists("backups")

	db, err := InitDB(dbFile)
	if err != nil {
		log.Fatalf("Critical Error: %v", err)
	}
	defer db.Close()

	// Restore mode: rebuild the SQLite database from the vault and exit.
	if *restoreFlag {
		dir := *restoreDirFlag
		if dir == "" {
			dir = vaultDir()
		}
		n, err := RestoreFromVault(dir, db)
		if err != nil {
			log.Fatalf("Restore failed: %v", err)
		}
		logInfo("restored %d recipe(s) from %s", n, dir)
		return
	}

	// Start the background Git sync worker
	if GlobalConfig.GitSync {
		go StartSyncWorker()
	}

	logInfo("-----------------------------------------------")
	logInfo("Morris Family Recipe Tracker v%s", Version)
	if GlobalConfig.GitSync {
		logInfo("git sync enabled (target: %s)", GlobalConfig.BackupPath)
	} else {
		logInfo("git sync disabled (local only mode)")
	}

	mux := http.NewServeMux()

	webFiles, err := fs.Sub(frontendFiles, "web")
	if err != nil {
		log.Fatalf("Critical Error: Could not load embedded web files: %v", err)
	}
	mux.Handle("/", http.FileServer(http.FS(webFiles)))

	// API ROUTES
	mux.HandleFunc("/api/recipes", func(w http.ResponseWriter, r *http.Request) {
		recipes, err := GetAllRecipes(db)
		if err != nil {
			logError("GetAllRecipes: %v", err)
			http.Error(w, "Failed to load recipes", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(recipes)
	})

	mux.HandleFunc("/api/save", handleSaveRecipe(db))

	mux.HandleFunc("/api/delete", handleDeleteRecipe(db))

	mux.HandleFunc("/api/deleted", handleDeleted(db))
	mux.HandleFunc("/api/restore", handleRestore(db))
	mux.HandleFunc("/api/purge", handlePurge(db))
	mux.HandleFunc("/api/purge-all", handlePurgeAll(db))

	mux.HandleFunc("/api/scrape", func(w http.ResponseWriter, r *http.Request) {
		targetURL := r.URL.Query().Get("url")
		logInfo("scrape %s", targetURL)
		if !isSafeURL(targetURL) {
			logError("scrape blocked unsafe URL: %s", targetURL)
			http.Error(w, "Invalid or disallowed URL", http.StatusBadRequest)
			return
		}
		resp, err := scrapeClient.Get(targetURL)
		if err != nil {
			logError("scrape fetch: %v", err)
			http.Error(w, "Failed to fetch URL", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			logError("scrape status: %d", resp.StatusCode)
			http.Error(w, fmt.Sprintf("Website returned status code: %d", resp.StatusCode), http.StatusBadRequest)
			return
		}

		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			logError("scrape parse: %v", err)
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

	mux.HandleFunc("/api/export/pdf", func(w http.ResponseWriter, r *http.Request) {
		title := r.URL.Query().Get("title")
		logInfo("export pdf: %s", title)
		isBooklet := r.URL.Query().Get("booklet") == "true"
		recipe, err := GetRecipeByTitle(db, title)
		if err != nil {
			logError("GetRecipeByTitle: %v", err)
			http.Error(w, "Recipe not found", http.StatusNotFound)
			return
		}

		if err := ExportToPDF(recipe, isBooklet); err != nil {
			logError("ExportToPDF: %v", err)
			http.Error(w, "Failed to export PDF", http.StatusInternalServerError)
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

	mux.HandleFunc("/api/export/cookbook", func(w http.ResponseWriter, r *http.Request) {
		isBooklet := r.URL.Query().Get("booklet") == "true"
		recipes, err := GetAllRecipes(db)
		if err != nil {
			logError("GetAllRecipes: %v", err)
			http.Error(w, "Failed to load recipes", http.StatusInternalServerError)
			return
		}

		if err := ExportMasterCookbook(recipes, isBooklet); err != nil {
			logError("ExportMasterCookbook: %v", err)
			http.Error(w, "Failed to export cookbook", http.StatusInternalServerError)
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
	logInfo("mobile access: http://%s:%s", localIP, GlobalConfig.Port)
	logInfo("-----------------------------------------------")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	handler := corsMiddleware(loggingMiddleware(recoveryMiddleware(mux)))
	_ = runSupervised(ctx, ":"+GlobalConfig.Port, handler)
	logInfo("server stopped.")
}

// newHTTPServer builds an http.Server with conservative timeouts so a slow
// client (slowloris) cannot tie up connections or goroutines indefinitely.
func newHTTPServer(handler http.Handler) *http.Server {
	return &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}

// nextBackoff returns the next restart delay, doubling prev up to a 30s cap.
func nextBackoff(prev time.Duration) time.Duration {
	const cap = 30 * time.Second
	if prev <= 0 {
		return time.Second
	}
	next := prev * 2
	if next > cap {
		return cap
	}
	return next
}

// shouldRestart reports whether a Serve return value warrants a restart. A
// nil return and http.ErrServerClosed (graceful shutdown) exit cleanly;
// anything else is a crash worth retrying.
func shouldRestart(err error) bool {
	return err != nil && !errors.Is(err, http.ErrServerClosed)
}

// sleepCtx blocks for d or until ctx is cancelled. It returns false when ctx
// was cancelled so the caller can stop retrying.
func sleepCtx(ctx context.Context, d time.Duration) bool {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-t.C:
		return true
	}
}

// runSupervised serves handler on addr until ctx is cancelled, restarting on
// transient listen/serve failures with exponential backoff. It returns nil on
// graceful shutdown.
func runSupervised(ctx context.Context, addr string, handler http.Handler) error {
	backoff := time.Second
	for {
		srv := newHTTPServer(handler)

		ln, err := net.Listen("tcp", addr)
		if err != nil {
			logWarn("listen %s: %v; retry in %s", addr, err, backoff)
			if !sleepCtx(ctx, backoff) {
				return nil
			}
			backoff = nextBackoff(backoff)
			continue
		}

		// Shut this server down when ctx is cancelled.
		go func(s *http.Server) {
			<-ctx.Done()
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			_ = s.Shutdown(shutdownCtx)
		}(srv)

		logInfo("listening on %s", addr)
		err = srv.Serve(ln)
		if !shouldRestart(err) {
			return nil
		}
		logError("server crashed: %v; restart in %s", err, backoff)
		if !sleepCtx(ctx, backoff) {
			return nil
		}
		backoff = nextBackoff(backoff)
	}
}

// checkScrapeRedirect is the redirect policy for scrapeClient. It blocks any
// hop that isSafeURL rejects and caps the chain length, closing the SSRF gap
// where a safe first URL redirects to a loopback/private address.
func checkScrapeRedirect(req *http.Request, via []*http.Request) error {
	if len(via) >= 10 {
		return errors.New("stopped after 10 redirects")
	}
	if !isSafeURL(req.URL.String()) {
		return fmt.Errorf("redirect to unsafe URL blocked: %s", req.URL.String())
	}
	return nil
}

// handleSaveRecipe is the POST /api/save handler. It rejects non-POST methods
// and persists a recipe (DB + vault mirror).
func handleSaveRecipe(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var rcp Recipe
		if err := json.NewDecoder(r.Body).Decode(&rcp); err != nil {
			logError("save decode: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		rcp.Title = strings.TrimSpace(rcp.Title)
		if rcp.Title == "" {
			http.Error(w, "Recipe title is required", http.StatusBadRequest)
			return
		}
		logInfo("save: %s", rcp.Title)
		if err := SaveRecipe(db, rcp); err != nil {
			logError("SaveRecipe: %v", err)
			http.Error(w, "Failed to save recipe", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

// handleDeleteRecipe is the POST /api/delete handler. It rejects non-POST
// methods and soft-deletes the recipe named by ?title=.
func handleDeleteRecipe(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		title := r.URL.Query().Get("title")
		logInfo("delete: %s", title)
		if err := DeleteRecipe(db, title); err != nil {
			logError("DeleteRecipe: %v", err)
			http.Error(w, "Failed to delete recipe", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

// handleDeleted is the GET /api/deleted handler. It lists soft-deleted
// recipes.
func handleDeleted(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		recipes, err := GetDeletedRecipes(db)
		if err != nil {
			logError("GetDeletedRecipes: %v", err)
			http.Error(w, "Failed to load deleted recipes", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(recipes)
	}
}

// handleRestore is the POST /api/restore handler. It un-deletes the recipe
// named by ?title= and re-mirrors it to the vault.
func handleRestore(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		title := r.URL.Query().Get("title")
		logInfo("restore: %s", title)
		if err := RestoreRecipe(db, title); err != nil {
			logError("RestoreRecipe: %v", err)
			http.Error(w, "Failed to restore recipe", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

// handlePurge is the POST /api/purge handler. It permanently removes the
// deleted recipe named by ?title= after the caller echoes purgeConfirmWord.
func handlePurge(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Query().Get("confirm") != purgeConfirmWord {
			http.Error(w, "type '"+purgeConfirmWord+"' to confirm permanent delete", http.StatusBadRequest)
			return
		}
		title := r.URL.Query().Get("title")
		logInfo("purge: %s", title)
		if err := PurgeRecipe(db, title); err != nil {
			logError("PurgeRecipe: %v", err)
			http.Error(w, "Failed to purge recipe", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

// handlePurgeAll is the POST /api/purge-all handler. It permanently removes
// every soft-deleted recipe after the caller echoes purgeConfirmWord.
func handlePurgeAll(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Query().Get("confirm") != purgeConfirmWord {
			http.Error(w, "type '"+purgeConfirmWord+"' to confirm permanent delete", http.StatusBadRequest)
			return
		}
		logInfo("purge all")
		if err := PurgeAllDeleted(db); err != nil {
			logError("PurgeAllDeleted: %v", err)
			http.Error(w, "Failed to purge deleted recipes", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
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
			logWarn("could not create directory %s: %v", path, err)
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

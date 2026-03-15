# Morris Family Recipe Tracker v4.0

A self-hosted, web-based application built in Go for managing a digital recipe library, scraping recipes from the web, and generating professional, auto-categorized PDF cookbooks.

## Core Features
* **Web-Based UI**: Manage your entire library from your browser or mobile device via a responsive frontend.
* **Smart Web Scraper**: Paste a URL to automatically extract recipe titles, ingredients, and instructions directly from food blogs.
* **Automated Master Cookbook**: Instantly compile your entire library into a single PDF with a generated Table of Contents, automatically grouped and alphabetized by recipe tags (e.g., "Dinner", "Dessert").
* **Dual PDF Modes**: Export individual recipes or the Master Cookbook in **Standard Letter (8.5" x 11")** or **Booklet (5.5" x 8.5")** sizes with perfect text-wrapping.
* **"Daily Bread" Git Sync**: Automatically saves plain-text copies of every recipe and pushes them to a local or remote Git repository for future-proof backup.
* **SQLite Backend**: Maintains a robust, searchable database to prevent duplicates and safely handle soft-deletes.

## Usage Instructions

### 1. Running the Server
Run the application from your terminal:
`go run .`

The console will display your local IP address. Open your web browser and navigate to:
* **Desktop:** `http://localhost:8080`
* **Mobile:** `http://<YOUR_LOCAL_IP>:8080`

### 2. Adding Recipes
* **Web Import:** Use the built-in scraper to paste a URL from a recipe site. The app will automatically extract and format the ingredients and steps.
* **Manual Entry:** Use the web interface to type out recipes, tag them, and save them directly to the database.

### 3. Backups & Git Sync
Text file backups are automatically generated in the `/backups` folder. If `enableGitSync` is set to `true` in `database.go`, these text files are instantly vaulted and pushed to your configured GitHub repository whenever a recipe is saved.

### 4. Printing for Physical Use (The "2-Up" Method)
To print Booklet-sized PDFs (5.5" x 8.5") efficiently onto standard paper to build a physical recipe box or binder:
1. Open the generated PDF and select **Print**.
2. Set Orientation to **Landscape**.
3. Set **Multiple Pages per Sheet** to **2**.
4. Cut the printed 8.5" x 11" sheets in half. The generated margins ensure enough space for standard hole-punching.

## Build Information

### Prerequisites
* [Go](https://go.dev/dl/) (v1.21+)
* Git (if using the auto-sync feature)

### Dependencies
Install the required Go modules before building:
`go get modernc.org/sqlite`
`go get github.com/jung-kurt/gofpdf`
`go get github.com/PuerkitoBio/goquery`

### Compiling
To build a standalone executable that you can run without the Go toolchain:
`go build -o RecipeTracker.exe .`
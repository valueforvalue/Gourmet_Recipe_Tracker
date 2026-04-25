# Morris Family Recipe Tracker v4.0

A self-hosted application built in Go for managing a digital recipe library, scraping recipes from the web, and generating professional, auto-categorized PDF cookbooks.  The same Go server powers **both** the browser-based web UI and an optional **Flutter Android app** — the SQLite database lives on the server and is never duplicated.

## Core Features
* **Web-Based UI**: Manage your entire library from any browser (desktop or mobile).
* **Android App** *(experimental)*: A native Flutter app that talks directly to the server's REST API, providing the same functionality on Android without duplicating the database.
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
* [Flutter SDK](https://flutter.dev/docs/get-started/install) 3.x+ (Android app only)

### Dependencies
Install the required Go modules before building:
`go get modernc.org/sqlite`
`go get github.com/jung-kurt/gofpdf`
`go get github.com/PuerkitoBio/goquery`

### Compiling
To build a standalone executable that you can run without the Go toolchain:
`go build -o RecipeTracker.exe .`

---

## Android App (Flutter)

The `flutter_app/` directory contains a Flutter project that provides the same features as the web interface through the server's existing REST API.  **The database schema and file are never modified — the app is purely a client.**

### Architecture overview

```
┌─────────────────────────────┐
│  Home Server (Go)           │
│  ─────────────────────────  │
│  /               (web UI)   │
│  /api/recipes               │
│  /api/save                  │
│  /api/delete                │
│  /api/scrape                │
│  /api/export/pdf            │
│  /api/export/cookbook       │
│                             │
│  recipes.db  (SQLite)       │
└──────────┬──────────────────┘
           │  HTTP  (local network)
     ┌─────┴──────┐
     │            │
 Chrome / Safari  Flutter Android App
 (unchanged)      (flutter_app/)
```

The server now adds **CORS headers** to every response so the Flutter app — running on a different host/port — can reach the API without browser security errors.  The web interface is completely unaffected because browsers follow CORS rules; the headers are simply ignored for same-origin requests.

### Flutter app features
| Feature | Web UI | Flutter App |
|---|---|---|
| Browse & search recipes | ✅ | ✅ |
| View recipe details | ✅ | ✅ |
| Create / edit recipes | ✅ | ✅ |
| Delete recipes (soft) | ✅ | ✅ |
| Import recipe from URL | ✅ | ✅ |
| Export single PDF | ✅ | ✅ |
| Export master cookbook PDF | ✅ | ✅ |
| Configurable server URL | — | ✅ |

### Building the Android app

1. Install [Flutter SDK 3.x+](https://flutter.dev/docs/get-started/install).
2. Run `flutter pub get` inside `flutter_app/`.
3. Connect an Android device or start an emulator.
4. Run `flutter run` — or build a release APK with `flutter build apk --release`.

### Connecting to your server

On first launch the app shows a **Settings** screen where you enter your server's address, e.g. `http://192.168.1.100:8080`.  This is saved on-device and can be updated at any time via the ⚙ icon in the app bar.

> **Tip:** You can reach your server outside your home network by configuring port-forwarding on your router or using a VPN (e.g. WireGuard / Tailscale).  For security, consider adding a reverse-proxy (Caddy or nginx) with HTTPS and basic-auth in front of the Go server before exposing it to the internet.
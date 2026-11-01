# Morris Family Recipe Tracker v4.1

A self-hosted application built in Go for managing a digital recipe library, scraping recipes from the web, and generating professional, auto-categorized PDF cookbooks.  The same Go server powers **both** the browser-based web UI and an optional **Flutter Android app** — the SQLite database lives on the server and is never duplicated.

## Core Features
* **Web-Based UI**: Manage your entire library from any browser (desktop or mobile).
* **Android App** *(experimental)*: A native Flutter app that talks directly to the server's REST API, providing the same functionality on Android without duplicating the database.
* **Smart Web Scraper**: Paste a URL to automatically extract recipe titles, ingredients, and instructions directly from food blogs.
* **Automated Master Cookbook**: Instantly compile your entire library into a single PDF with a generated Table of Contents, automatically grouped and alphabetized by recipe tags (e.g., "Dinner", "Dessert").
* **Dual PDF Modes**: Export individual recipes or the Master Cookbook in **Standard Letter (8.5" x 11")** or **Booklet (5.5" x 8.5")** sizes with perfect text-wrapping.
* **"Daily Bread" Git Sync** *(opt-in)*: Automatically mirrors every recipe as a plain-text `.txt` file plus a `manifest.json` index, committed and pushed to a separate Git repository for future-proof backup.
* **SQLite Backend**: Maintains a robust, searchable database to prevent duplicates and safely handle soft-deletes.

## Usage Instructions

### 1. Running the Server
Run the server from your terminal:

    just run     # or: go run .

The console prints your local IP address. Open a browser at:
* **Desktop:** `http://localhost:8080`
* **Mobile:** `http://<YOUR_LOCAL_IP>:8080`

The server supervises itself: it restarts on a transient crash (e.g. a
port conflict) with exponential backoff, and shuts down cleanly on
Ctrl+C (SIGINT/SIGTERM). A panicking sync task is logged and contained,
never allowed to crash the process. HTTP read/write/idle timeouts guard
against slow clients.

### Configuration

Settings live in `config.json` next to the binary; a default file is
created (owner-only permissions) on first run.

| Field | Default | Meaning |
|---|---|---|
| `port` | `8080` | HTTP listen port (1–65535) |
| `db_file` | `recipes.db` | SQLite database path |
| `backup_path` | `<home>/morris_recipe_vault` | vault directory (required when `git_sync` is true) |
| `git_sync` | `false` | mirror + commit/push the vault to git |
| `git_remote_url` | `""` | optional remote linked on first run |
| `auto_import` | `false` | on startup, import the vault `recipes.json` when the DB is empty |
| `test_mode` | `false` | force sync to a disposable test repo + test vault (TESTMODE) |

Each field can be overridden by an environment variable:

| Field | Env var |
|---|---|
| `port` | `RECIPE_PORT` |
| `db_file` | `RECIPE_DB_FILE` |
| `backup_path` | `RECIPE_BACKUP_PATH` |
| `git_sync` | `RECIPE_GIT_SYNC` (`true`/`false`) |
| `git_remote_url` | `RECIPE_GIT_REMOTE` |
| `auto_import` | `RECIPE_AUTO_IMPORT` (`true`/`false`) |
| `test_mode` | `RECIPE_TEST_MODE` (`true`/`false`) |

Config is validated at startup: a bad port, empty `db_file`, missing
`backup_path` when git sync is on, or whitespace in the remote URL fails
fast with a clear error instead of silently misbehaving.

### 2. Adding Recipes
* **Web Import:** Use the built-in scraper to paste a URL from a recipe site. The app will automatically extract and format the ingredients and steps.
* **Manual Entry:** Use the web interface to type out recipes, tag them, and save them directly to the database.

### 3. Backups, Vault & Restore
Every recipe is mirrored as a human-readable `.txt` file plus a `manifest.json` index (title → file, tags, last-updated). A full JSON snapshot (`recipes.json`) of the active library is also written after every change — this is the fast import path.

* **Git sync on**: the vault lives in `backup_path` — a separate Git repo — and every save or delete is committed and pushed automatically.
* **Git sync off** (default): the vault lives in the local `backups/` folder.

Git sync is **off by default** for new installs. To enable it, set
`"git_sync": true` and a `backup_path` in `config.json`.

To rebuild the SQLite database from the vault after a crash:
```
go run . -restore
```
Or rebuild from the JSON snapshot (recommended):
```
go run . -import /path/to/vault/recipes.json
```
Or restore from a specific folder:
```
go run . -restore -restore-dir /path/to/vault
```
Enable `"auto_import": true` to import the snapshot automatically on startup
when the database is empty.

> **Caveat:** with `auto_import` on, a database you *deliberately* cleared will
> be re-populated from `recipes.json` on the next startup (empty DB looks the
> same as a fresh install). To start truly fresh, delete `recipes.json` too.

Deleting a recipe also removes its vault file (Git history still keeps the old version).

### 4. Printing for Physical Use (The "2-Up" Method)
To print Booklet-sized PDFs (5.5" x 8.5") efficiently onto standard paper to build a physical recipe box or binder:
1. Open the generated PDF and select **Print**.
2. Set Orientation to **Landscape**.
3. Set **Multiple Pages per Sheet** to **2**.
4. Cut the printed 8.5" x 11" sheets in half. The generated margins ensure enough space for standard hole-punching.

## Build Information

### Prerequisites
* [Go](https://go.dev/dl/) 1.26+
* Git (if using the git-sync feature)
* [Elm](https://guide.elm-lang.org/install/elm.html) 0.19 (only to rebuild the web UI)
* [Flutter SDK](https://flutter.dev/docs/get-started/install) 3.x+ (Android app only)

### Building

`just` is the build entry point (see `Justfile`):

    just build    # compile the Go server into morris-tracker.exe
    just test     # run the Go test suite
    just web      # rebuild web/js/elm.js from web/src/Main.elm
    just check    # build + test + web (full verification)
    just run      # run the server on :8080
    just apk      # build the Android release APK
    just release  # zip server exe + APK + runner into recipe-tracker-release.zip

Go dependencies are declared in `go.mod` and fetched automatically — no
manual `go get` needed.

`run_tracker.bat` just launches `morris-tracker.exe` — it no longer rebuilds
from source or checks for updates.

## Security

* `/api/save` and `/api/delete` accept POST only; other methods get `405`.
* Scrape URLs are validated against loopback/private/link-local addresses
  (SSRF guard), and every redirect hop is re-validated.
* There is **no authentication** — by design for a LAN-only home server.
  If you expose it to the internet, put a reverse proxy (Caddy/nginx) with
  HTTPS and auth in front.

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
| Filter by category (tag) | ✅ | ✅ |
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
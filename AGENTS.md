# AGENTS.md

Self-hosted recipe manager. One Go server (stdlib `net/http`) serves the
REST API, the PDF export, and an embedded Elm SPA; a Flutter Android app
in `flutter_app/` is a second client against the same API. Single SQLite
database on the server — clients never hold their own copy.

## File map

| Path | Role |
|---|---|
| `main.go` | entry point, `http.ServeMux` routes, CORS / recovery / logging middleware, scrape URL SSRF guard, supervised restart |
| `config.go` | `config.json` load/save (port, db_file, backup_path, git_sync, git_remote_url, auto_import, test_mode) |
| `database.go` | schema init (+ `PRAGMA user_version`) + recipe CRUD (`modernc.org/sqlite`, soft delete) |
| `vault.go` | plain-text sidecar mirror + `manifest.json` + git add/commit/push worker |
| `restore.go` | rebuild the DB from vault `.txt` files |
| `backup.go` | JSON snapshot (`recipes.json`) with app/schema provenance + import |
| `ingredient.go` | recipe-scaling ingredient parser + scaler + scanner |
| `scraper.go` | web recipe scraper (schema.org JSON-LD → microdata → heuristics) |
| `logging.go` | rotating `tracker.log` + `logInfo`/`logWarn`/`logError` |
| `pdf.go` | single-recipe + master cookbook PDF (`gofpdf`, letter / booklet) |
| `parser.go` | parse a vault `.txt` back into a `Recipe` |
| `models.go` | `Recipe` struct — the shared API shape |
| `web/src/Main.elm` | the whole web UI (Elm 0.19) |
| `web/js/elm.js` | compiled Elm, **committed** because `go:embed` needs it |
| `web/css/app.css` | SPA stylesheet — seasonal themes + design tokens (CSS variables) |
| `VERSION` | release version source of truth (stamped into the exe via `just`) |
| `Justfile` | `build` / `test` / `web` / `release` / `version` recipes |
| `flutter_app/lib/services/api_service.dart` | the Flutter HTTP client |
| `docs/agents/` | agent-stack framework docs (see below) |
| `.github/copilot-instructions.md` | Copilot's *communication channel* rules only — it defers to this file and `CONTEXT.md` for engineering behaviour |

## Commands

```bash
go build ./...        # compile
just build            # compile + stamp VERSION into the exe
go test ./...         # Go test suite (just test)
go run .              # serve on :8080 (reads config.json)
cd web && elm make src/Main.elm --output=js/elm.js   # rebuild the SPA (just web)
just release          # stamp version + zip exe / APK / runner / config
cd flutter_app && flutter test && flutter run
```

## Hard rules

- **`web/js/elm.js` is a build artifact that is committed on purpose.** A
  change to `web/src/Main.elm` is not shipped until `js/elm.js` is rebuilt
  in the same commit.
- **Never hand-edit `web/js/elm.js`.**
- `models.go`'s `Recipe` JSON shape is a cross-language contract: the Elm
  decoder in `web/src/Main.elm` and `flutter_app/lib/models/recipe.dart`
  both break silently if a field is renamed. Change all three together.
- The vault mirror (`vault.go`) is the backup of record. A write path that
  touches `recipes.db` but skips `SyncSidecar` loses data on restore.
- Vault writes go through the sync worker, not direct file writes from
  handlers.
- `config.json`, `recipes.db`, `*.exe`, `*.apk`, `backups/` are gitignored —
  never commit user data or binaries.

## Framework conventions

This repo adopts the [agent-stack](https://github.com/valueforvalue/agent-stack)
framework; `core/` landed under `docs/agents/`. Keep this file short
(per `docs/agents/laws.md` tier-0 size ceiling) and link instead of inlining.

- Session rules: [`docs/agents/session-protocol.md`](docs/agents/session-protocol.md), [`laws.md`](docs/agents/laws.md)
- Commit + branch: [`docs/agents/commit-and-branch.md`](docs/agents/commit-and-branch.md) (single-branch `main`, Conventional Commits)
- Non-trivial work (3+ files): [`docs/agents/rpci.md`](docs/agents/rpci.md)
- Features / slices: [`docs/agents/feature-protocol.md`](docs/agents/feature-protocol.md)
- Tests: [`docs/agents/tdd.md`](docs/agents/tdd.md), [`testing-philosophy.md`](docs/agents/testing-philosophy.md)
- Cross-layer changes: [`docs/agents/code-changes.md`](docs/agents/code-changes.md)
- Bug hunting: [`docs/agents/bug-patterns.md`](docs/agents/bug-patterns.md)
- Durable decisions: [`docs/agents/adr-discipline.md`](docs/agents/adr-discipline.md)
- Domain glossary: [`CONTEXT.md`](CONTEXT.md) — read before schema or copy work
- [`CHANGELOG.md`](CHANGELOG.md) `[Unreleased]` — update in the same commit
- Issues: [`docs/agents/issues/`](docs/agents/issues/)

Tool-specific files are not authoritative: `.github/copilot-instructions.md`
carries conversation mechanics for Copilot and explicitly yields to this file
and `CONTEXT.md`. Add a `CLAUDE.md` / `.cursorrules` the same way if you need
per-tool channel rules — never duplicate engineering rules into them.

**No stack addendum is installed** (the framework ships only `go-htmx`, which
does not match this repo: no templ, no HTMX, no chi, no Wails). Tier-2 for
this repo is `CONTEXT.md` plus the file map above.

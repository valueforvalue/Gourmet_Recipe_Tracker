# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

Commit subjects follow [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/).

## [Unreleased]

### Added
- Supervised server restart: transient listen/serve errors retry with
  exponential backoff, sync-worker panics are contained, and SIGINT/SIGTERM
  trigger graceful shutdown. HTTP read/write/idle timeouts close the
  slowloris gap. (Regression: `TestNewHTTPServerTimeouts`, `TestNextBackoff`,
  `TestShouldRestart`, `TestSleepCtxCancelled`, `TestRecoverTask*`)
- `just apk` (build the Android release APK) and `just release` (zip the
  server exe, APK, runner, and config example — binaries only, no source).
- `AGENTS.md`, `CONTEXT.md` and this changelog — agent-facing repo docs
  (domain glossary, file map, hard rules).
- Web recipe list gains clickable category (tag) filter chips — matching the
  Flutter app — plus a "No recipes found." empty state.
- Web recipe list shows a total recipe count and per-tag counts on the
  filter chips.
- Soft-delete management API: `GET /api/deleted`, `POST /api/restore`,
  `POST /api/purge`, and `POST /api/purge-all`. Purge is permanent and
  requires echoing the confirmation phrase `yesdeary`; restore re-mirrors the
  vault sidecar. (Regression: `TestGetDeletedRecipes`, `TestRestoreRecipe`,
  `TestPurgeRecipe`, `TestPurgeAllDeleted`, `TestPurgeHandlersRequireConfirmation`)
- Web "Deleted" view: checkbox multi-select, restore, and permanent purge
  (purge requires typing `yesdeary` in the confirm box).
- JSON snapshot backup: `recipes.json` (full active library) is written to the
  vault after every change and committed/pushed; `-import <file>` rebuilds the
  DB from it, and `auto_import` (default off) imports on startup when the DB
  is empty. (Regression: `TestWriteSnapshotRoundTrip`, `TestImportFromJSON`,
  `TestIsRecipesEmpty`)
- Recipe scraper rewritten to extract schema.org/Recipe JSON-LD first (title,
  ingredients, instructions, category/cuisine/keyword tags, yield, time,
  description), falling back to microdata (`itemprop`) and then CSS-class
  heuristics. Handles `@graph` nesting and `HowToStep`/`HowToSection`/
  `PropertyValue` shapes. (Regression: `TestScrapeJSONLD*`, `TestScrapeMicrodata`,
  `TestScrapeHeuristic`)
- Scraper enhancements: prep/cook time in notes, cautious ingredient
  normalization (strip bullets + drop exact duplicates), `HowToSection` step
  prefixing, host strategy overrides, per-host politeness delay + one retry on
  transient failures, and a scrape-result empty check in the web UI.
- Recipe scaling (tolerant parser): a Go ingredient parser turns free-text
  lines into (quantity, unit, name), a `-scan` command flags unscalable
  ingredients, and `GET /api/scale` multiplies scalable quantities by a factor.
  The web reader gains ½×/1×/2×/3× controls with a prominent temporary
  "scaled" banner. (Regression: `TestParseIngredient`, `TestScaleIngredient`,
  `TestFormatRat`, `TestScanRecipes`)
- `test_mode` config flag (and `RECIPE_TEST_MODE` env): when true, git sync is
  forced on and redirected to a disposable test repo + `<backup_path>_test`
  vault, so sync tests never touch real backup data. (Regression:
  `TestLoadConfigTestModeRedirectsSync`)
- Seasonal color themes for the web UI: spring / summer / autumn / winter
  palettes applied via CSS custom properties, with Elm setting `data-theme`
  from the current month (northern-hemisphere astronomical seasons).
- Version + schema provenance: a `VERSION` file and Justfile `-ldflags` stamp
  the build version into the exe; `SchemaVersion` is stored in SQLite via
  `PRAGMA user_version`; both are tagged in `recipes.json`
  (`app_version` / `schema_version` / `exported_at`) and in the startup log.
  (Regression: `TestWriteSnapshotRoundTrip`, `TestImportFromJSONSnapshotObject`)
- Web app bar gains a "Feedback" view that posts name / email / message to
  the submit-form endpoint in-app (AJAX), with inline send/error status.
- Flutter app: seasonal color themes (spring / summer / autumn / winter)
  matching the web UI, driven by a `theme/` module + `seasonFromMonth`.
- Server `/api/version` endpoint reports `app_version` + `schema_version`.
  (Regression: `TestVersionEndpoint`)
- Flutter Settings "About" shows the app version (via `package_info_plus`) and
  the server's app + schema version.
- Flutter: "Deleted" tab with restore + permanent purge (single + purge-all
  with `yesdeary` confirmation), backed by the server's soft-delete endpoints.

### Changed
- `git_sync` now defaults to `false` for brand-new config files; an existing
  `config.json` keeps its stored value. Git backup is opt-in.
- Logging: all diagnostics now go through a rotating `tracker.log`
  (10 MB → `tracker.log.1`) with unified `[INFO]`/`[WARN]`/`[ERROR]` levels.
  Previously `fmt.Print*` calls in config/db/vault/restore bypassed the log
  file. (Regression: `TestRotatingLogWriter`)
- `just build` now targets `morris-tracker.exe` (matches the existing
  deployment binary name); `run_tracker.bat` just launches that exe instead
  of rebuilding from source or checking for updates.
- Web UI mobile polish: 44px tap targets, Material-style app bar with
  "Recipe Tracker" title, rounded search field, and reader back/edit actions —
  aligned with the Flutter app's look and feel.
- `test_mode` now defaults to `true` on a fresh install (first-run config and
  the shipped `config.json.example`), so new/release deployments run against
  the disposable test repo until flipped to production.
- `just release` ships a drop-in `config.json` (test-mode default) and
  `README.md` with the production-ready checklist.
- Web UI polish: softer layered shadows, subtle gradients, hover/press motion,
  focus rings, and a warmer serif stack — removes the flat look (CSS-only, no
  markup change).
- Web UI "eye candy" pass: glassmorphism app bar, gradient headings and
  buttons, pill shapes, colored hover glow, and gradient card surfaces
  (CSS-only, theme-aware via CSS variables).
- Seasonal background textures: each theme gets a unique motif — spring
  blossoms (dots), summer sun rays, autumn leaves (diamonds), winter snow
  (crosses) — layered over the aurora glow.

### Fixed
- Config load/save now return real errors, validate port / db_file /
  backup_path / git_remote_url, parse `RECIPE_GIT_SYNC` strictly via
  `strconv.ParseBool`, and write files with `0600` permissions.
  (Regression: `TestConfigValidate`, `TestLoadConfig*`, `TestSaveConfigPermissions`)
- `/api/save` and `/api/delete` now reject non-POST requests with `405`;
  scrape redirects are re-validated so a safe URL cannot bounce to an
  internal address. (Regression: `TestMutatingRoutesRequirePost`,
  `TestCheckScrapeRedirect`)
- Web "daily verse" no longer changes while typing (seeded by recipe count,
  not filter text); search placeholder now matches what the search covers.
- Web "daily verse" now actually rotates by calendar day — seeded by days
  since epoch and refreshed every minute via `elm/time` (was keyed to recipe
  count, so "daily" was a misnomer).
- Web reader notes now preserve line breaks and render URLs (including the
  scraped recipe's source) as clickable links.
- PDF notes now render line-by-line with a `NOTES` header, and scraped
  recipes' source URLs are clickable links in the PDF. (Regression:
  `TestSplitLineLinks`)
- Empty collections (`/api/recipes`, `/api/deleted`) now JSON-encode as `[]`
  instead of `null` when there are no rows. (Regression:
  `TestEmptyCollectionsAreNonNil`)
- Git vault sync now commits before pulling, so `git pull --rebase` runs on a
  clean tree instead of always failing with "You have unstaged changes".
  `git add`/`commit` are serialized against the snapshot write, fixing the
  "short read while indexing recipes.json" race.
- Git vault sync resolves the checked-out branch instead of hardcoding
  `master`, so fresh repos on `main` sync correctly. (Regression:
  `TestGitCurrentBranch`)
- `git_remote_url` is now authoritative: a stale vault `origin` (e.g. a
  deleted repo) is redirected with `git remote set-url`, so config.json
  recovers the push target. (Regression: `TestEnsureRemote`)
- A permanently unreachable remote (deleted repo, bad credentials) now cools
  down network sync for a minute instead of spamming the log on every save;
  local sidecar writes and commits still happen. (Regression:
  `TestIsPermanentPushError`, `TestSyncBackoff`)
- Recipe scraper now sends a browser `User-Agent` (plus `Accept`/
  `Accept-Language`) instead of Go's default `Go-http-client/1.1`, so sites
  that gate on UA — allrecipes returned `402` — serve the page. (Regression:
  `TestNewScrapeRequestSetsBrowserHeaders`)

### Maintenance
- Documented every exported identifier with `Name — contract` doc comments
  (framework `laws.md` compliance; docs-only, no behavior change).
- Adopted the agent-stack framework: `core/` rule docs under `docs/agents/`,
  issue framework under `docs/agents/issues/`, GitHub issue + PR templates,
  `pragmatic-programmer` skill at `.pi/skills/pragmatic-programmer/`.
  No stack addendum installed (the shipped `go-htmx` addendum does not match
  this repo's Go stdlib + Elm + Flutter stack).
- Applied the 6-axis label taxonomy to the GitHub repository.
- Untracked `web/elm-stuff/` build cache (was committed; already gitignored,
  now removed from the index so `elm make` no longer dirties the tree).
- Added a `Justfile` for build tasks (`build`, `test`, `web`, `run`,
  `flutter-test`, `flutter-run`, `check`).
- Refreshed `README.md` (config, resilience, security, `just` build flow)
  and `config.json.example` to match this session's changes.
- Ignore SQLite WAL/SHM runtime files (`*.db-shm`, `*.db-wal`) and the
  `test_empty_recipes.db` test artifact so `git status` stays clean.

<!--
  Conventions:
  - One bullet per logical change
  - Reference the issue number if one exists
  - Reference the regression net (test name, smoke probe)
  - Internal refactors that don't change user-visible behaviour go under Maintenance
-->

## [4.1] - 2026-04-09

System overhaul. Recorded retroactively from git history; earlier releases
are summarised in [`README.md`](README.md) rather than itemised here.

### Fixed
- Register the `-restore` flag so it coexists with `-version`.

## [4.0]

- Flutter Android client as a second API consumer, CORS-enabled API,
  configurable server URL on the device.

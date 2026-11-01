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

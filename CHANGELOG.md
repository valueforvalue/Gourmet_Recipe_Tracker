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
- `AGENTS.md`, `CONTEXT.md` and this changelog — agent-facing repo docs
  (domain glossary, file map, hard rules).

### Maintenance
- Adopted the agent-stack framework: `core/` rule docs under `docs/agents/`,
  issue framework under `docs/agents/issues/`, GitHub issue + PR templates,
  `pragmatic-programmer` skill at `.pi/skills/pragmatic-programmer/`.
  No stack addendum installed (the shipped `go-htmx` addendum does not match
  this repo's Go stdlib + Elm + Flutter stack).
- Applied the 6-axis label taxonomy to the GitHub repository.

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

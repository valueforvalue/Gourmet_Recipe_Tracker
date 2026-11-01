# Framework Bootstrap — Session-Start Checklist

This repo adopted the [agent-stack](https://github.com/valueforvalue/agent-stack)
framework on 2026-11-01. Read these docs at session start.

## Tier-0 — always loaded

These cross-cut every task. Load them once.

- [ ] [`docs/agents/session-protocol.md`](docs/agents/session-protocol.md) — agent session rules
- [ ] [`docs/agents/laws.md`](docs/agents/laws.md) — non-negotiable laws
- [ ] [`docs/agents/commit-and-branch.md`](docs/agents/commit-and-branch.md) — commit shape + branch policy
- [ ] [`docs/agents/docs-index-scheme.md`](docs/agents/docs-index-scheme.md) — 3-tier progressive disclosure
- [ ] [`docs/agents/glossary-discipline.md`](docs/agents/glossary-discipline.md) — glossary maintenance
- [ ] [`AGENTS.md`](AGENTS.md) — project file map, commands, hard rules
- [ ] [`CONTEXT.md`](CONTEXT.md) — domain glossary + project laws

## Tier-1 — load by task role

Load only the ones that match your task.

- [ ] [`docs/agents/rpci.md`](docs/agents/rpci.md) — non-trivial work (3+ files or design questions)
- [ ] [`docs/agents/tdd.md`](docs/agents/tdd.md) — any feature or bug fix
- [ ] [`docs/agents/feature-protocol.md`](docs/agents/feature-protocol.md) — adding a feature
- [ ] [`docs/agents/code-changes.md`](docs/agents/code-changes.md) — cross-layer change (Go ⇄ Elm ⇄ Dart)
- [ ] [`docs/agents/bug-patterns.md`](docs/agents/bug-patterns.md) — hunting a bug or reviewing a fix
- [ ] [`docs/agents/pragmatic-principles.md`](docs/agents/pragmatic-principles.md) — principle review during Plan
- [ ] [`docs/agents/adr-discipline.md`](docs/agents/adr-discipline.md) — recording a durable decision
- [ ] [`docs/agents/complexity.md`](docs/agents/complexity.md) — net-complexity test before an abstraction
- [ ] [`docs/agents/testing-philosophy.md`](docs/agents/testing-philosophy.md) — what "tested" means here

## Tier-2 — on demand

**No stack addendum is installed.** The framework ships only `go-htmx.md`
(templ + HTMX + chi + Wails), none of which this repo uses — this is a Go
stdlib `net/http` server, an Elm SPA, and a Flutter client. Load nothing from
`docs/agents/addenda/` (it does not exist here).

Tier-2 for this repo is instead:

- `CONTEXT.md` — domain vocabulary and project laws
- `AGENTS.md` — file map, build/test commands, cross-language contract rules
- The source itself: `models.go` (shared `Recipe` shape), `vault.go` (backup of
  record), `web/src/Main.elm` (web client), `flutter_app/lib/services/api_service.dart`
  (Android client)

If a Go + Elm + Flutter addendum is written later, it belongs at
`docs/agents/addenda/go-elm-flutter.md` and this section is updated to point at it.

## Issue framework

Adopted. See [`docs/agents/issues/`](docs/agents/issues/) for the label
taxonomy, bug + feature templates, and triage workflow. GitHub-side templates
live in [`.github/ISSUE_TEMPLATE/`](.github/ISSUE_TEMPLATE/); the 6-axis label
set was applied to the repository.

## Bundled skills

Selected: `pragmatic-programmer` → [`.pi/skills/pragmatic-programmer/`](.pi/skills/pragmatic-programmer/),
exposed as `/skill:pragmatic-programmer`.

## Branch model

`single-branch`. Work lands on `main` via Conventional Commits; no develop /
release lanes.

## Conventions in this repo

- Commit subjects follow Conventional Commits (`feat:`, `fix:`, `docs:`, …)
- `CHANGELOG.md` `[Unreleased]` block updated per slice, same commit as the change
- Glossary (`CONTEXT.md`) updated whenever a new domain term is introduced
- Issue tracker uses the 6-axis label taxonomy (see `docs/agents/issues/label-taxonomy.md`)
- Three-client contract: a `Recipe` field change touches `models.go`, the Elm
  decoder, and `flutter_app/lib/models/recipe.dart` in one commit
- `web/js/elm.js` is regenerated from `web/src/Main.elm`, never hand-edited

# Morris Family Recipe Tracker

A self-hosted recipe library that runs as one Go binary on a home server.
It scrapes recipes from the web, stores them in SQLite, mirrors them as
plain-text files into a git-backed vault, and renders PDF cookbooks. Two
clients talk to it: an Elm SPA served by the Go binary itself, and a
Flutter Android app that points at the server over the LAN.

## Language

**Recipe**: one dish record — title, tags, ingredients, instructions.
The unit the whole app is organised around.
_Avoid_: dish, entry, item, card

**Title**: the recipe's identity key. Unique in the DB and used as the
filename stem in the vault and as the `?title=` query param on the API.
_Avoid_: name, slug, id

**Tag**: a free-text category on a recipe (`Dinner`, `Dessert`, …). Drives
the automatic grouping and alphabetisation in the master cookbook.
_Avoid_: category, label, collection

**Soft delete**: a recipe is marked deleted, not removed. Lets a scrape or
a restore re-expose it rather than silently duplicating it.
_Avoid_: archive, trash, hide

**Vault**: the `backup_path` directory (default `morris_recipe_vault`)
holding one `.txt` per recipe plus `manifest.json`. Committed and pushed to
a separate git repo. This — not `recipes.db` — is the durable store.
_Avoid_: backup folder, mirror dir, daily bread (historical name, dropped)

**Sidecar**: the single `.txt` file inside the vault that represents one
recipe. Human-readable on purpose.
_Avoid_: mirror file, dump

**Manifest**: `manifest.json` inside the vault — title → filename, tags,
last-updated. The index that makes restore possible without re-parsing
every file blindly.
_Avoid_: index, catalog

**Sync worker**: the single background goroutine that turns vault writes
into `git add` / `commit` / `push`. All vault mutation funnels through it.
_Avoid_: git hook, pusher, background job

**Restore**: rebuild `recipes.db` from vault sidecars (`-restore` flag).
_Avoid_: import, rebuild, reseed

**Scrape**: fetch a recipe URL and extract title / ingredients / steps via
`goquery`. Bounded in time and restricted to safe (non-private) hosts.
_Avoid_: import from web, crawl, fetch

**Cookbook**: the multi-recipe PDF, auto-grouped and alphabetised by tag.
"Letter" and "Booklet" are page sizes (8.5×11 and 5.5×8.5), not content modes.
_Avoid_: compilation, book, collection

**Client**: either the Elm SPA or the Flutter app. Both are stateless views
over the server API; neither owns data.
_Avoid_: frontend, app (ambiguous — always say which client)

## Relationships

- A **recipe** has exactly one **title**, zero or more **tags**.
- A **recipe** has exactly one **sidecar** in the **vault**, indexed by
  exactly one **manifest** entry.
- A **cookbook** contains zero or more **recipes**, grouped by **tag**.
- A **soft delete** removes the sidecar and manifest entry but keeps the row.
- A **restore** recreates rows from sidecars; the **manifest** decides which
  files are recipes.

## Example dialogue

> **Dev:** "I renamed the title field in `models.go` and the Go tests pass."
> **Maintainer:** "Tests don't cover the other two clients. The Elm decoder
> and `recipe.dart` both read `title`. Rename all three, rebuild `elm.js`,
> then check a **restore** still matches **manifest** entries by **title**."

## Historical

- **"Daily Bread"** — the git-sync feature's marketing name in the README.
  The domain term is **vault**. The README keeps the name; code and docs
  do not.
- **`elm.js` as a vendored asset** — it looks like a third-party library
  but is compiled from `web/src/Main.elm` here. It stays committed only
  because `go:embed` needs it at build time.
- **`morrisrecipetracker.apk` / `recipe-tracker.exe`** — committed binaries
  from before they were gitignored. Present in history, not current output.

## Adding features

See [`docs/agents/feature-protocol.md`](docs/agents/feature-protocol.md).

## Laws (non-negotiable)

1. **Vault before database.** Any write path that changes `recipes.db` must
   also enqueue the sidecar sync. A row without a sidecar is not saved.
   (Origin: vault.go sync worker design.)
2. **Three-way `Recipe` contract.** Changing a field in `models.go` requires
   the matching Elm decoder and Dart model in the same change.
3. **`web/js/elm.js` is rebuilt, never edited.** Ship it in the same commit
   as the `.elm` source change.
4. **Scrape targets are validated.** `isSafeURL` / `isSafeIP` guard SSRF;
   never add a fetch that bypasses them.

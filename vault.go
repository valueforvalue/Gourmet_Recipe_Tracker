// File: vault.go
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// --- Vault (plain-text backup + manifest) ---
//
// Every recipe is mirrored as a human-readable .txt file, and the whole set is
// indexed by a manifest.json.  With GitSync enabled the vault lives inside the
// configured Git repository (BackupPath); otherwise it lives in the local
// backups/ folder.  The vault is the single source of truth for disaster
// recovery: `-restore` rebuilds the SQLite database from it.

// SyncTask is a unit of vault work queued for the sync worker. Op is either
// "save" or "delete".
type SyncTask struct {
	Op     string // "save" or "delete"
	Title  string
	Recipe Recipe
}

var syncQueue = make(chan SyncTask, 100)

// ManifestEntry indexes one recipe in the vault manifest.
type ManifestEntry struct {
	Title     string   `json:"title"`
	File      string   `json:"file"`
	Tags      []string `json:"tags"`
	UpdatedAt string   `json:"updated_at"`
}

// Manifest is the vault index: title -> filename, tags, last-updated. It is
// what lets -restore rebuild the database without blind re-parsing.
type Manifest struct {
	Version   int             `json:"version"`
	UpdatedAt string          `json:"updated_at"`
	Recipes   []ManifestEntry `json:"recipes"`
}

const manifestFileName = "manifest.json"

// manifestMu serializes read-modify-write access to manifest.json.  The git
// worker already processes tasks one at a time, but local-only mode writes
// from request-handler goroutines, so the lock is still required.
var manifestMu sync.Mutex

// vaultDir returns where backups live for the current configuration.
func vaultDir() string {
	if GlobalConfig.GitSync {
		return GlobalConfig.BackupPath
	}
	return "backups"
}

// buildSidecarContent renders a recipe as the round-trippable text format that
// ParseRecipeFile can read back.
func buildSidecarContent(r Recipe) string {
	return fmt.Sprintf("RECIPE: %s\nTAGS: %s\n\nINGREDIENTS\n- %s\n\nINSTRUCTIONS\n%s\n\nNOTES: %s",
		r.Title,
		strings.Join(r.Tags, ", "),
		strings.Join(r.Ingredients, "\n- "),
		strings.Join(r.Instructions, "\n"),
		r.Notes)
}

// SyncSidecar enqueues (Git mode) or writes (local mode) the vault mirror.
func SyncSidecar(r Recipe) error {
	if GlobalConfig.GitSync {
		select {
		case syncQueue <- SyncTask{Op: "save", Title: r.Title, Recipe: r}:
			return nil
		default:
			return fmt.Errorf("sync queue full, backup for %q skipped", r.Title)
		}
	}
	_, err := writeVaultEntry(vaultDir(), r)
	return err
}

// SyncVaultDelete removes a recipe's vault file.  In Git mode it is queued so
// the removal is committed alongside the next worker run.
func SyncVaultDelete(title string) {
	if GlobalConfig.GitSync {
		select {
		case syncQueue <- SyncTask{Op: "delete", Title: title}:
			return
		default:
			logError("sync: queue full, vault delete for %q skipped", title)
			return
		}
	}
	if _, err := removeVaultEntry(vaultDir(), title); err != nil {
		logError("sync: %v", err)
	}
}

// writeVaultEntry writes the .txt file and updates the manifest.  It returns
// the filename written so callers can reference it in a commit message.
func writeVaultEntry(dir string, r Recipe) (string, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}

	fileName := sanitizeFilename(r.Title) + ".txt"
	content := buildSidecarContent(r)
	if err := os.WriteFile(filepath.Join(dir, fileName), []byte(content), 0644); err != nil {
		return "", err
	}

	manifestMu.Lock()
	defer manifestMu.Unlock()
	m := loadManifestLocked(dir)
	upsertManifestEntry(&m, ManifestEntry{
		Title:     r.Title,
		File:      fileName,
		Tags:      r.Tags,
		UpdatedAt: time.Now().Format(time.RFC3339),
	})
	return fileName, saveManifest(dir, m)
}

// removeVaultEntry deletes the .txt file and its manifest entry.
func removeVaultEntry(dir string, title string) (string, error) {
	fileName := sanitizeFilename(title) + ".txt"
	_ = os.Remove(filepath.Join(dir, fileName))

	manifestMu.Lock()
	defer manifestMu.Unlock()
	m := loadManifestLocked(dir)
	m.Recipes = removeManifestEntry(m.Recipes, title)
	return fileName, saveManifest(dir, m)
}

func loadManifest(dir string) Manifest {
	manifestMu.Lock()
	defer manifestMu.Unlock()
	return loadManifestLocked(dir)
}

func loadManifestLocked(dir string) Manifest {
	m := Manifest{Version: 1, Recipes: []ManifestEntry{}}
	data, err := os.ReadFile(filepath.Join(dir, manifestFileName))
	if err != nil {
		return m
	}
	_ = json.Unmarshal(data, &m)
	if m.Recipes == nil {
		m.Recipes = []ManifestEntry{}
	}
	return m
}

func saveManifest(dir string, m Manifest) error {
	m.Version = 1
	m.UpdatedAt = time.Now().Format(time.RFC3339)
	sort.Slice(m.Recipes, func(i, j int) bool { return m.Recipes[i].Title < m.Recipes[j].Title })
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, manifestFileName), data, 0644)
}

func upsertManifestEntry(m *Manifest, e ManifestEntry) {
	for i := range m.Recipes {
		if m.Recipes[i].Title == e.Title {
			m.Recipes[i] = e
			return
		}
	}
	m.Recipes = append(m.Recipes, e)
}

func removeManifestEntry(entries []ManifestEntry, title string) []ManifestEntry {
	out := entries[:0]
	for _, e := range entries {
		if e.Title != title {
			out = append(out, e)
		}
	}
	return out
}

// StartSyncWorker processes vault sync tasks one at a time, committing and
// pushing each change to the configured Git remote. A panicking task is
// contained so it cannot crash the worker (or the process).
func StartSyncWorker() {
	for task := range syncQueue {
		if err := processSyncTaskGuarded(task); err != nil {
			logError("sync: %v", err)
		}
	}
}

// processSyncTaskGuarded runs one sync task, converting any panic into an
// error so a single bad task cannot take down the worker.
func processSyncTaskGuarded(task SyncTask) error {
	return recoverTask(fmt.Sprintf("sync %s %q", task.Op, task.Title), func() error {
		return processSyncTask(task)
	})
}

// recoverTask runs fn and converts any panic into a returned error. It lets
// the sync worker survive a panicking task instead of crashing the process.
func recoverTask(name string, fn func() error) (err error) {
	defer func() {
		if rec := recover(); rec != nil {
			err = fmt.Errorf("%s panicked: %v", name, rec)
		}
	}()
	return fn()
}

// syncBackoffDuration is how long network sync (pull/push) is skipped after a
// permanent remote failure (e.g. the remote repository was deleted). It is a
// var so tests can shrink it.
var syncBackoffDuration = time.Minute

// syncBackoffMu guards syncBackoffUntil.
var syncBackoffMu sync.Mutex

// syncBackoffUntil is the earliest time network sync should be retried.
var syncBackoffUntil time.Time

// syncNetworkReady reports whether a pull/push should be attempted now. It is
// false during the cooldown after a permanent remote failure.
func syncNetworkReady() bool {
	syncBackoffMu.Lock()
	defer syncBackoffMu.Unlock()
	return !time.Now().Before(syncBackoffUntil)
}

// markSyncBackoff starts a cooldown so the worker stops hammering a remote
// that is permanently unreachable (missing repo, bad credentials). Local
// sidecar writes and commits still happen; only the network hop is deferred.
func markSyncBackoff() {
	syncBackoffMu.Lock()
	defer syncBackoffMu.Unlock()
	syncBackoffUntil = time.Now().Add(syncBackoffDuration)
}

// isPermanentPushError reports whether a push error will not resolve on its
// own (missing/deleted remote, bad credentials, not a git repository), as
// opposed to a transient network hiccup worth retrying.
func isPermanentPushError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	for _, marker := range []string{
		"Repository not found",
		"fatal: repository",
		"does not appear to be a git repository",
		"could not read from remote repository",
		"remote: Not Found",
		"Authentication failed",
		"access denied",
		"could not read Username",
	} {
		if strings.Contains(msg, marker) {
			return true
		}
	}
	return false
}

func processSyncTask(task SyncTask) error {
	dir := vaultDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	if err := ensureGitIdentity(dir); err != nil {
		return err
	}
	branch, err := gitCurrentBranch(dir)
	if err != nil {
		return err
	}

	switch task.Op {
	case "save":
		if _, err := writeVaultEntry(dir, task.Recipe); err != nil {
			return err
		}
	case "delete":
		if _, err := removeVaultEntry(dir, task.Title); err != nil {
			return err
		}
	case "commit":
		// The snapshot was already written synchronously by SyncSnapshot;
		// this task only commits and pushes it.
	default:
		return fmt.Errorf("unknown sync op %q", task.Op)
	}

	// Commit locally first: SyncSnapshot already dirtied the tree with
	// recipes.json, and this task's sidecar write adds to it. Committing before
	// pulling leaves a clean tree so `git pull --rebase` can run (it refuses on
	// unstaged changes).
	if err := gitCommit(dir, task.Op, task.Title); err != nil {
		return err
	}

	if !syncNetworkReady() {
		logWarn("sync: push skipped (remote unavailable, cooling down)")
		return nil
	}
	if err := gitPull(dir, branch); err != nil {
		logWarn("sync: pull: %v", err)
	}
	if err := gitPush(dir, branch); err != nil {
		if isPermanentPushError(err) {
			markSyncBackoff()
		}
		return err
	}
	logInfo("git: %s vaulted", task.Title)
	return nil
}

// ensureGitIdentity sets a local identity in the vault repo when neither a
// local nor a global git identity exists, otherwise commits fail.
func ensureGitIdentity(dir string) error {
	if gitConfigValue(dir, "user.email") != "" {
		return nil
	}
	if out, _ := exec.Command("git", "config", "--global", "user.email").Output(); strings.TrimSpace(string(out)) != "" {
		return nil
	}
	exec.Command("git", "-C", dir, "config", "user.email", "recipe-tracker@localhost").Run()
	exec.Command("git", "-C", dir, "config", "user.name", "Recipe Tracker").Run()
	return nil
}

func gitConfigValue(dir, key string) string {
	out, err := exec.Command("git", "-C", dir, "config", key).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// gitCurrentBranch returns the vault repo's checked-out branch name (e.g.
// "master" or "main"). Sync must never hardcode a branch: a fresh `git init`
// can land on either name depending on git version and user config. It uses
// symbolic-ref (not rev-parse) so it also works on an unborn HEAD before the
// first commit.
func gitCurrentBranch(dir string) (string, error) {
	out, err := exec.Command("git", "-C", dir, "symbolic-ref", "--short", "HEAD").CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("resolve branch: %s", strings.TrimSpace(string(out)))
	}
	branch := strings.TrimSpace(string(out))
	if branch == "" {
		return "", fmt.Errorf("resolve branch: empty")
	}
	return branch, nil
}

// ensureRemote points the repo's origin at url, adding it when absent and
// replacing it when it points elsewhere. It makes config.json's git_remote_url
// the single source of truth for where backups are pushed.
func ensureRemote(dir, url string) error {
	out, _ := exec.Command("git", "-C", dir, "remote", "get-url", "origin").Output()
	current := strings.TrimSpace(string(out))
	if current == url {
		return nil
	}
	action := "set-url"
	if current == "" {
		action = "add"
	}
	if out, err := exec.Command("git", "-C", dir, "remote", action, "origin", url).CombinedOutput(); err != nil {
		return fmt.Errorf("%s: %s", action, strings.TrimSpace(string(out)))
	}
	return nil
}

func gitPull(dir, branch string) error {
	out, err := exec.Command("git", "-C", dir, "pull", "--rebase", "origin", branch).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return nil
}

func gitCommit(dir, op, title string) error {
	// SyncSnapshot writes recipes.json (atomic tmp+rename) from request
	// handlers. `git add -A` racing that rename produces "short read while
	// indexing recipes.json". Holding snapshotMu serializes indexing against
	// snapshot writes. The lock covers only the local add+commit, never the
	// network pull/push, so a slow remote cannot stall snapshot writes.
	snapshotMu.Lock()
	defer snapshotMu.Unlock()

	if out, err := exec.Command("git", "-C", dir, "add", "-A").CombinedOutput(); err != nil {
		return fmt.Errorf("add: %s", strings.TrimSpace(string(out)))
	}

	verb := "Update"
	if op == "delete" {
		verb = "Remove"
	}
	msg := fmt.Sprintf("%s %s", verb, title)
	if out, err := exec.Command("git", "-C", dir, "commit", "-m", msg).CombinedOutput(); err != nil {
		// A commit with nothing staged (e.g. identical re-save) is not fatal.
		if !strings.Contains(string(out), "nothing to commit") {
			return fmt.Errorf("commit: %s", strings.TrimSpace(string(out)))
		}
	}
	return nil
}

func gitPush(dir, branch string) error {
	if out, err := exec.Command("git", "-C", dir, "push", "origin", branch).CombinedOutput(); err != nil {
		return fmt.Errorf("push: %s", strings.TrimSpace(string(out)))
	}
	return nil
}

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

type SyncTask struct {
	Op     string // "save" or "delete"
	Title  string
	Recipe Recipe
}

var syncQueue = make(chan SyncTask, 100)

type ManifestEntry struct {
	Title     string   `json:"title"`
	File      string   `json:"file"`
	Tags      []string `json:"tags"`
	UpdatedAt string   `json:"updated_at"`
}

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
			fmt.Printf(" [Sync Error]: queue full, vault delete for %q skipped\n", title)
			return
		}
	}
	if _, err := removeVaultEntry(vaultDir(), title); err != nil {
		fmt.Printf(" [Sync Error]: %v\n", err)
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
// pushing each change to the configured Git remote.
func StartSyncWorker() {
	for task := range syncQueue {
		if err := processSyncTask(task); err != nil {
			fmt.Printf(" [Sync Error]: %v\n", err)
		}
	}
}

func processSyncTask(task SyncTask) error {
	dir := vaultDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	if err := ensureGitIdentity(dir); err != nil {
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
	default:
		return fmt.Errorf("unknown sync op %q", task.Op)
	}

	// Rebase onto the remote before pushing so a remote that moved ahead does
	// not cause a hard push failure.
	if err := gitPull(dir); err != nil {
		fmt.Printf(" [Sync Warning]: pull: %v\n", err)
	}
	return gitAddCommitPush(dir, task.Op, task.Title)
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

func gitPull(dir string) error {
	out, err := exec.Command("git", "-C", dir, "pull", "--rebase", "origin", "master").CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return nil
}

func gitAddCommitPush(dir, op, title string) error {
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

	if out, err := exec.Command("git", "-C", dir, "push", "origin", "master").CombinedOutput(); err != nil {
		return fmt.Errorf("push: %s", strings.TrimSpace(string(out)))
	}
	fmt.Printf(" [Git Success]: %s vaulted to GitHub.\n", title)
	return nil
}

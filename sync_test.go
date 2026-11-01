package main

import (
	"errors"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestGitCurrentBranch(t *testing.T) {
	dir := t.TempDir()
	if out, err := exec.Command("git", "-C", dir, "init", "-q").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v (%s)", err, out)
	}

	// Compare against git's own report rather than hardcoding master/main, so
	// the test stays correct regardless of the user's init.defaultBranch.
	wantOut, err := exec.Command("git", "-C", dir, "symbolic-ref", "--short", "HEAD").Output()
	if err != nil {
		t.Fatalf("git symbolic-ref: %v", err)
	}
	want := strings.TrimSpace(string(wantOut))

	got, err := gitCurrentBranch(dir)
	if err != nil {
		t.Fatalf("gitCurrentBranch: %v", err)
	}
	if got != want {
		t.Errorf("gitCurrentBranch = %q, want %q", got, want)
	}
}

func TestEnsureRemote(t *testing.T) {
	dir := t.TempDir()
	if out, err := exec.Command("git", "-C", dir, "init", "-q").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v (%s)", err, out)
	}

	// Add when missing.
	if err := ensureRemote(dir, "https://example.com/a.git"); err != nil {
		t.Fatalf("ensureRemote add: %v", err)
	}
	if got := remoteURL(t, dir); got != "https://example.com/a.git" {
		t.Errorf("after add, origin = %q", got)
	}

	// Replace when it points elsewhere (stale/deleted remote).
	if err := ensureRemote(dir, "https://example.com/b.git"); err != nil {
		t.Fatalf("ensureRemote set-url: %v", err)
	}
	if got := remoteURL(t, dir); got != "https://example.com/b.git" {
		t.Errorf("after set-url, origin = %q", got)
	}

	// No-op when already correct.
	if err := ensureRemote(dir, "https://example.com/b.git"); err != nil {
		t.Fatalf("ensureRemote no-op: %v", err)
	}
}

func remoteURL(t *testing.T, dir string) string {
	t.Helper()
	out, err := exec.Command("git", "-C", dir, "remote", "get-url", "origin").Output()
	if err != nil {
		t.Fatalf("git remote get-url: %v", err)
	}
	return strings.TrimSpace(string(out))
}

func TestIsPermanentPushError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"deleted repo", errors.New("fatal: repository 'https://github.com/x/y.git/' not found"), true},
		{"not a git repo", errors.New("does not appear to be a git repository"), true},
		{"auth failed", errors.New("remote: Invalid username or password\nfatal: Authentication failed"), true},
		{"transient network", errors.New("ssh: connect to host github.com port 22: Connection refused"), false},
		{"unknown", errors.New("something else"), false},
	}
	for _, tt := range tests {
		if got := isPermanentPushError(tt.err); got != tt.want {
			t.Errorf("%s: isPermanentPushError() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestSyncBackoff(t *testing.T) {
	origDur := syncBackoffDuration
	defer func() {
		syncBackoffDuration = origDur
		syncBackoffUntil = time.Time{}
	}()

	syncBackoffDuration = time.Hour
	markSyncBackoff()
	if syncNetworkReady() {
		t.Error("expected network sync to be cooled down after a permanent failure")
	}

	syncBackoffDuration = 0
	markSyncBackoff()
	if !syncNetworkReady() {
		t.Error("expected network sync ready after zero cooldown")
	}
}

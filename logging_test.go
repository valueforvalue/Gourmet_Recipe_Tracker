package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRotatingLogWriter(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.log")

	w := &rotatingLogWriter{path: path, maxBytes: 10}
	defer w.Close()

	// 5 bytes, under limit: no rotation.
	if _, err := w.Write([]byte("12345")); err != nil {
		t.Fatalf("write 1: %v", err)
	}
	// 5 more, exactly at limit (10): still no rotation.
	if _, err := w.Write([]byte("67890")); err != nil {
		t.Fatalf("write 2: %v", err)
	}
	if _, err := os.Stat(path + ".1"); !os.IsNotExist(err) {
		t.Fatalf("unexpected .1 backup before exceeding limit (err=%v)", err)
	}

	// 1 more byte pushes past the limit: rotate.
	if _, err := w.Write([]byte("x")); err != nil {
		t.Fatalf("write 3: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil || string(got) != "x" {
		t.Errorf("current log = %q, want %q (err=%v)", got, "x", err)
	}
	backup, err := os.ReadFile(path + ".1")
	if err != nil || string(backup) != "1234567890" {
		t.Errorf("backup = %q, want %q (err=%v)", backup, "1234567890", err)
	}
}

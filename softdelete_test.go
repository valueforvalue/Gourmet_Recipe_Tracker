package main

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func newSoftDeleteTestDB(t *testing.T) (db *sql.DB) {
	t.Helper()
	db, err := InitDB(filepath.Join(t.TempDir(), "softdelete.db"))
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestGetDeletedRecipes(t *testing.T) {
	GlobalConfig.GitSync = false
	db := newSoftDeleteTestDB(t)

	active := Recipe{Title: "Active Recipe"}
	deleted := Recipe{Title: "Deleted Recipe"}
	for _, r := range []Recipe{active, deleted} {
		if err := saveRecipeDB(db, r); err != nil {
			t.Fatalf("saveRecipeDB(%s): %v", r.Title, err)
		}
	}
	if err := DeleteRecipe(db, "Deleted Recipe"); err != nil {
		t.Fatalf("DeleteRecipe: %v", err)
	}

	got, err := GetDeletedRecipes(db)
	if err != nil {
		t.Fatalf("GetDeletedRecipes: %v", err)
	}
	if len(got) != 1 || got[0].Title != "Deleted Recipe" {
		t.Errorf("GetDeletedRecipes = %+v, want only Deleted Recipe", got)
	}
}

func TestRestoreRecipe(t *testing.T) {
	GlobalConfig.GitSync = false
	db := newSoftDeleteTestDB(t)

	title := "Restore Me"
	if err := saveRecipeDB(db, Recipe{Title: title, Ingredients: []string{"flour"}}); err != nil {
		t.Fatal(err)
	}
	if err := DeleteRecipe(db, title); err != nil {
		t.Fatal(err)
	}
	if err := RestoreRecipe(db, title); err != nil {
		t.Fatalf("RestoreRecipe: %v", err)
	}

	got, err := GetRecipeByTitle(db, title)
	if err != nil || got.Title != title {
		t.Errorf("restored recipe not visible: %+v err=%v", got, err)
	}

	// Restore re-mirrors the vault sidecar.
	sidecar := filepath.Join(vaultDir(), sanitizeFilename(title)+".txt")
	if _, err := os.Stat(sidecar); err != nil {
		t.Errorf("expected sidecar re-mirrored at %s, got %v", sidecar, err)
	}
	removeVaultEntry(vaultDir(), title)
}

func TestRestoreMissingRecipe(t *testing.T) {
	db := newSoftDeleteTestDB(t)
	if err := RestoreRecipe(db, "Does Not Exist"); err == nil {
		t.Error("RestoreRecipe should fail for a missing title")
	}
}

func TestPurgeRecipe(t *testing.T) {
	GlobalConfig.GitSync = false
	db := newSoftDeleteTestDB(t)

	title := "Purge Me"
	if err := saveRecipeDB(db, Recipe{Title: title}); err != nil {
		t.Fatal(err)
	}
	if err := DeleteRecipe(db, title); err != nil {
		t.Fatal(err)
	}
	if err := PurgeRecipe(db, title); err != nil {
		t.Fatalf("PurgeRecipe: %v", err)
	}

	got, err := GetDeletedRecipes(db)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Errorf("expected no deleted recipes after purge, got %+v", got)
	}
	// Row is gone entirely — restoring should now fail.
	if err := RestoreRecipe(db, title); err == nil {
		t.Error("restore after purge should fail (row removed)")
	}
}

func TestPurgeAllDeleted(t *testing.T) {
	GlobalConfig.GitSync = false
	db := newSoftDeleteTestDB(t)

	for _, title := range []string{"A", "B", "Keep"} {
		if err := saveRecipeDB(db, Recipe{Title: title}); err != nil {
			t.Fatal(err)
		}
	}
	for _, title := range []string{"A", "B"} {
		if err := DeleteRecipe(db, title); err != nil {
			t.Fatal(err)
		}
	}

	if err := PurgeAllDeleted(db); err != nil {
		t.Fatalf("PurgeAllDeleted: %v", err)
	}

	deleted, _ := GetDeletedRecipes(db)
	if len(deleted) != 0 {
		t.Errorf("expected no deleted recipes, got %+v", deleted)
	}
	if _, err := GetRecipeByTitle(db, "Keep"); err != nil {
		t.Errorf("active recipe should survive purge-all: %v", err)
	}
}

func TestPurgeHandlersRequireConfirmation(t *testing.T) {
	GlobalConfig.GitSync = false
	db := newSoftDeleteTestDB(t)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/purge", handlePurge(db))
	mux.HandleFunc("/api/purge-all", handlePurgeAll(db))
	mux.HandleFunc("/api/restore", handleRestore(db))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	// Purge without the confirmation word -> 400.
	for _, path := range []string{"/api/purge?title=x", "/api/purge?title=x&confirm=nope", "/api/purge-all"} {
		resp, err := http.Post(srv.URL+path, "application/json", nil)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("POST %s = %d, want 400", path, resp.StatusCode)
		}
		resp.Body.Close()
	}

	// Mutating endpoints reject GET.
	for _, path := range []string{"/api/purge", "/api/purge-all", "/api/restore"} {
		resp, err := http.Get(srv.URL + path)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Errorf("GET %s = %d, want 405", path, resp.StatusCode)
		}
		resp.Body.Close()
	}
}

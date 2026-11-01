package main

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func TestMutatingRoutesRequirePost(t *testing.T) {
	GlobalConfig.GitSync = false

	db, err := InitDB(filepath.Join(t.TempDir(), "routes.db"))
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	defer db.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/save", handleSaveRecipe(db))
	mux.HandleFunc("/api/delete", handleDeleteRecipe(db))

	srv := httptest.NewServer(mux)
	defer srv.Close()

	for _, path := range []string{"/api/save", "/api/delete?title=x"} {
		resp, err := http.Get(srv.URL + path)
		if err != nil {
			t.Fatalf("GET %s: %v", path, err)
		}
		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Errorf("GET %s = %d, want %d", path, resp.StatusCode, http.StatusMethodNotAllowed)
		}
		resp.Body.Close()
	}
}

func TestCheckScrapeRedirect(t *testing.T) {
	// Redirect to a loopback address must be blocked.
	loopback, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1/internal", nil)
	if err := checkScrapeRedirect(loopback, nil); err == nil {
		t.Error("redirect to 127.0.0.1 was allowed, want blocked")
	}

	// Redirect to a public address is allowed.
	public, _ := http.NewRequest(http.MethodGet, "https://8.8.8.8/page", nil)
	if err := checkScrapeRedirect(public, nil); err != nil {
		t.Errorf("redirect to public IP blocked: %v", err)
	}

	// Redirect chain past the limit must be blocked.
	via := make([]*http.Request, 10)
	if err := checkScrapeRedirect(public, via); err == nil {
		t.Error("redirect chain over limit was allowed, want blocked")
	}
}

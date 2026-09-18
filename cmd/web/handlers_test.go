package main

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Zigelzi/roundtrip/internal/db"
)

// newTestApp builds an application backed by a fresh, migrated SQLite database
// in a temp file, and returns it alongside the raw handle for test setup.
func newTestApp(t *testing.T) (*application, *sql.DB) {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	database, err := db.InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	t.Cleanup(func() { database.Close() })
	if err := db.RunMigrations(database); err != nil {
		t.Fatalf("RunMigrations: %v", err)
	}
	return &application{queries: db.New(database)}, database
}

// Scenario: Visitor sees a personalized greeting.
func TestIndexGreetsSeededUser(t *testing.T) {
	app, database := newTestApp(t)

	// Default seeded user is greeted.
	if body := get(t, app, "/"); !strings.Contains(body, "Parent 1") {
		t.Errorf("response does not greet the seeded user %q", "Parent 1")
	}

	// The greeting must come from the database, not a hardcoded string:
	// change the stored name and the page must follow.
	const changed = "Testuser-42"
	if _, err := database.Exec("UPDATE app_user SET name = ? WHERE id = 1", changed); err != nil {
		t.Fatalf("update seeded name: %v", err)
	}
	body := get(t, app, "/")
	if !strings.Contains(body, changed) {
		t.Errorf("greeting did not reflect the updated database name %q", changed)
	}
	if strings.Contains(body, "Parent 1") {
		t.Errorf("greeting still shows the old name; it is not read from the database")
	}

	// The page must link the Tailwind stylesheet.
	if !strings.Contains(body, "/static/tailwind.css") {
		t.Errorf("index page does not link the Tailwind stylesheet")
	}
}

// Scenario: Tailwind stylesheet is generated and served.
func TestTailwindStylesheetServed(t *testing.T) {
	app, _ := newTestApp(t)

	rec := httptest.NewRecorder()
	app.routes().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/static/tailwind.css", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (did `tailwindcss` build run?)", rec.Code)
	}
	body := rec.Body.String()
	if len(strings.TrimSpace(body)) == 0 {
		t.Fatal("stylesheet is empty")
	}
	if !strings.Contains(body, "text-3xl") {
		t.Errorf("stylesheet has no rule for .text-3xl, a class used on the index page")
	}
}

// get performs a GET against the app's router and returns the response body,
// failing the test on a non-200 status.
func get(t *testing.T, app *application, path string) string {
	t.Helper()
	rec := httptest.NewRecorder()
	app.routes().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET %s: status = %d, want 200", path, rec.Code)
	}
	return rec.Body.String()
}

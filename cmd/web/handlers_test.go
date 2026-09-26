package main

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Zigelzi/roundtrip/internal/db"
	"github.com/pressly/goose/v3"
)

// newTestApp builds an application backed by a fresh, migrated SQLite database
// in a temp file, and returns it alongside the raw handle for test setup.
func newTestApp(t *testing.T) (*application, *sql.DB) {
	t.Helper()
	goose.SetLogger(goose.NopLogger())
	dbPath := filepath.Join(t.TempDir(), "test.db")
	database, err := db.InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	t.Cleanup(func() { database.Close() })
	if err := db.RunMigrations(database); err != nil {
		t.Fatalf("RunMigrations: %v", err)
	}
	app := newApplication(database)
	// A fixed default clock keeps date-dependent pages deterministic; tests
	// that care about "today" pin their own.
	app.now = func() time.Time { return time.Date(2026, 9, 1, 9, 0, 0, 0, time.UTC) }
	return app, database
}

// The front page links the Tailwind stylesheet (carried over from the
// Milestone 0 greeting scenario, whose greeting was removed in 01).
func TestIndexLinksStylesheet(t *testing.T) {
	app, _ := newTestApp(t)

	if body := get(t, app, "/"); !strings.Contains(body, "/static/tailwind.css") {
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

// send performs a request (form-encoded when form is non-nil) against the
// app's router and returns the recorder without following redirects.
func send(t *testing.T, app *application, method, path string, form url.Values) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(form.Encode()))
	if form != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	rec := httptest.NewRecorder()
	app.routes().ServeHTTP(rec, req)
	return rec
}

// inputTag returns the first <input ...> tag in body whose name attribute is
// name, so assertions can check that one field's attributes.
func inputTag(t *testing.T, body, name string) string {
	t.Helper()
	i := strings.Index(body, `name="`+name+`"`)
	if i < 0 {
		t.Fatalf("no input named %q in page", name)
	}
	start := strings.LastIndex(body[:i], "<input")
	end := strings.Index(body[i:], ">")
	if start < 0 || end < 0 {
		t.Fatalf("input named %q is not an <input> tag", name)
	}
	return body[start : i+end+1]
}

package main

import (
	"database/sql"
	"fmt"
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

// Helpers shared by tests across milestones: building the app, sending
// requests, setting up the database, and cutting pieces out of HTML.
// Helpers that only one milestone's tests use stay in that milestone's file.
//
// Several test files declare their own copies of view helpers (URL and id
// builders such as prepareItemURL or packedPartID), under the same names as
// in package view. That is deliberate, not duplication to tidy up: the tests
// are written before the code (spec/constitution.md), so they state the
// URLs and ids the page must have instead of importing the code's answer.

// App and requests.

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

// sendHTMX posts a form the way htmx does, marked with the HX-Request header.
func sendHTMX(t *testing.T, app *application, path string, form url.Values) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	rec := httptest.NewRecorder()
	app.routes().ServeHTTP(rec, req)
	return rec
}

// fixedNow parses an RFC 3339 instant for pinning the app's clock.
func fixedNow(t *testing.T, instant string) func() time.Time {
	t.Helper()
	at, err := time.Parse(time.RFC3339, instant)
	if err != nil {
		t.Fatalf("parse %q: %v", instant, err)
	}
	return func() time.Time { return at }
}

// Database setup.

func insertTrip(t *testing.T, database *sql.DB, destination, departure string, days int) int64 {
	t.Helper()
	res, err := database.Exec("INSERT INTO trip (destination, departure_date, duration_days) VALUES (?, ?, ?)", destination, departure, days)
	if err != nil {
		t.Fatalf("insert trip: %v", err)
	}
	id, _ := res.LastInsertId()
	return id
}

func memberID(t *testing.T, database *sql.DB, name string) int64 {
	t.Helper()
	var id int64
	if err := database.QueryRow("SELECT id FROM family_member WHERE name = ?", name).Scan(&id); err != nil {
		t.Fatalf("family member %q: %v", name, err)
	}
	return id
}

func insertItem(t *testing.T, database *sql.DB, tripID int64, member, name string, quantity int) int64 {
	t.Helper()
	res, err := database.Exec("INSERT INTO item (trip_id, family_member_id, name, name_key, quantity) VALUES (?, ?, ?, ?, ?)",
		tripID, memberID(t, database, member), name, strings.ToLower(name), quantity)
	if err != nil {
		t.Fatalf("insert item: %v", err)
	}
	id, _ := res.LastInsertId()
	return id
}

// prepareItem marks an already-inserted item prepared, standing in for the
// parent having gathered it on the bed on an earlier visit (milestone 08).
func prepareItem(t *testing.T, database *sql.DB, itemID int64) {
	t.Helper()
	if _, err := database.Exec("UPDATE item SET status = 'prepared' WHERE id = ?", itemID); err != nil {
		t.Fatalf("prepare item %d: %v", itemID, err)
	}
}

// packItem marks an already-inserted item packed, standing in for the parent
// having ticked it on an earlier visit.
func packItem(t *testing.T, database *sql.DB, itemID int64) {
	t.Helper()
	if _, err := database.Exec("UPDATE item SET status = 'packed' WHERE id = ?", itemID); err != nil {
		t.Fatalf("pack item %d: %v", itemID, err)
	}
}

// insertBasic adds a basic last in a member's list, as the add form will:
// the name key is lower-cased in Go (SQL lower() folds only ASCII, so "Ä"
// would not match), and a member's first basic gets position 1.
func insertBasic(t *testing.T, database *sql.DB, member int64, name string, perDay, fixed int) {
	t.Helper()
	_, err := database.Exec(`INSERT INTO basic_item (family_member_id, position, name, name_key, per_day, fixed)
		VALUES (?, (SELECT COALESCE(MAX(position), 0) + 1 FROM basic_item WHERE family_member_id = ?), ?, ?, ?, ?)`,
		member, member, name, strings.ToLower(name), perDay, fixed)
	if err != nil {
		t.Fatalf("insert basic %q: %v", name, err)
	}
}

func count(t *testing.T, database *sql.DB, table string) int {
	t.Helper()
	var n int
	if err := database.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&n); err != nil {
		t.Fatalf("count %s: %v", table, err)
	}
	return n
}

// URLs.

func tripURLFor(id int64) string {
	return fmt.Sprintf("/trips/%d", id)
}

func packURLFor(id int64) string {
	return fmt.Sprintf("/trips/%d/pack", id)
}

// Cutting HTML.

// memberSection returns the HTML of one family member's section on the trip
// page. The page contract: each member is a <section data-member="Name">.
func memberSection(t *testing.T, body, member string) string {
	t.Helper()
	i := strings.Index(body, `data-member="`+member+`"`)
	if i < 0 {
		t.Fatalf("no section for family member %q", member)
	}
	start := strings.LastIndex(body[:i], "<section")
	end := strings.Index(body[i:], "</section>")
	if start < 0 || end < 0 {
		t.Fatalf("section for %q is malformed", member)
	}
	return body[start : i+end]
}

// memberHeading is the first <h2> in a piece of the page.
func memberHeading(t *testing.T, s string) string {
	t.Helper()
	start := strings.Index(s, "<h2")
	end := strings.Index(s, "</h2>")
	if start < 0 || end < start {
		t.Fatalf("no <h2> heading in:\n%s", s)
	}
	return s[start:end]
}

// between returns the part of s from the first from up to the next to, or
// to the end of s when to does not follow.
func between(t *testing.T, s, from, to string) string {
	t.Helper()
	i := strings.Index(s, from)
	if i < 0 {
		t.Fatalf("no %q in:\n%s", from, s)
	}
	if j := strings.Index(s[i:], to); j >= 0 {
		return s[i : i+j]
	}
	return s[i:]
}

// enclosingTag returns the opening tag (starting with open, e.g. "<form")
// that comes last before marker in body.
func enclosingTag(t *testing.T, body, marker, open string) string {
	t.Helper()
	i := strings.Index(body, marker)
	if i < 0 {
		t.Fatalf("page has no %s", marker)
	}
	start := strings.LastIndex(body[:i], open)
	if start < 0 {
		t.Fatalf("no %s before %s", open, marker)
	}
	end := strings.Index(body[start:], ">")
	return body[start : start+end+1]
}

// tagWith returns the opening tag (starting with open, e.g. "<input") that
// contains marker, such as one of its attributes.
func tagWith(t *testing.T, body, marker, open string) string {
	t.Helper()
	i := strings.Index(body, marker)
	if i < 0 {
		t.Fatalf("page has no %s", marker)
	}
	start := strings.LastIndex(body[:i], open)
	end := strings.Index(body[i:], ">")
	if start < 0 || end < 0 {
		t.Fatalf("%s is not inside a %s tag", marker, open)
	}
	return body[start : i+end+1]
}

// inputTag returns the first <input ...> tag in body whose name attribute is
// name, so assertions can check that one field's attributes.
func inputTag(t *testing.T, body, name string) string {
	t.Helper()
	return tagWith(t, body, `name="`+name+`"`, "<input")
}

// inputTagLabelled returns the <input ...> tag with the given aria-label.
func inputTagLabelled(t *testing.T, body, label string) string {
	t.Helper()
	return tagWith(t, body, `aria-label="`+label+`"`, "<input")
}

// buttonTag returns the <button ...> opening tag with the given aria-label.
func buttonTag(t *testing.T, body, label string) string {
	t.Helper()
	return tagWith(t, body, `aria-label="`+label+`"`, "<button")
}

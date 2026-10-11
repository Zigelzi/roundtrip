package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

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

package main

import (
	"fmt"
	"html"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// S15 (automated part): adding and removing items in place. Whether the name
// field keeps focus on a phone is checked manually (see the spec).

func TestTripPageWiresInPlaceAdd(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)

	body := get(t, app, tripURLFor(tripID))

	if !strings.Contains(body, `src="/static/htmx.min.js"`) {
		t.Errorf("page does not load htmx")
	}
	section := memberSection(t, body, "Parent 1")
	parent := memberID(t, database, "Parent 1")
	if want := fmt.Sprintf(`hx-target="#items-member-%d"`, parent); !strings.Contains(section, want) {
		t.Errorf("Parent 1's add form does not target their item list (%s)", want)
	}
	if want := fmt.Sprintf(`id="items-member-%d"`, parent); !strings.Contains(section, want) {
		t.Errorf("Parent 1's item list has no %s to swap", want)
	}
}

func TestAddItemInPlaceRepliesWithOnlyTheList(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	parent := memberID(t, database, "Parent 1")

	rec := sendHTMX(t, app, fmt.Sprintf("/trips/%d/members/%d/items", tripID, parent), url.Values{"name": {"Underwear"}, "quantity": {"4"}})

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	assertOnlyMemberList(t, rec.Body.String(), parent)
	if body := rec.Body.String(); !strings.Contains(body, "Underwear") || !strings.Contains(body, "4 pcs") {
		t.Errorf("reply does not list Underwear, 4 pcs:\n%s", body)
	}
	if n := count(t, database, "item"); n != 1 {
		t.Errorf("%d items saved, want 1", n)
	}
}

func TestAddItemInPlaceShowsErrorsInTheList(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	parent := memberID(t, database, "Parent 1")

	rec := sendHTMX(t, app, fmt.Sprintf("/trips/%d/members/%d/items", tripID, parent), url.Values{"name": {"Socks"}, "quantity": {"0"}})

	if rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want 422", rec.Code)
	}
	assertOnlyMemberList(t, rec.Body.String(), parent)
	if want := html.EscapeString("Quantity must be a whole number, at least 1"); !strings.Contains(rec.Body.String(), want) {
		t.Errorf("reply does not explain the problem")
	}
}

func TestRemoveItemInPlaceRepliesWithOnlyTheList(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	item := insertItem(t, database, tripID, "Parent 1", "Underwear", 4)
	insertItem(t, database, tripID, "Parent 1", "Socks", 2)

	rec := sendHTMX(t, app, fmt.Sprintf("/trips/%d/items/%d/delete", tripID, item), nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	assertOnlyMemberList(t, body, memberID(t, database, "Parent 1"))
	if strings.Contains(body, "Underwear") || !strings.Contains(body, "Socks") {
		t.Errorf("reply should list Socks but not Underwear:\n%s", body)
	}
}

// assertOnlyMemberList checks a reply is just one member's item list: not a
// whole page, and without the add form (which must stay put to keep focus).
func assertOnlyMemberList(t *testing.T, body string, memberID int64) {
	t.Helper()
	if strings.Contains(body, "<html") {
		t.Errorf("reply is a whole page, want only the member's list")
	}
	if strings.Contains(body, `name="name"`) {
		t.Errorf("reply contains the add form; replacing it would lose focus")
	}
	if want := fmt.Sprintf(`id="items-member-%d"`, memberID); !strings.Contains(body, want) {
		t.Errorf("reply lacks %s", want)
	}
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

// S16 (automated part): the − and + buttons exist next to the quantity and
// are type="button" — a button's default type is submit, so otherwise a
// tap would add the item instead of changing the number.
func TestQuantityStepperButtonsDoNotSubmit(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)

	section := memberSection(t, get(t, app, tripURLFor(tripID)), "Parent 1")

	for _, label := range []string{"Decrease quantity", "Increase quantity"} {
		tag := buttonTag(t, section, label)
		if !strings.Contains(tag, `type="button"`) {
			t.Errorf("%q button would submit the form: %s", label, tag)
		}
	}
}

// buttonTag returns the <button ...> opening tag with the given aria-label.
func buttonTag(t *testing.T, body, label string) string {
	t.Helper()
	i := strings.Index(body, `aria-label="`+label+`"`)
	if i < 0 {
		t.Fatalf("no button labelled %q", label)
	}
	start := strings.LastIndex(body[:i], "<button")
	end := strings.Index(body[i:], ">")
	if start < 0 || end < 0 {
		t.Fatalf("%q is not a <button>", label)
	}
	return body[start : i+end+1]
}

// In `make dev`, templ's live-reload proxy injects a script into every HTML
// reply. htmx fragments must opt out, or each swap would wrap the list in
// <html><body> and add another reload script. Full pages keep the script.
func TestHTMXRepliesSkipDevReloadInjection(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	parent := memberID(t, database, "Parent 1")

	rec := sendHTMX(t, app, fmt.Sprintf("/trips/%d/members/%d/items", tripID, parent), url.Values{"name": {"Socks"}, "quantity": {"1"}})
	if got := rec.Header().Get("templ-skip-modify"); got != "true" {
		t.Errorf("htmx reply templ-skip-modify = %q, want \"true\"", got)
	}

	page := send(t, app, http.MethodGet, tripURLFor(tripID), nil)
	if got := page.Header().Get("templ-skip-modify"); got != "" {
		t.Errorf("full page has templ-skip-modify = %q; it would lose live reload", got)
	}
}

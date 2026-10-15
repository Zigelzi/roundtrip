package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

// Milestone 09 (spec/milestones/09-visual-refresh.md), the automated halves
// of the scenarios. The look itself is checked on a phone.

// S2: Basics is beside the title on the trips page.
func TestTripsPageHasBasicsBesideTheTitle(t *testing.T) {
	app, _ := newTestApp(t)

	header := pageHeader(t, get(t, app, "/"))

	if !strings.Contains(header, ">Roundtrip</h1>") {
		t.Errorf("trips page header has no Roundtrip title:\n%s", header)
	}
	basics := strings.Index(header, `href="/basics"`)
	newTrip := strings.Index(header, `href="/trips/new"`)
	if basics < 0 {
		t.Errorf("trips page header has no link to the basics:\n%s", header)
	}
	if newTrip < 0 {
		t.Errorf("trips page header has no New trip link:\n%s", header)
	}
	if basics >= 0 && newTrip >= 0 && basics > newTrip {
		t.Errorf("Basics comes after New trip; it belongs beside the title, above New trip")
	}
}

// S3: A parent with no trips is pointed to the basics first.
func TestNoTripsPointsToTheBasics(t *testing.T) {
	app, _ := newTestApp(t)

	_, main, ok := strings.Cut(get(t, app, "/"), "</header>")
	if !ok {
		t.Fatalf("trips page has no <header>")
	}

	for _, want := range []string{"No trips yet", "basics already on the list"} {
		if !strings.Contains(main, want) {
			t.Errorf("empty trips page does not say %q", want)
		}
	}
	link := linkWithText(t, main, "Check basics")
	if !strings.Contains(link, `href="/basics"`) {
		t.Errorf("Check basics does not go to the basics page: %s", link)
	}
}

// S4: Every inner page has a back button at the top.
func TestInnerPagesHaveABackButton(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	insertItem(t, database, tripID, "Parent 1", "Socks", 4)

	pages := []struct {
		name, path, label, href string
		sticky                  bool
	}{
		{"trip page", tripURLFor(tripID), "Back to trips", "/", true},
		{"packing page", packURLFor(tripID), "Back to trip", tripURLFor(tripID), true},
		{"basics page", basicsURL, "Back to trips", "/", true},
		{"new trip page", "/trips/new", "Back to trips", "/", false},
	}
	for _, p := range pages {
		t.Run(p.name, func(t *testing.T) {
			body := get(t, app, p.path)
			if p.sticky {
				body = pageHeader(t, body)
			}
			link := tagWith(t, body, `aria-label="`+p.label+`"`, "<a")
			if !strings.Contains(link, `href="`+p.href+`"`) {
				t.Errorf("%s does not go to %s: %s", p.label, p.href, link)
			}
		})
	}
}

// S5: Tapping anywhere on a row opens it for editing. The row's tap is the
// button that opens the edit, so the name has to be inside it, and ✕ has to
// be a button of its own.
func TestWholeTripRowOpensTheEdit(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	insertItem(t, database, tripID, "Parent 1", "Socks", 4)

	body := get(t, app, tripURLFor(tripID))

	open := buttonHTML(t, body, "Change quantity of Socks")
	for _, want := range []string{">Socks<", "4 pcs"} {
		if !strings.Contains(open, want) {
			t.Errorf("the button that opens the edit does not hold %q: %s", want, open)
		}
	}
	if strings.Contains(open, "Remove") {
		t.Errorf("✕ is inside the button that opens the edit: %s", open)
	}
	buttonTag(t, body, "Remove Socks")
}

// S5, the basics page.
func TestWholeBasicRowOpensTheEdit(t *testing.T) {
	app, database := newTestApp(t)
	insertBasic(t, database, memberID(t, database, "Child 1"), "Snorkel", 1, 1)

	body := get(t, app, basicsURL)

	open := buttonHTML(t, body, "Change quantity of Snorkel")
	for _, want := range []string{">Snorkel<", "1 / day + 1 extra"} {
		if !strings.Contains(open, want) {
			t.Errorf("the button that opens the edit does not hold %q: %s", want, open)
		}
	}
	if strings.Contains(open, "Remove") {
		t.Errorf("✕ is inside the button that opens the edit: %s", open)
	}
	buttonTag(t, body, "Remove Snorkel")
}

// S8: The packing page header shows the trip and its progress.
func TestPackingHeaderShowsTripAndProgress(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	parent := memberID(t, database, "Parent 1")
	socks := insertItem(t, database, tripID, "Parent 1", "Socks", 4)
	prepareItem(t, database, socks)
	insertItem(t, database, tripID, "Parent 1", "Hat", 1)
	insertItem(t, database, tripID, "Parent 2", "Raincoat", 1)

	header := pageHeader(t, get(t, app, packURLFor(tripID)))
	if !strings.Contains(header, "Parainen</h1>") {
		t.Errorf("packing header's title is not the trip's name:\n%s", header)
	}
	if !strings.Contains(header, "Packing · 0 of 3 packed") {
		t.Errorf("packing header does not read \"Packing · 0 of 3 packed\":\n%s", header)
	}

	rec := sendHTMX(t, app, packItemURL(tripID, parent, socks), url.Values{"packed_shown": {"false"}})
	if rec.Code != http.StatusOK {
		t.Fatalf("pack: status = %d, want 200", rec.Code)
	}
	if progress := packFragment(t, rec.Body.String(), "pack-progress"); !strings.Contains(progress, "Packing · 1 of 3 packed") {
		t.Errorf("reply to packing does not read \"Packing · 1 of 3 packed\":\n%s", progress)
	}
}

// S11: The packing page shows a quantity only when there is more than one.
// The trip page keeps "1 pc" (trip_page_test.go).
func TestPackingPageHidesASingleQuantity(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	insertItem(t, database, tripID, "Parent 1", "Underwear", 4)
	insertItem(t, database, tripID, "Parent 1", "Toothbrush", 1)
	insertItem(t, database, tripID, "Parent 1", "Hat", 1)
	packItem(t, database, insertItem(t, database, tripID, "Parent 1", "Swimsuit", 1))
	prepareItem(t, database, insertItem(t, database, tripID, "Parent 1", "Towel", 1))

	section := memberSection(t, get(t, app, packURLFor(tripID)), "Parent 1")

	if row := packRow(t, section, "Underwear"); !strings.Contains(row, "4 pcs") {
		t.Errorf("Underwear does not show 4 pcs: %s", row)
	}
	for _, name := range []string{"Toothbrush", "Towel", "Swimsuit"} {
		if row := packRow(t, section, name); strings.Contains(row, "1 pc") {
			t.Errorf("%s shows \"1 pc\" on the packing page: %s", name, row)
		}
	}
}

// pageHeader is the page's <header>, the sticky top bar.
func pageHeader(t *testing.T, body string) string {
	t.Helper()
	start, end := strings.Index(body, "<header"), strings.Index(body, "</header>")
	if start < 0 || end < start {
		t.Fatalf("page has no <header>")
	}
	return body[start:end]
}

// buttonHTML is the whole <button> labelled label, from its opening tag to
// its </button>, so a test can see what it holds.
func buttonHTML(t *testing.T, body, label string) string {
	t.Helper()
	tag := buttonTag(t, body, label)
	start := strings.Index(body, tag)
	end := strings.Index(body[start:], "</button>")
	if end < 0 {
		t.Fatalf("button %q is not closed", label)
	}
	return body[start : start+end]
}

// linkWithText is the <a> tag of the first link whose text includes text.
func linkWithText(t *testing.T, body, text string) string {
	t.Helper()
	for _, part := range strings.Split(body, "<a ")[1:] {
		if inner, _, ok := strings.Cut(part, "</a>"); ok && strings.Contains(inner, text) {
			tag, _, _ := strings.Cut(part, ">")
			return fmt.Sprintf("<a %s>", tag)
		}
	}
	t.Fatalf("no link with text %q", text)
	return ""
}

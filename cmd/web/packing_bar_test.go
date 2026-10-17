package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// Milestone 12 (spec/milestones/12-packing-bar.md), the packing button as a
// bar at the bottom of the trip page.

// packingBarID is the bar's id: tests read it alone, and add and remove
// replies swap it out of band.
const packingBarID = "packing-bar"

// packingBar returns the bar from its opening tag up to the end of its link,
// which is everything the bar says.
func packingBar(t *testing.T, body string) string {
	t.Helper()
	marker := `id="` + packingBarID + `"`
	i := strings.Index(body, marker)
	if i < 0 {
		t.Fatalf("no packing bar (%s) in:\n%s", marker, body)
	}
	start := strings.LastIndex(body[:i], "<")
	end := strings.Index(body[i:], "</a>")
	if end < 0 {
		t.Fatalf("packing bar has no link:\n%s", body[start:])
	}
	return body[start : i+end]
}

// wantBarLabel checks the bar reads label and none of the other two labels.
func wantBarLabel(t *testing.T, bar, label string) {
	t.Helper()
	for _, l := range []string{"Start packing", "Continue packing", "All packed"} {
		if got := strings.Contains(bar, l); got != (l == label) {
			t.Errorf("bar contains %q = %v, want %v:\n%s", l, got, l == label, bar)
		}
	}
}

// S1: The packing button is at the bottom of the trip page. Also Q1 (a trip
// with no items) and Q2 (prepared only), which read "Start packing" too.
func TestPackingBarStartsPacking(t *testing.T) {
	cases := []struct {
		name     string
		items    bool
		prepared bool
	}{
		{"nothing packed", true, false},
		{"no items yet (Q1)", false, false},
		{"prepared only (Q2)", true, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			app, database := newTestApp(t)
			tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
			if c.items {
				socks := insertItem(t, database, tripID, "Parent 1", "Socks", 3)
				insertItem(t, database, tripID, "Parent 1", "Hat", 1)
				insertItem(t, database, tripID, "Parent 1", "Towel", 1)
				if c.prepared {
					prepareItem(t, database, socks)
				}
			}

			body := get(t, app, tripURLFor(tripID))

			bar := packingBar(t, body)
			wantBarLabel(t, bar, "Start packing")
			if want := `href="` + packURLFor(tripID) + `"`; !strings.Contains(bar, want) {
				t.Errorf("bar does not open the packing page (%s):\n%s", want, bar)
			}
			header := between(t, body, "<header", "</header>")
			if strings.Contains(header, packURLFor(tripID)) {
				t.Errorf("header still links to the packing page:\n%s", header)
			}
		})
	}
}

// S2: Part way through, the bar shows how far packing has got.
func TestPackingBarShowsPackedCount(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	packItem(t, database, insertItem(t, database, tripID, "Parent 1", "Socks", 3))
	insertItem(t, database, tripID, "Parent 1", "Hat", 1)
	insertItem(t, database, tripID, "Parent 2", "Towel", 1)

	bar := packingBar(t, get(t, app, tripURLFor(tripID)))

	wantBarLabel(t, bar, "Continue packing")
	if tag := enclosingTag(t, bar, ">1/3<", "<span"); !strings.Contains(tag, `aria-hidden="true"`) {
		t.Errorf("the visible 1/3 is read aloud: %s", tag)
	}
	if !strings.Contains(bar, "1 of 3 packed") {
		t.Errorf("bar has no spoken \"1 of 3 packed\":\n%s", bar)
	}
}

// S3: When everything is packed, the bar says so.
func TestPackingBarAllPacked(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	for _, name := range []string{"Socks", "Hat", "Towel"} {
		packItem(t, database, insertItem(t, database, tripID, "Parent 1", name, 1))
	}

	bar := packingBar(t, get(t, app, tripURLFor(tripID)))

	wantBarLabel(t, bar, "All packed")
	if want := `href="` + packURLFor(tripID) + `"`; !strings.Contains(bar, want) {
		t.Errorf("bar does not open the packing page (%s):\n%s", want, bar)
	}
}

// S4: The count follows adding and removing items. The htmx reply carries
// the bar as an out-of-band fragment beside the member's list.
func TestPackingBarFollowsAddAndRemove(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	parent := memberID(t, database, "Parent 1")
	socks := insertItem(t, database, tripID, "Parent 1", "Socks", 3)
	packItem(t, database, socks)
	hat := insertItem(t, database, tripID, "Parent 1", "Hat", 1)
	towel := insertItem(t, database, tripID, "Parent 1", "Towel", 1)

	reply := func(rec *httptest.ResponseRecorder) string {
		t.Helper()
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
		bar := packingBar(t, rec.Body.String())
		if tag := enclosingTag(t, bar, `id="`+packingBarID+`"`, "<"); !strings.Contains(tag, `hx-swap-oob="true"`) {
			t.Errorf("bar in the reply is not an out-of-band swap: %s", tag)
		}
		return bar
	}

	rec := sendHTMX(t, app, fmt.Sprintf("/trips/%d/members/%d/items", tripID, parent), url.Values{"name": {"Sunscreen"}, "quantity": {"1"}})
	bar := reply(rec)
	wantBarLabel(t, bar, "Continue packing")
	if !strings.Contains(bar, "1/4") {
		t.Errorf("after adding Sunscreen, bar does not read 1/4:\n%s", bar)
	}

	rec = sendHTMX(t, app, fmt.Sprintf("/trips/%d/items/%d/delete", tripID, socks), nil)
	wantBarLabel(t, reply(rec), "Start packing")

	packItem(t, database, hat)
	packItem(t, database, towel)
	var sunscreen int64
	if err := database.QueryRow("SELECT id FROM item WHERE trip_id = ? AND name = 'Sunscreen'", tripID).Scan(&sunscreen); err != nil {
		t.Fatalf("find Sunscreen: %v", err)
	}
	rec = sendHTMX(t, app, fmt.Sprintf("/trips/%d/items/%d/delete", tripID, sunscreen), nil)
	wantBarLabel(t, reply(rec), "All packed")
}

// S7 (observable part): an open row carries the mark the hiding rule looks
// for (data-editing), and the page after saving does not. That the browser
// then hides the bar is checked on a phone.
func TestOpenRowMarksTheBarHidden(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	socks := insertItem(t, database, tripID, "Parent 1", "Socks", 3)

	open := get(t, app, fmt.Sprintf("%s?edit=%d", tripURLFor(tripID), socks))
	items := between(t, open, `id="trip-items"`, "<datalist")
	if !strings.Contains(items, "data-editing") {
		t.Errorf("the open row has no data-editing mark inside the swapped items")
	}
	packingBar(t, open)

	// Saving redirects, and htmx follows the redirect to the page it swaps
	// in, so the page at the redirect is what the parent sees after saving.
	rec := send(t, app, http.MethodPost, fmt.Sprintf("/trips/%d/items/%d", tripID, socks), url.Values{"quantity": {"4"}})
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("saving: status = %d, want 303", rec.Code)
	}
	saved, _, _ := strings.Cut(rec.Header().Get("Location"), "#")
	if strings.Contains(get(t, app, saved), "data-editing") {
		t.Errorf("page after saving (%s) still carries data-editing", saved)
	}
}

// Code review, finding 1: coming back with the Back gesture must not show
// the bar's count from before packing. Browsers reuse a cached copy of the
// page on Back unless told not to, so the trip page says no-store. (The
// pageshow reload covers the other case, a page kept whole in memory; a
// manual check.)
func TestTripPageIsNotReusedOnBack(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)

	rec := send(t, app, http.MethodGet, tripURLFor(tripID), nil)

	if got := rec.Header().Get("Cache-Control"); !strings.Contains(got, "no-store") {
		t.Errorf("Cache-Control = %q, want no-store", got)
	}
}

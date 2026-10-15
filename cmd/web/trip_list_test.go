package main

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"
)

// Milestone 14 (spec/milestones/14-trip-list-progress.md): the trips list's
// past trips group and the packing progress on upcoming trips.

// pastGroup splits the trips list at the past trips group: what comes before
// it (the upcoming trips), the group's opening <details> tag, and the group
// itself. ok is false when the page has no past trips group.
func pastGroup(t *testing.T, body string) (upcoming, openTag, group string, ok bool) {
	t.Helper()
	i := strings.Index(body, "Past trips (")
	if i < 0 {
		return body, "", "", false
	}
	start := strings.LastIndex(body[:i], "<details")
	if start < 0 {
		t.Fatalf("the past trips heading is not inside a <details> group")
	}
	end := strings.Index(body[start:], "</details>")
	if end < 0 {
		t.Fatalf("the past trips group is not closed")
	}
	tagEnd := strings.Index(body[start:], ">")
	return body[:start], body[start : start+tagEnd+1], body[start : start+end], true
}

// S1: Past trips have their own group.
// S6: A parent can open the past trips group (the observable part: it is a
// <details> group whose summary is its heading).
func TestPastTripsHaveTheirOwnGroup(t *testing.T) {
	app, database := newTestApp(t)
	app.now = fixedNow(t, "2026-09-19T09:00:00Z") // 19.9.2026 in Finland

	// Inserted out of order so the order must come from sorting.
	insertTrip(t, database, "Kuopio", "2026-06-10", 2)
	insertTrip(t, database, "Levi", "2026-12-01", 5)
	insertTrip(t, database, "Turku", "2026-08-01", 3)
	insertTrip(t, database, "Parainen", "2026-09-20", 3)

	body := get(t, app, "/")

	upcoming, openTag, group, ok := pastGroup(t, body)
	if !ok {
		t.Fatalf("no past trips group on the trips list")
	}
	parainen, levi := strings.Index(upcoming, "Parainen"), strings.Index(upcoming, "Levi")
	if parainen < 0 || levi < 0 || parainen > levi {
		t.Errorf("upcoming trips: Parainen@%d, Levi@%d; want both above the past group, nearest first", parainen, levi)
	}
	if !strings.Contains(group, "Past trips (2)") {
		t.Errorf("past group heading does not read \"Past trips (2)\":\n%s", group)
	}
	summary := between(t, group, "<summary", "</summary>")
	if !strings.Contains(summary, "Past trips (2)") {
		t.Errorf("\"Past trips (2)\" is not the group's summary, so it cannot be tapped to open it (S6):\n%s", summary)
	}
	turku, kuopio := strings.Index(group, "Turku"), strings.Index(group, "Kuopio")
	if turku < 0 || kuopio < 0 || turku > kuopio {
		t.Errorf("past trips: Turku@%d, Kuopio@%d; want both in the group, most recent first", turku, kuopio)
	}
	// 14/R9: closed by default, past trips are rarely what the app is
	// opened for.
	if strings.Contains(openTag, " open") {
		t.Errorf("past trips group is open by default, want closed: %s", openTag)
	}
}

// S2: A trip in progress is not past.
// S3: A trip is past from the day after its return.
func TestTripIsPastFromTheDayAfterItsReturn(t *testing.T) {
	tests := []struct {
		name string
		now  string
		past bool
	}{
		{"the day before departure", "2026-09-19T09:00:00Z", false},
		{"departure day", "2026-09-20T09:00:00Z", false},
		{"during the trip", "2026-09-21T09:00:00Z", false},
		{"return day", "2026-09-22T09:00:00Z", false},
		{"the day after return", "2026-09-23T09:00:00Z", true},
		// 22:30 UTC on 22.9 is already 23.9 in Finland (01/S5: "today" is
		// the Finnish date).
		{"just after midnight in Finland", "2026-09-22T22:30:00Z", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app, database := newTestApp(t)
			app.now = fixedNow(t, tt.now)
			insertTrip(t, database, "Parainen", "2026-09-20", 3) // 20.–22.9.2026

			body := get(t, app, "/")

			upcoming, _, group, _ := pastGroup(t, body)
			inPast := strings.Contains(group, "Parainen")
			inUpcoming := strings.Contains(upcoming, "Parainen")
			if inPast != tt.past || inUpcoming == tt.past {
				t.Errorf("Parainen in upcoming = %v, in past = %v; want past = %v", inUpcoming, inPast, tt.past)
			}
		})
	}
}

// S4: No past group without past trips.
func TestNoPastGroupWithoutPastTrips(t *testing.T) {
	app, database := newTestApp(t)
	app.now = fixedNow(t, "2026-09-19T09:00:00Z")
	insertTrip(t, database, "Parainen", "2026-09-20", 3)
	insertTrip(t, database, "Levi", "2026-12-01", 5)

	body := get(t, app, "/")

	if _, _, _, ok := pastGroup(t, body); ok {
		t.Errorf("trips list has a past trips group with no past trips")
	}
}

// S5: Only past trips.
func TestOnlyPastTrips(t *testing.T) {
	app, database := newTestApp(t)
	app.now = fixedNow(t, "2026-09-19T09:00:00Z")
	insertTrip(t, database, "Kuopio", "2026-06-10", 2)

	body := get(t, app, "/")

	upcoming, openTag, group, ok := pastGroup(t, body)
	if !ok {
		t.Fatalf("no past trips group on a list of only past trips")
	}
	if !strings.Contains(group, "Past trips (1)") || !strings.Contains(group, "Kuopio") {
		t.Errorf("past group does not hold Kuopio under \"Past trips (1)\":\n%s", group)
	}
	if strings.Contains(openTag, " open") {
		t.Errorf("past trips group is open, want closed (14/R9): %s", openTag)
	}
	if strings.Contains(body, "No trips yet") {
		t.Errorf("trips list says \"No trips yet\" while it has a past trip")
	}
	// Q3: no empty upcoming section above the group.
	if strings.Contains(upcoming, "<ul") {
		t.Errorf("trips list has an upcoming list with no upcoming trips")
	}
}

// tripRow is the trips list row of the trip going to destination: its whole
// <li>, from the list item to its end.
func tripRow(t *testing.T, body, destination string) string {
	t.Helper()
	i := strings.Index(body, ">"+destination+"<")
	if i < 0 {
		t.Fatalf("trips list has no trip to %q", destination)
	}
	start := strings.LastIndex(body[:i], "<li")
	end := strings.Index(body[i:], "</li>")
	if start < 0 || end < 0 {
		t.Fatalf("trip to %q is not a list item", destination)
	}
	return body[start : i+end]
}

// insertItems adds n items for Parent 1, the first packed of them packed and
// the next prepared of them prepared, the rest left planned.
func insertItems(t *testing.T, database *sql.DB, tripID int64, n, packed, prepared int) {
	t.Helper()
	for i := range n {
		id := insertItem(t, database, tripID, "Parent 1", fmt.Sprintf("Item %d", i+1), 1)
		switch {
		case i < packed:
			packItem(t, database, id)
		case i < packed+prepared:
			prepareItem(t, database, id)
		}
	}
}

// S7: An upcoming trip shows how far packing has got.
func TestUpcomingTripShowsPackingProgress(t *testing.T) {
	app, database := newTestApp(t)
	app.now = fixedNow(t, "2026-09-19T09:00:00Z")
	levi := insertTrip(t, database, "Levi", "2026-12-01", 5)
	insertItems(t, database, levi, 40, 18, 6)

	row := tripRow(t, get(t, app, "/"), "Levi")

	if !strings.Contains(row, "18 of 40 packed") {
		t.Errorf("Levi's row does not read \"18 of 40 packed\":\n%s", row)
	}
	// 18 of 40 is 45%, 6 of 40 is 15%.
	wantBar(t, "Levi's row", row, 45, 15)
	// Screen readers hear the words; the bar is hidden from them: the
	// segments sit inside an aria-hidden element that has not closed yet.
	if i := strings.Index(row, `data-bar="packed"`); i >= 0 {
		h := strings.LastIndex(row[:i], `aria-hidden="true"`)
		if h < 0 || strings.Contains(row[h:i], "</") {
			t.Errorf("the bar is not hidden from screen readers:\n%s", row)
		}
	}
	// The same count as the trip page's packing bar (14/R4, 12/S2).
	if bar := packingBar(t, get(t, app, tripURLFor(levi))); !strings.Contains(bar, "18 of 40 packed") {
		t.Errorf("the trip page's packing bar does not agree with the list's \"18 of 40 packed\":\n%s", bar)
	}
}

// S8: A fully packed trip reads "All packed".
func TestFullyPackedTripReadsAllPacked(t *testing.T) {
	app, database := newTestApp(t)
	app.now = fixedNow(t, "2026-09-19T09:00:00Z")
	levi := insertTrip(t, database, "Levi", "2026-12-01", 5)
	insertItems(t, database, levi, 3, 3, 0)

	row := tripRow(t, get(t, app, "/"), "Levi")

	if !strings.Contains(row, "All packed") {
		t.Errorf("Levi's row does not read \"All packed\":\n%s", row)
	}
	if strings.Contains(row, "3 of 3 packed") {
		t.Errorf("Levi's row reads \"3 of 3 packed\" beside \"All packed\":\n%s", row)
	}
	wantBar(t, "Levi's row", row, 100, 0)
}

// S9: A trip with nothing on it shows no progress.
// S10: Past trips show no progress.
func TestNoProgressWithoutItemsOrOnPastTrips(t *testing.T) {
	app, database := newTestApp(t)
	app.now = fixedNow(t, "2026-09-19T09:00:00Z")
	insertTrip(t, database, "Levi", "2026-12-01", 5) // upcoming, no items
	kuopio := insertTrip(t, database, "Kuopio", "2026-06-10", 2)
	insertItems(t, database, kuopio, 4, 2, 1)

	body := get(t, app, "/")

	for _, trip := range []struct{ destination, dates string }{
		{"Levi", "1.–5.12.2026"},
		{"Kuopio", "10.–11.6.2026"},
	} {
		row := tripRow(t, body, trip.destination)
		if !strings.Contains(row, trip.dates) {
			t.Errorf("%s's row does not show its dates %q:\n%s", trip.destination, trip.dates, row)
		}
		if strings.Contains(row, "data-bar") || strings.Contains(row, "packed") {
			t.Errorf("%s's row shows packing progress, want none:\n%s", trip.destination, row)
		}
	}
}

// S1: past trips are listed most recent first, by when they ended: a long
// trip that departed earlier but came back later is above a short one that
// sat inside it (14, Code review).
func TestPastTripsMostRecentReturnFirst(t *testing.T) {
	app, database := newTestApp(t)
	app.now = fixedNow(t, "2026-09-20T09:00:00Z")
	insertTrip(t, database, "Lapland", "2026-09-01", 14) // 1.–14.9
	insertTrip(t, database, "Turku", "2026-09-05", 2)    // 5.–6.9

	_, _, group, ok := pastGroup(t, get(t, app, "/"))
	if !ok {
		t.Fatalf("no past trips group")
	}
	lapland, turku := strings.Index(group, "Lapland"), strings.Index(group, "Turku")
	if lapland < 0 || turku < 0 || lapland > turku {
		t.Errorf("past trips: Lapland@%d, Turku@%d; want Lapland first, it came back last", lapland, turku)
	}
}

package main

import (
	"database/sql"
	"html"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

// Milestone 01 (spec/milestones/01-packing-list.md), trip scenarios S1–S6.

// S1: Parent with no trips is prompted to create one.
func TestNoTripsPromptsToCreate(t *testing.T) {
	app, _ := newTestApp(t)

	body := get(t, app, "/")

	if !strings.Contains(body, "No trips yet") {
		t.Errorf("front page does not say there are no trips yet")
	}
	if !strings.Contains(body, `href="/trips/new"`) {
		t.Errorf("front page has no link to create a trip")
	}
}

// S2: Parent sees their trips, upcoming first.
func TestTripsListedUpcomingFirst(t *testing.T) {
	app, database := newTestApp(t)
	app.now = fixedNow(t, "2026-09-19T09:00:00Z") // 19.9.2026 in Finland

	// Inserted out of order so the list order must come from sorting.
	insertTrip(t, database, "Far away", "2026-12-01", 5)
	insertTrip(t, database, "Past trip", "2026-06-10", 2)
	insertTrip(t, database, "Parainen", "2026-09-20", 3)

	body := get(t, app, "/")

	for _, want := range []string{"Parainen", "20.–22.9.2026", "3 days", "Far away", "1.–5.12.2026", "5 days", "Past trip", "10.–11.6.2026", "2 days"} {
		if !strings.Contains(body, want) {
			t.Errorf("front page does not show %q", want)
		}
	}
	nearest, later, past := strings.Index(body, "Parainen"), strings.Index(body, "Far away"), strings.Index(body, "Past trip")
	if !(nearest < later && later < past) {
		t.Errorf("order = Parainen@%d, Far away@%d, Past trip@%d; want nearest upcoming, later upcoming, then past", nearest, later, past)
	}
	if !strings.Contains(body, `href="/trips/`) {
		t.Errorf("trips are not linked to their pages")
	}
}

// S3: Create-trip form starts with sensible defaults.
func TestCreateTripFormDefaults(t *testing.T) {
	app, _ := newTestApp(t)

	body := get(t, app, "/trips/new")

	date := inputTag(t, body, "departure_date")
	if strings.Contains(date, "value=") && !strings.Contains(date, `value=""`) {
		t.Errorf("departure date is prefilled: %s", date)
	}
	end := inputTag(t, body, "end_date")
	if strings.Contains(end, "value=") && !strings.Contains(end, `value=""`) {
		t.Errorf("return date is prefilled: %s", end)
	}
	if duration := inputTag(t, body, "duration_days"); !strings.Contains(duration, `value="3"`) {
		t.Errorf("duration does not default to 3: %s", duration)
	}
}

// S4: Parent creates a trip.
func TestCreateTrip(t *testing.T) {
	app, database := newTestApp(t)
	app.now = fixedNow(t, "2026-09-19T09:00:00Z")

	rec := send(t, app, http.MethodPost, "/trips", url.Values{
		"destination":    {"Parainen"},
		"departure_date": {"2026-09-20"},
		"duration_days":  {"3"},
	})

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want 303 redirect to the new trip", rec.Code)
	}
	var id int64
	var destination, departure string
	var duration int64
	err := database.QueryRow("SELECT id, destination, departure_date, duration_days FROM trip").Scan(&id, &destination, &departure, &duration)
	if err != nil {
		t.Fatalf("trip was not saved: %v", err)
	}
	if destination != "Parainen" || departure != "2026-09-20" || duration != 3 {
		t.Errorf("saved trip = %q %q %d, want Parainen 2026-09-20 3", destination, departure, duration)
	}
	if loc, want := rec.Header().Get("Location"), tripURLFor(id); loc != want {
		t.Errorf("redirect to %q, want the trip's page %q", loc, want)
	}
}

// S5: Parent creates a trip departing today.
func TestCreateTripDepartingToday(t *testing.T) {
	app, database := newTestApp(t)
	// 22:30 UTC on 19.9 is already 20.9 (01:30) in Finland, so this also
	// proves "today" is the Finnish date, not the UTC one.
	app.now = fixedNow(t, "2026-09-19T22:30:00Z")

	rec := send(t, app, http.MethodPost, "/trips", url.Values{
		"destination":    {"Parainen"},
		"departure_date": {"2026-09-20"},
		"duration_days":  {"3"},
	})

	if rec.Code != http.StatusSeeOther {
		t.Errorf("status = %d, want 303: a trip departing today must be accepted", rec.Code)
	}
	if n := count(t, database, "trip"); n != 1 {
		t.Errorf("%d trips saved, want 1", n)
	}
}

// S6: Invalid trip input is rejected.
func TestCreateTripRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name        string
		destination string
		departure   string
		end         string
		duration    string
		wantMessage string
	}{
		{"empty destination", "  ", "2026-09-25", "", "3", "Enter a destination"},
		{"no departure date", "Parainen", "", "", "3", "Choose a departure date"},
		{"departure before today", "Parainen", "2026-09-19", "", "3", "Departure date can't be in the past"},
		{"return before departure", "Parainen", "2026-09-25", "2026-09-24", "3", "Return date can't be before the departure date"},
		{"duration below 1", "Parainen", "2026-09-25", "", "0", "Duration must be a whole number of days, at least 1"},
		{"duration not a whole number", "Parainen", "2026-09-25", "", "2.5", "Duration must be a whole number of days, at least 1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app, database := newTestApp(t)
			// 20.9.2026 in Finland but still 19.9 in UTC: rejecting 19.9 proves
			// "before today" uses the Finnish date.
			app.now = fixedNow(t, "2026-09-19T22:30:00Z")

			rec := send(t, app, http.MethodPost, "/trips", url.Values{
				"destination":    {tt.destination},
				"departure_date": {tt.departure},
				"end_date":       {tt.end},
				"duration_days":  {tt.duration},
			})

			if rec.Code != http.StatusUnprocessableEntity {
				t.Errorf("status = %d, want 422", rec.Code)
			}
			if n := count(t, database, "trip"); n != 0 {
				t.Errorf("%d trips saved, want 0", n)
			}
			body := rec.Body.String()
			if !strings.Contains(body, html.EscapeString(tt.wantMessage)) {
				t.Errorf("response does not explain the problem %q", tt.wantMessage)
			}
			if tt.departure != "" && !strings.Contains(inputTag(t, body, "departure_date"), `value="`+tt.departure+`"`) {
				t.Errorf("typed departure date %q was not kept", tt.departure)
			}
			if strings.TrimSpace(tt.destination) != "" && !strings.Contains(inputTag(t, body, "destination"), `value="`+tt.destination+`"`) {
				t.Errorf("typed destination %q was not kept", tt.destination)
			}
			if tt.end != "" && !strings.Contains(inputTag(t, body, "end_date"), `value="`+tt.end+`"`) {
				t.Errorf("typed return date %q was not kept", tt.end)
			}
			if !strings.Contains(inputTag(t, body, "duration_days"), `value="`+tt.duration+`"`) {
				t.Errorf("typed duration %q was not kept", tt.duration)
			}
		})
	}
}

// S17: Parent sets the return date instead of the duration.
func TestCreateTripWithReturnDate(t *testing.T) {
	tests := []struct{ name, duration string }{
		{"duration left empty", ""},
		// Without the in-browser syncing the duration could be stale; the
		// return date the parent picked decides.
		{"stale duration", "3"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app, database := newTestApp(t)
			app.now = fixedNow(t, "2026-09-19T09:00:00Z")

			rec := send(t, app, http.MethodPost, "/trips", url.Values{
				"destination":    {"Levi"},
				"departure_date": {"2026-12-20"},
				"end_date":       {"2026-12-27"},
				"duration_days":  {tt.duration},
			})

			if rec.Code != http.StatusSeeOther {
				t.Fatalf("status = %d, want 303", rec.Code)
			}
			var days int
			if err := database.QueryRow("SELECT duration_days FROM trip").Scan(&days); err != nil {
				t.Fatalf("trip was not saved: %v", err)
			}
			if days != 8 {
				t.Errorf("duration = %d days, want 8 (20.–27.12, departure day included)", days)
			}
		})
	}
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

func insertTrip(t *testing.T, database *sql.DB, destination, departure string, days int) int64 {
	t.Helper()
	res, err := database.Exec("INSERT INTO trip (destination, departure_date, duration_days) VALUES (?, ?, ?)", destination, departure, days)
	if err != nil {
		t.Fatalf("insert trip: %v", err)
	}
	id, _ := res.LastInsertId()
	return id
}

func count(t *testing.T, database *sql.DB, table string) int {
	t.Helper()
	var n int
	if err := database.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&n); err != nil {
		t.Fatalf("count %s: %v", table, err)
	}
	return n
}

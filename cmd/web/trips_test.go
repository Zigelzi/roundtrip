package main

import (
	"html"
	"net/http"
	"net/url"
	"strings"
	"testing"
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
	// 13/S1: nights replaced the return date and days (2 nights is 3 days).
	if nights := inputTag(t, body, "nights"); !strings.Contains(nights, `value="2"`) {
		t.Errorf("nights do not default to 2: %s", nights)
	}
}

// S4: Parent creates a trip.
func TestCreateTrip(t *testing.T) {
	app, database := newTestApp(t)
	app.now = fixedNow(t, "2026-09-19T09:00:00Z")

	rec := send(t, app, http.MethodPost, "/trips", url.Values{
		"destination":    {"Parainen"},
		"departure_date": {"2026-09-20"},
		"nights":         {"2"},
		"members":        everyone,
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
		"nights":         {"2"},
		"members":        everyone,
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
		wantMessage string
	}{
		// The return date and days cases moved to 13/S4 when nights replaced them.
		{"empty destination", "  ", "2026-09-25", "Enter a destination"},
		{"no departure date", "Parainen", "", "Choose a departure date"},
		{"departure before today", "Parainen", "2026-09-19", "Departure date can't be in the past"},
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
				"nights":         {"2"},
				"members":        everyone,
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
			if !strings.Contains(inputTag(t, body, "nights"), `value="2"`) {
				t.Errorf("chosen nights were not kept")
			}
		})
	}
}

// S17 (return date instead of the duration) was retired in milestone 13:
// the return date is no longer typed, it follows from the nights (13/S2).

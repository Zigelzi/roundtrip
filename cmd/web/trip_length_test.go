package main

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
)

// Milestone 04 (spec/milestones/04-trip-basics.md), slice 2.

// S11: A trip is at most 14 days. Since milestone 13 the form takes nights,
// so 14 days is 13 nights, and the message is 13/S4's.
func TestTripIsAtMostFourteenDays(t *testing.T) {
	tests := []struct {
		name     string
		nights   string
		wantDays int
		created  bool
	}{
		{"14 days as 13 nights", "13", 14, true},
		{"15 days as 14 nights", "14", 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app, database := newTestApp(t)
			app.now = fixedNow(t, "2026-09-19T09:00:00Z")

			rec := send(t, app, http.MethodPost, "/trips", url.Values{
				"destination":    {"Parainen"},
				"departure_date": {"2027-06-01"},
				"nights":         {tt.nights},
				"members":        everyone,
			})

			if tt.created {
				if rec.Code != http.StatusSeeOther {
					t.Fatalf("status = %d, want 303: a %d-day trip must be created", rec.Code, tt.wantDays)
				}
				var days int
				if err := database.QueryRow("SELECT duration_days FROM trip").Scan(&days); err != nil {
					t.Fatalf("trip was not saved: %v", err)
				}
				if days != tt.wantDays {
					t.Errorf("saved duration = %d, want %d", days, tt.wantDays)
				}
				return
			}

			if rec.Code != http.StatusUnprocessableEntity {
				t.Errorf("status = %d, want 422", rec.Code)
			}
			if n := count(t, database, "trip"); n != 0 {
				t.Errorf("%d trips saved, want 0", n)
			}
			if n := count(t, database, "item"); n != 0 {
				t.Errorf("%d items saved, want 0", n)
			}
			body := rec.Body.String()
			if !strings.Contains(body, "Choose 1 to 13 nights.") {
				t.Errorf("response does not explain the 13-night limit")
			}
			if !strings.Contains(inputTag(t, body, "departure_date"), `value="2027-06-01"`) {
				t.Errorf("typed departure date was not kept")
			}
		})
	}
}

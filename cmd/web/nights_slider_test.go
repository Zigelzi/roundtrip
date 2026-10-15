package main

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
)

// Milestone 13 (spec/milestones/13-nights-slider.md).

// returnHint holds the return date's place until a departure is chosen (13/Q2).
const returnHint = "Pick a departure date to see the return date."

// nightsMessage explains the 1 to 13 nights limit (13/S4).
const nightsMessage = "Choose 1 to 13 nights."

// S1: The new trip form asks for nights.
func TestNewTripFormAsksForNights(t *testing.T) {
	app, _ := newTestApp(t)

	body := get(t, app, "/trips/new")

	inputTag(t, body, "destination")
	inputTag(t, body, "departure_date")
	nights := inputTag(t, body, "nights")
	for _, want := range []string{`type="range"`, `min="1"`, `max="13"`, `value="2"`} {
		if !strings.Contains(nights, want) {
			t.Errorf("nights slider lacks %s: %s", want, nights)
		}
	}
	for _, gone := range []string{`name="end_date"`, `name="duration_days"`} {
		if strings.Contains(body, gone) {
			t.Errorf("form still has the field %s", gone)
		}
	}
	if !strings.Contains(body, "2 nights") {
		t.Errorf("form does not show the starting value as %q", "2 nights")
	}
	if line := returnLine(t, body); line != returnHint {
		t.Errorf("return date line = %q, want the hint %q", line, returnHint)
	}
}

// returnLine returns the text of the line under the slider that shows the
// return date, or the hint while there is no departure.
func returnLine(t *testing.T, body string) string {
	t.Helper()
	p := between(t, body, `id="return-date"`, "</p>")
	return strings.TrimSpace(p[strings.Index(p, ">")+1:])
}

// S2: The form shows the return date.
// The live update while the slider moves is checked on the phone; this is the
// return date the server draws when it already knows the departure and nights.
func TestFormShowsReturnDate(t *testing.T) {
	tests := []struct{ departure, want string }{
		{"2026-12-20", "Sun 27.12.2026"},
		{"2026-12-21", "Mon 28.12.2026"},
		// Across a year, the same day arithmetic as the trip's date range.
		{"2026-12-30", "Wed 6.1.2027"},
	}
	for _, tt := range tests {
		t.Run(tt.departure, func(t *testing.T) {
			app, _ := newTestApp(t)
			app.now = fixedNow(t, "2026-09-19T09:00:00Z")

			rec := send(t, app, http.MethodPost, "/trips", url.Values{
				"destination":    {""},
				"departure_date": {tt.departure},
				"nights":         {"7"},
				"members":        everyone,
			})

			body := rec.Body.String()
			if line := returnLine(t, body); line != "Return "+tt.want {
				t.Errorf("return date line = %q, want %q", line, "Return "+tt.want)
			}
			if !strings.Contains(body, "7 nights") {
				t.Errorf("form does not read %q", "7 nights")
			}
		})
	}
}

// S3: A trip is saved with the nights chosen.
func TestTripSavedWithNights(t *testing.T) {
	app, database := newTestApp(t)
	app.now = fixedNow(t, "2026-09-19T09:00:00Z")

	rec := send(t, app, http.MethodPost, "/trips", url.Values{
		"destination":    {"Levi"},
		"departure_date": {"2026-12-20"},
		"nights":         {"7"},
		"members":        everyone,
	})

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want 303", rec.Code)
	}
	if body := get(t, app, "/"); !strings.Contains(body, "20.–27.12.2026") {
		t.Errorf("trips list does not show the trip as 20.–27.12.2026")
	}
	// Child 1's hair tie is one per day with no spare.
	var hairTies int
	if err := database.QueryRow("SELECT quantity FROM item WHERE name = 'Hair tie'").Scan(&hairTies); err != nil {
		t.Fatalf("read hair ties: %v", err)
	}
	if hairTies != 8 {
		t.Errorf("hair ties = %d, want 8 (7 nights is 8 days)", hairTies)
	}
}

// S4: A trip is 1 to 13 nights.
func TestTripIsOneToThirteenNights(t *testing.T) {
	tests := []struct {
		nights   string
		wantDays int
		created  bool
		// After a refusal the slider comes back within 1 to 13, where the
		// browser would put it anyway, and its words match it.
		slider, words string
	}{
		{"1", 2, true, "", ""},
		{"13", 14, true, "", ""},
		{"0", 0, false, "1", "1 night"},
		{"14", 0, false, "13", "13 nights"},
		{"", 0, false, "2", "2 nights"},
		{"2.5", 0, false, "2", "2 nights"},
	}
	for _, tt := range tests {
		t.Run("nights "+tt.nights, func(t *testing.T) {
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
					t.Fatalf("status = %d, want 303: %s nights must be created", rec.Code, tt.nights)
				}
				var days int
				if err := database.QueryRow("SELECT duration_days FROM trip").Scan(&days); err != nil {
					t.Fatalf("trip was not saved: %v", err)
				}
				if days != tt.wantDays {
					t.Errorf("saved duration = %d days, want %d", days, tt.wantDays)
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
			if !strings.Contains(body, nightsMessage) {
				t.Errorf("response does not say %q", nightsMessage)
			}
			if !strings.Contains(inputTag(t, body, "destination"), `value="Parainen"`) {
				t.Errorf("typed destination was not kept")
			}
			if !strings.Contains(inputTag(t, body, "departure_date"), `value="2027-06-01"`) {
				t.Errorf("typed departure date was not kept")
			}
			slider := inputTag(t, body, "nights")
			if !strings.Contains(slider, `value="`+tt.slider+`"`) || !strings.Contains(slider, `aria-valuetext="`+tt.words+`"`) {
				t.Errorf("slider = %s, want value %s read as %q", slider, tt.slider, tt.words)
			}
			if label := between(t, body, `id="nights-value"`, "</span>"); !strings.HasSuffix(label, ">"+tt.words) {
				t.Errorf("nights label = %q, want %q", label, tt.words)
			}
		})
	}
}

// S5: The form keeps the nights after a mistake.
// S6's spoken value is checked here too: the slider carries "7 nights".
func TestFormKeepsNightsAfterMistake(t *testing.T) {
	app, _ := newTestApp(t)
	app.now = fixedNow(t, "2026-09-19T09:00:00Z")

	rec := send(t, app, http.MethodPost, "/trips", url.Values{
		"destination":    {""},
		"departure_date": {"2026-12-20"},
		"nights":         {"7"},
		"members":        everyone,
	})

	body := rec.Body.String()
	if !strings.Contains(body, "Enter a destination.") {
		t.Errorf("form does not ask for a destination")
	}
	nights := inputTag(t, body, "nights")
	if !strings.Contains(nights, `value="7"`) {
		t.Errorf("slider is not at 7 nights: %s", nights)
	}
	if !strings.Contains(nights, `aria-valuetext="7 nights"`) {
		t.Errorf("slider is not read out as %q: %s", "7 nights", nights)
	}
	if line := returnLine(t, body); line != "Return Sun 27.12.2026" {
		t.Errorf("return date line = %q, want the return date for 7 nights", line)
	}
}

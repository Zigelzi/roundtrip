package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Zigelzi/roundtrip/cmd/web/view"
	"github.com/Zigelzi/roundtrip/internal/db"
)

// dateLayout is how dates travel in forms (<input type="date">) and the database.
const dateLayout = "2006-01-02"

const defaultDurationDays = "3"

func (app *application) handleNewTrip(w http.ResponseWriter, r *http.Request) {
	form := view.TripForm{DurationDays: defaultDurationDays}
	app.render(w, r, http.StatusOK, view.NewTrip(form, app.today().Format(dateLayout)))
}

func (app *application) handleCreateTrip(w http.ResponseWriter, r *http.Request) {
	form := view.TripForm{
		Destination:   strings.TrimSpace(r.PostFormValue("destination")),
		DepartureDate: r.PostFormValue("departure_date"),
		EndDate:       r.PostFormValue("end_date"),
		DurationDays:  strings.TrimSpace(r.PostFormValue("duration_days")),
	}
	today := app.today()
	params, errs := validateTrip(form, today)
	if len(errs) > 0 {
		form.Errors = errs
		app.render(w, r, http.StatusUnprocessableEntity, view.NewTrip(form, today.Format(dateLayout)))
		return
	}
	id, err := app.queries.CreateTrip(r.Context(), params)
	if err != nil {
		app.serverError(w, "create trip", err)
		return
	}
	http.Redirect(w, r, view.TripURL(id), http.StatusSeeOther)
}

// validateTrip checks the submitted trip form. It returns the values to store,
// or one user-facing message per problem.
func validateTrip(form view.TripForm, today time.Time) (db.CreateTripParams, []string) {
	var errs []string
	if form.Destination == "" {
		errs = append(errs, "Enter a destination.")
	}
	departure, err := time.Parse(dateLayout, form.DepartureDate)
	switch {
	case form.DepartureDate == "":
		errs = append(errs, "Choose a departure date.")
	case err != nil:
		errs = append(errs, "Choose a valid departure date.")
	case departure.Before(today):
		errs = append(errs, "Departure date can't be in the past.")
	}
	departureOK := err == nil

	// A return date, when given, decides the duration: it's what the parent
	// picked, while the duration field may be stale if the in-browser syncing
	// didn't run. The departure day counts, so 20.–27.12 is 8 days.
	var days int
	if form.EndDate != "" {
		end, err := time.Parse(dateLayout, form.EndDate)
		switch {
		case err != nil:
			errs = append(errs, "Choose a valid return date.")
		case departureOK && end.Before(departure):
			errs = append(errs, "Return date can't be before the departure date.")
		case departureOK:
			days = int(end.Sub(departure).Hours()/24) + 1
		}
	} else if days, err = strconv.Atoi(form.DurationDays); err != nil || days < 1 {
		errs = append(errs, "Duration must be a whole number of days, at least 1.")
	}
	return db.CreateTripParams{
		Destination:   form.Destination,
		DepartureDate: form.DepartureDate,
		DurationDays:  int64(days),
	}, errs
}

// sortTrips turns stored trips (ordered by departure date) into the front
// page list: upcoming trips first, nearest at the top, then past trips, most
// recent first. A trip departing today counts as upcoming.
func sortTrips(rows []db.Trip, today time.Time) ([]view.Trip, error) {
	var upcoming, past []view.Trip
	for _, row := range rows {
		trip, err := toViewTrip(row)
		if err != nil {
			return nil, err
		}
		if trip.Departure.Before(today) {
			past = append([]view.Trip{trip}, past...)
		} else {
			upcoming = append(upcoming, trip)
		}
	}
	return append(upcoming, past...), nil
}

func toViewTrip(row db.Trip) (view.Trip, error) {
	departure, err := time.Parse(dateLayout, row.DepartureDate)
	if err != nil {
		return view.Trip{}, fmt.Errorf("trip %d has invalid departure date %q: %w", row.ID, row.DepartureDate, err)
	}
	return view.Trip{
		ID:           row.ID,
		Destination:  row.Destination,
		Departure:    departure,
		DurationDays: row.DurationDays,
	}, nil
}

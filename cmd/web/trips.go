package main

import (
	"context"
	"fmt"
	"log"
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

// maxTripDays caps a trip's length, so a typo can't create a trip with
// hundreds of basics that can't be deleted.
const maxTripDays = 14

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
	id, err := app.createTripWithBasics(r.Context(), params)
	if err != nil {
		// Nothing was saved, so the parent can simply try again with the
		// form as they left it.
		log.Printf("failed to create trip: %v", err)
		form.Errors = []string{"The trip could not be created. Please try again."}
		app.render(w, r, http.StatusInternalServerError, view.NewTrip(form, today.Format(dateLayout)))
		return
	}
	http.Redirect(w, r, view.TripURL(id), http.StatusSeeOther)
}

// createTripWithBasics saves the trip and everyone's basics as its items in
// one transaction: either the trip arrives with all of them, or nothing is
// saved. Items are added in ListBasicItems order (owner, then catalogue
// position), which is the order the trip page lists them.
func (app *application) createTripWithBasics(ctx context.Context, params db.CreateTripParams) (int64, error) {
	tx, err := app.database.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback() // a no-op once committed
	qtx := app.queries.WithTx(tx)

	id, err := qtx.CreateTrip(ctx, params)
	if err != nil {
		return 0, fmt.Errorf("save trip: %w", err)
	}
	basics, err := qtx.ListBasicItems(ctx)
	if err != nil {
		return 0, fmt.Errorf("list basics: %w", err)
	}
	for _, b := range basics {
		err := qtx.CreateItem(ctx, db.CreateItemParams{
			TripID:         id,
			FamilyMemberID: b.FamilyMemberID,
			Name:           b.Name,
			NameKey:        itemKey(b.Name),
			Quantity:       b.PerDay*params.DurationDays + b.Fixed,
		})
		if err != nil {
			return 0, fmt.Errorf("add basic %q for member %d: %w", b.Name, b.FamilyMemberID, err)
		}
	}
	return id, tx.Commit()
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
	if days > maxTripDays {
		errs = append(errs, fmt.Sprintf("A trip can be at most %d days.", maxTripDays))
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

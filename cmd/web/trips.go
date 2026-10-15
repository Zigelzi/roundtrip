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

// dateLayout is how dates travel in forms and the database (view.DateLayout).
const dateLayout = view.DateLayout

// choosePerson is the message for a trip with nobody going.
const choosePerson = "Choose at least one person who is going."

func (app *application) handleNewTrip(w http.ResponseWriter, r *http.Request) {
	people, err := app.queries.ListPeople(r.Context())
	if err != nil {
		app.serverError(w, "list people", err)
		return
	}
	form := view.TripForm{People: tripPeople(people, nil)} // the slider starts at its default
	for i := range form.People {
		form.People[i].Ticked = true // everyone goes unless the parent unticks them
	}
	app.render(w, r, http.StatusOK, view.NewTrip(form, app.today().Format(dateLayout)))
}

// tripPeople turns the people into tick boxes, ticking those in going.
func tripPeople(people []db.FamilyMember, going map[int64]bool) []view.TripPerson {
	boxes := make([]view.TripPerson, len(people))
	for i, p := range people {
		boxes[i] = view.TripPerson{ID: p.ID, Name: p.Name, Ticked: going[p.ID]}
	}
	return boxes
}

// goingFromForm reads the ticked boxes. Only ids of real people count: a
// forged or stale value, and the Family bucket's id, are ignored, because the
// Family follows the children. A repeated id counts once.
func goingFromForm(r *http.Request, people []db.FamilyMember) map[int64]bool {
	known := make(map[int64]bool, len(people))
	for _, p := range people {
		known[p.ID] = true
	}
	going := map[int64]bool{}
	for _, v := range r.PostForm["members"] {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil && known[id] {
			going[id] = true
		}
	}
	return going
}

// wantsBasicsOf says whether the basics owned by owner go on a trip with these
// people going. The Family bucket's basics need both children.
func wantsBasicsOf(owner int64, going map[int64]bool) bool {
	if owner != familyBucketID {
		return going[owner]
	}
	for _, child := range childIDs {
		if !going[child] {
			return false
		}
	}
	return true
}

func (app *application) handleCreateTrip(w http.ResponseWriter, r *http.Request) {
	form := view.TripForm{
		Destination:   strings.TrimSpace(r.PostFormValue("destination")),
		DepartureDate: r.PostFormValue("departure_date"),
		Nights:        strings.TrimSpace(r.PostFormValue("nights")),
	}
	people, err := app.queries.ListPeople(r.Context())
	if err != nil {
		app.serverError(w, "list people", err)
		return
	}
	going := goingFromForm(r, people)
	form.People = tripPeople(people, going)
	today := app.today()
	params, errs := validateTrip(form, today)
	if len(going) == 0 {
		errs = append(errs, choosePerson)
	}
	if len(errs) > 0 {
		form.Errors = errs
		app.render(w, r, http.StatusUnprocessableEntity, view.NewTrip(form, today.Format(dateLayout)))
		return
	}
	id, err := app.createTripWithBasics(r.Context(), params, going)
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

// createTripWithBasics saves the trip and the basics of the people going (see
// wantsBasicsOf) as its items in one transaction: either the trip arrives with
// all of them, or nothing is saved. Items are added in ListBasicItems order
// (owner, then catalogue position), which is the order the trip page lists
// them.
func (app *application) createTripWithBasics(ctx context.Context, params db.CreateTripParams, going map[int64]bool) (int64, error) {
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
		if !wantsBasicsOf(b.FamilyMemberID, going) {
			continue
		}
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

	// Trips are stored in days and the departure day counts, so 7 nights,
	// 20.–27.12, is 8 days.
	nights, err := strconv.Atoi(form.Nights)
	if err != nil || nights < 1 || nights > view.MaxTripNights {
		errs = append(errs, fmt.Sprintf("Choose 1 to %d nights.", view.MaxTripNights))
	}
	return db.CreateTripParams{
		Destination:   form.Destination,
		DepartureDate: form.DepartureDate,
		DurationDays:  int64(nights + 1),
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

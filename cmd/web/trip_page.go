package main

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/Zigelzi/roundtrip/cmd/web/view"
)

const defaultQuantity = "1"

// errTripNotFound marks a trip id that doesn't exist (or isn't a number).
var errTripNotFound = errors.New("trip not found")

func (app *application) handleTrip(w http.ResponseWriter, r *http.Request) {
	tripID, err := tripIDFromPath(r)
	if err != nil {
		app.tripNotFound(w, r)
		return
	}
	page, err := app.tripPage(r.Context(), tripID)
	if errors.Is(err, errTripNotFound) {
		app.tripNotFound(w, r)
		return
	}
	if err != nil {
		app.serverError(w, "load trip page", err)
		return
	}
	app.render(w, r, http.StatusOK, view.TripPage(page))
}

// tripPage loads everything the trip page shows: the trip, each family
// member's items with an empty add form, and item names to suggest.
func (app *application) tripPage(ctx context.Context, tripID int64) (view.TripPageData, error) {
	row, err := app.queries.GetTrip(ctx, tripID)
	if errors.Is(err, sql.ErrNoRows) {
		return view.TripPageData{}, errTripNotFound
	}
	if err != nil {
		return view.TripPageData{}, err
	}
	trip, err := toViewTrip(row)
	if err != nil {
		return view.TripPageData{}, err
	}
	members, err := app.queries.ListFamilyMembers(ctx)
	if err != nil {
		return view.TripPageData{}, err
	}
	items, err := app.queries.ListTripItems(ctx, tripID)
	if err != nil {
		return view.TripPageData{}, err
	}

	suggestions, err := app.queries.ListItemNames(ctx)
	if err != nil {
		return view.TripPageData{}, err
	}

	page := view.TripPageData{Trip: trip, Suggestions: suggestions}
	for _, m := range members {
		section := view.MemberItems{ID: m.ID, Name: m.Name, Form: view.ItemForm{Quantity: defaultQuantity}}
		for _, it := range items {
			if it.FamilyMemberID == m.ID {
				section.Items = append(section.Items, view.Item{ID: it.ID, Name: it.Name, Quantity: it.Quantity})
			}
		}
		page.Members = append(page.Members, section)
	}
	return page, nil
}

func tripIDFromPath(r *http.Request) (int64, error) {
	return strconv.ParseInt(r.PathValue("id"), 10, 64)
}

func (app *application) tripNotFound(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, http.StatusNotFound, view.NotFound("Trip not found"))
}

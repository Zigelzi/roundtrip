package main

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/Zigelzi/roundtrip/cmd/web/view"
	"github.com/Zigelzi/roundtrip/internal/db"
)

// Item lifecycle values stored in item.status (see spec/domain-model.md).
// Milestone 02 implements these two; prepared / needs_buying / bought follow.
const (
	statusPlanned = "planned"
	statusPacked  = "packed"
)

func (app *application) handlePackPage(w http.ResponseWriter, r *http.Request) {
	tripID, err := tripIDFromPath(r)
	if err != nil {
		app.tripNotFound(w, r)
		return
	}
	page, err := app.packPage(r.Context(), tripID, memberFilter(r))
	if errors.Is(err, errTripNotFound) {
		app.tripNotFound(w, r)
		return
	}
	if err != nil {
		app.serverError(w, "load packing page", err)
		return
	}
	app.render(w, r, http.StatusOK, view.PackPage(page))
}

func (app *application) handlePackItem(w http.ResponseWriter, r *http.Request) {
	app.setItemStatus(w, r, statusPacked)
}

func (app *application) handleUnpackItem(w http.ResponseWriter, r *http.Request) {
	app.setItemStatus(w, r, statusPlanned)
}

// setItemStatus ticks an item packed or back again. The list to re-render is
// the item's own member; the path's member is the fallback for when the item
// has since been removed and there is no row left to ask.
func (app *application) setItemStatus(w http.ResponseWriter, r *http.Request, status string) {
	tripID, err := tripIDFromPath(r)
	if err != nil {
		app.tripNotFound(w, r)
		return
	}
	memberID, err := strconv.ParseInt(r.PathValue("memberID"), 10, 64)
	if err != nil {
		app.tripNotFound(w, r)
		return
	}
	itemID, err := strconv.ParseInt(r.PathValue("itemID"), 10, 64)
	if err != nil {
		app.tripNotFound(w, r)
		return
	}

	owner, err := app.queries.SetItemStatus(r.Context(), db.SetItemStatusParams{Status: status, ID: itemID, TripID: tripID})
	switch {
	case errors.Is(err, sql.ErrNoRows):
		// The item is gone — removed on the trip page by the other parent, or
		// an id from another trip. Nothing is saved; the parent gets the list
		// as it now stands and carries on packing (S10). With no row to ask,
		// the path says which list that is.
		owner = memberID
	case err != nil:
		app.serverError(w, "set item status", err)
		return
	}
	// Re-render the item's real owner, not whoever the path claims: the
	// changed list is the one that has to come back fresh.
	app.packingChanged(w, r, tripID, owner, memberFilter(r))
}

// packingChanged answers a tap: htmx gets the changed fragments, a plain
// form post is redirected back to the packing page.
func (app *application) packingChanged(w http.ResponseWriter, r *http.Request, tripID, memberID, filter int64) {
	if !isHTMX(r) {
		http.Redirect(w, r, view.WithFilter(view.PackURL(tripID), filter), http.StatusSeeOther)
		return
	}
	page, err := app.packPage(r.Context(), tripID, filter)
	if err != nil {
		app.serverError(w, "load packing page", err)
		return
	}
	i := packMemberIndex(page.Members, memberID)
	if i < 0 {
		// Only reachable by hand-editing the member out of the URL. Nothing
		// was written, so this is a bad request, not a server fault.
		app.render(w, r, http.StatusNotFound, view.NotFound("Family member not found"))
		return
	}
	app.render(w, r, http.StatusOK, view.PackChanged(page, page.Members[i]))
}

func packMemberIndex(members []view.PackMember, id int64) int {
	for i, m := range members {
		if m.ID == id {
			return i
		}
	}
	return -1
}

// packPage loads the trip and splits each member's items into what is still
// to pack and what is already in the bag.
func (app *application) packPage(ctx context.Context, tripID, filter int64) (view.PackPageData, error) {
	row, err := app.queries.GetTrip(ctx, tripID)
	if errors.Is(err, sql.ErrNoRows) {
		return view.PackPageData{}, errTripNotFound
	}
	if err != nil {
		return view.PackPageData{}, err
	}
	trip, err := toViewTrip(row)
	if err != nil {
		return view.PackPageData{}, err
	}
	members, err := app.queries.ListFamilyMembers(ctx)
	if err != nil {
		return view.PackPageData{}, err
	}
	items, err := app.queries.ListTripItems(ctx, tripID)
	if err != nil {
		return view.PackPageData{}, err
	}

	page := view.PackPageData{Trip: trip, Filter: filter}
	for _, m := range members {
		section := view.PackMember{ID: m.ID, Name: m.Name}
		for _, it := range items {
			if it.FamilyMemberID != m.ID {
				continue
			}
			item := view.Item{ID: it.ID, Name: it.Name, Quantity: it.Quantity}
			if it.Status == statusPacked {
				section.Packed = append(section.Packed, item)
			} else {
				section.ToPack = append(section.ToPack, item)
			}
		}
		page.TripTotal += len(section.Packed) + len(section.ToPack)
		page.Members = append(page.Members, section)
	}
	// A filter naming nobody (a stale or hand-edited link) would otherwise
	// render a page with no sections at all, which reads as an empty trip.
	if page.Filter != 0 && packMemberIndex(page.Members, page.Filter) < 0 {
		page.Filter = 0
	}
	// The progress counts what is on screen: narrowed to one member, it is
	// that member's own progress (S5).
	for _, m := range page.Shown() {
		page.Packed += len(m.Packed)
		page.Total += len(m.Packed) + len(m.ToPack)
	}
	return page, nil
}

// memberFilter reads the family member the page is narrowed to. Anything
// unparseable means the whole family, so a mangled link still shows a page.
func memberFilter(r *http.Request) int64 {
	id, err := strconv.ParseInt(r.URL.Query().Get("member"), 10, 64)
	if err != nil || id < 1 {
		return 0
	}
	return id
}

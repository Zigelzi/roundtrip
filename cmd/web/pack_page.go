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
// Milestone 08 implements these three; needs_buying / bought follow.
const (
	statusPlanned  = "planned"
	statusPrepared = "prepared"
	statusPacked   = "packed"
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

// Each handler names its step: the item moves only if it is in the state the
// step starts from (08 Scope 6).

func (app *application) handlePrepareItem(w http.ResponseWriter, r *http.Request) {
	app.moveItem(w, r, statusPlanned, statusPrepared)
}

func (app *application) handlePackItem(w http.ResponseWriter, r *http.Request) {
	app.moveItem(w, r, statusPrepared, statusPacked)
}

func (app *application) handleUnpackItem(w http.ResponseWriter, r *http.Request) {
	app.moveItem(w, r, statusPacked, statusPrepared)
}

func (app *application) handleUnprepareItem(w http.ResponseWriter, r *http.Request) {
	app.moveItem(w, r, statusPrepared, statusPlanned)
}

// moveItem takes an item one step, from one status to the next. The list to
// re-render is the item's own member; the path's member is the fallback for
// when the item has since been removed and there is no row left to ask.
func (app *application) moveItem(w http.ResponseWriter, r *http.Request, from, to string) {
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

	owner, err := app.queries.MoveItemStatus(r.Context(), db.MoveItemStatusParams{
		ToStatus: to, ID: itemID, TripID: tripID, FromStatus: from,
	})
	if errors.Is(err, sql.ErrNoRows) {
		// Nothing moved. Either the item is in another state (the other
		// parent got there first: S8, S9, S10) or it is gone (02/S10).
		// Either way the reply shows the list as it really is, so ask whose
		// list that is.
		owner, err = app.queries.GetItemOwner(r.Context(), db.GetItemOwnerParams{ID: itemID, TripID: tripID})
		if errors.Is(err, sql.ErrNoRows) {
			// Removed on the trip page by the other parent, or an id from
			// another trip. With no row to ask, the path says which list
			// the parent was looking at.
			owner, err = memberID, nil
		}
	}
	if err != nil {
		app.serverError(w, "move item status", err)
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
	owner := page.Members[i]
	app.render(w, r, http.StatusOK, view.PackChanged(page, owner, wholePartSwap(r, len(owner.Packed))))
}

// wholePartSwap says whether a member's packed part has to be sent whole.
// Every tap form says whether its page shows the part (packed_shown). Only
// when it does, and the member still has something packed, are the count
// and the list swapped alone, so an open part stays open (02/S3, 08/S5,
// also for a repeated tap that changes nothing). Otherwise the whole wrapper
// goes: the part appears or goes, or the page was out of date and has
// nowhere to put a count and a list (S9, S10, or the other parent packed
// this member's first item). A tap without the field gets the whole part:
// it may close an open part, but it never leaves the page wrong.
func wholePartSwap(r *http.Request, packedCount int) bool {
	return r.PostFormValue("packed_shown") != "1" || packedCount == 0
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
// to pack (planned and prepared, in the order they were added) and what is
// already in the bag (their packed part).
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

	page := view.PackPageData{Trip: trip, Filter: filter, FamilyBucketID: familyBucketID}
	for _, m := range members {
		section := view.PackMember{ID: m.ID, Name: m.Name}
		for _, it := range items {
			if it.FamilyMemberID != m.ID {
				continue
			}
			item := view.Item{ID: it.ID, Name: it.Name, Quantity: it.Quantity, Prepared: it.Status == statusPrepared}
			if it.Status == statusPacked {
				section.Packed = append(section.Packed, item)
			} else {
				section.ToPack = append(section.ToPack, item)
			}
		}
		page.TripTotal += len(section.Packed) + len(section.ToPack)
		page.TripPacked += len(section.Packed)
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

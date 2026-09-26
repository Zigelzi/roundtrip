package main

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Zigelzi/roundtrip/cmd/web/view"
	"github.com/Zigelzi/roundtrip/internal/db"
)

func (app *application) handleAddItem(w http.ResponseWriter, r *http.Request) {
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
	// The trip page data is loaded up front: it holds the member's current
	// items for the duplicate check, and is re-rendered if the input is invalid.
	// A row being edited stays in edit mode (the form carries ?edit=) in the
	// htmx reply and when the input is invalid. A plain successful add
	// redirects to the member's section and closes it; accepted, since it
	// only happens without JavaScript.
	page, err := app.tripPage(r.Context(), tripID, view.ItemEdit{ItemID: editParam(r)})
	if errors.Is(err, errTripNotFound) {
		app.tripNotFound(w, r)
		return
	}
	if err != nil {
		app.serverError(w, "load trip page", err)
		return
	}
	i := memberIndex(page.Members, memberID)
	if i < 0 {
		app.render(w, r, http.StatusNotFound, view.NotFound("Family member not found"))
		return
	}

	form := view.ItemForm{
		Name:     strings.TrimSpace(r.PostFormValue("name")),
		Quantity: strings.TrimSpace(r.PostFormValue("quantity")),
	}
	quantity, errs := validateItem(form, page.Members[i])
	if len(errs) > 0 {
		form.Errors = errs
		page.Members[i].Form = form
		if isHTMX(r) {
			app.render(w, r, http.StatusUnprocessableEntity, view.MemberItemList(tripID, page.Members[i], page.Edit))
			return
		}
		app.render(w, r, http.StatusUnprocessableEntity, view.TripPage(page))
		return
	}
	err = app.queries.CreateItem(r.Context(), db.CreateItemParams{
		TripID:         tripID,
		FamilyMemberID: memberID,
		Name:           form.Name,
		NameKey:        itemKey(form.Name),
		Quantity:       quantity,
	})
	// A unique violation is a double-tapped Add: the first request already
	// saved the item, so treat it as success.
	if err != nil && !db.IsUniqueViolation(err) {
		app.serverError(w, "create item", err)
		return
	}
	app.itemsChanged(w, r, tripID, memberID)
}

// itemsChanged answers a successful add or remove: htmx gets the member's
// fresh item list to swap in; a plain form post is redirected back to the
// member's section.
func (app *application) itemsChanged(w http.ResponseWriter, r *http.Request, tripID, memberID int64) {
	if !isHTMX(r) {
		http.Redirect(w, r, view.MemberURL(tripID, memberID), http.StatusSeeOther)
		return
	}
	page, err := app.tripPage(r.Context(), tripID, view.ItemEdit{ItemID: editParam(r)})
	if err != nil {
		app.serverError(w, "load trip page", err)
		return
	}
	i := memberIndex(page.Members, memberID)
	if i < 0 {
		app.serverError(w, "find member", fmt.Errorf("member %d missing from trip %d", memberID, tripID))
		return
	}
	app.render(w, r, http.StatusOK, view.MemberItemList(tripID, page.Members[i], page.Edit))
}

// isHTMX reports whether the request came from htmx rather than a plain form.
func isHTMX(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

// itemKey is how item names compare: case-insensitive, Ä/ä included (Go's
// ToLower is Unicode-aware). Stored as item.name_key.
func itemKey(name string) string {
	return strings.ToLower(name)
}

// validateItem checks a member's add-item form. It returns the quantity to
// store, or one user-facing message per problem.
func validateItem(form view.ItemForm, member view.MemberItems) (int64, []string) {
	var errs []string
	if form.Name == "" {
		errs = append(errs, "Enter an item name.")
	} else {
		for _, it := range member.Items {
			if itemKey(it.Name) == itemKey(form.Name) {
				errs = append(errs, fmt.Sprintf("%s already has %s.", member.Name, it.Name))
				break
			}
		}
	}
	quantity, ok := parseQuantity(form.Quantity)
	if !ok {
		errs = append(errs, badQuantity)
	}
	return quantity, errs
}

// badQuantity is the one message for any quantity that isn't a whole number
// of at least 1, shared by adding an item and changing its quantity.
const badQuantity = "Quantity must be a whole number, at least 1."

// parseQuantity reads a typed quantity; ok is false unless it is a whole
// number of at least 1.
func parseQuantity(s string) (quantity int64, ok bool) {
	quantity, err := strconv.ParseInt(s, 10, 64)
	return quantity, err == nil && quantity >= 1
}

// handleChangeQuantity saves an item's new quantity (S8). Only the quantity
// changes; a packed item stays packed (R2). An invalid value re-renders the
// trip page with the row still in edit mode (S9).
func (app *application) handleChangeQuantity(w http.ResponseWriter, r *http.Request) {
	tripID, err := tripIDFromPath(r)
	if err != nil {
		app.tripNotFound(w, r)
		return
	}
	itemID, err := strconv.ParseInt(r.PathValue("itemID"), 10, 64)
	if err != nil {
		app.itemNotFound(w, r)
		return
	}
	input := strings.TrimSpace(r.PostFormValue("quantity"))
	quantity, ok := parseQuantity(input)
	if !ok {
		edit := view.ItemEdit{ItemID: itemID, Quantity: input, Errors: []string{badQuantity}}
		page, err := app.tripPage(r.Context(), tripID, edit)
		if errors.Is(err, errTripNotFound) {
			app.tripNotFound(w, r)
			return
		}
		if err != nil {
			app.serverError(w, "load trip page", err)
			return
		}
		// tripPage drops an edit for an item that is not on this trip.
		if page.Edit.ItemID == 0 {
			app.itemNotFound(w, r)
			return
		}
		app.render(w, r, http.StatusUnprocessableEntity, view.TripPage(page))
		return
	}
	_, err = app.queries.UpdateItemQuantity(r.Context(), db.UpdateItemQuantityParams{
		Quantity: quantity,
		ID:       itemID,
		TripID:   tripID,
	})
	if errors.Is(err, sql.ErrNoRows) {
		app.itemNotFound(w, r)
		return
	}
	if err != nil {
		app.serverError(w, "update item quantity", err)
		return
	}
	http.Redirect(w, r, view.ItemURL(tripID, itemID), http.StatusSeeOther)
}

func (app *application) itemNotFound(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, http.StatusNotFound, view.NotFound("Item not found"))
}

func (app *application) handleRemoveItem(w http.ResponseWriter, r *http.Request) {
	tripID, err := tripIDFromPath(r)
	if err != nil {
		app.tripNotFound(w, r)
		return
	}
	itemID, err := strconv.ParseInt(r.PathValue("itemID"), 10, 64)
	if err != nil {
		app.tripNotFound(w, r)
		return
	}
	memberID, err := app.queries.DeleteItem(r.Context(), db.DeleteItemParams{ID: itemID, TripID: tripID})
	if errors.Is(err, sql.ErrNoRows) {
		// Already gone (e.g. a double tap). htmx: 204 leaves the page as is,
		// since the first tap already updated the list.
		if isHTMX(r) {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		http.Redirect(w, r, view.TripURL(tripID), http.StatusSeeOther)
		return
	}
	if err != nil {
		app.serverError(w, "delete item", err)
		return
	}
	app.itemsChanged(w, r, tripID, memberID)
}

func memberIndex(members []view.MemberItems, id int64) int {
	for i, m := range members {
		if m.ID == id {
			return i
		}
	}
	return -1
}

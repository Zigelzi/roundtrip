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
	quantity, errs := validateItem(form, page.Members[i], 0)
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
// fresh item list to swap in, and the packing bar with its new count (12/S4);
// a plain form post is redirected back to the member's section.
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
	app.render(w, r, http.StatusOK, view.ItemsChanged(page, page.Members[i]))
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

// validateItem checks a member's add-item or edit-item form. It returns the
// quantity to store, or one user-facing message per problem. ignoreID is the
// item being edited, which is not a duplicate of itself (15/S3); adding
// passes 0.
func validateItem(form view.ItemForm, member view.MemberItems, ignoreID int64) (int64, []string) {
	var errs []string
	if form.Name == "" {
		errs = append(errs, "Enter an item name.")
	} else {
		for _, it := range member.Items {
			if it.ID != ignoreID && itemKey(it.Name) == itemKey(form.Name) {
				errs = append(errs, alreadyHasItem(member.Name, it.Name))
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

// alreadyHasItem is the message for a name the member already has.
func alreadyHasItem(member, name string) string {
	return fmt.Sprintf("%s already has %s.", member, name)
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

// handleSaveItem saves an item's new name and quantity (04/S8, 15). Only
// those change; a packed item stays packed (04/R2). A save with no name field
// at all, from a page opened before names were editable, keeps the name
// (15/S13). An invalid value re-renders the trip page with the row still in
// edit mode and what was typed in its fields (04/S9, 15/S4, S5).
func (app *application) handleSaveItem(w http.ResponseWriter, r *http.Request) {
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
	typedName := strings.TrimSpace(r.PostFormValue("name"))
	_, hasName := r.PostForm["name"]
	typedQuantity := strings.TrimSpace(r.PostFormValue("quantity"))

	// The page is loaded first: it says whether the item is on this trip, and
	// holds the owner's other items for the duplicate check. tripPage drops an
	// edit for an item that is not on this trip.
	edit := view.ItemEdit{ItemID: itemID, Name: typedName, Quantity: typedQuantity, Submitted: true}
	page, err := app.tripPage(r.Context(), tripID, edit)
	if errors.Is(err, errTripNotFound) {
		app.tripNotFound(w, r)
		return
	}
	if err != nil {
		app.serverError(w, "load trip page", err)
		return
	}
	if page.Edit.ItemID == 0 {
		app.itemNotFound(w, r)
		return
	}
	owner, current := findItem(page.Members, itemID)
	if !hasName {
		typedName = current.Name
		page.Edit.Name = typedName
	}

	quantity, errs := validateItem(view.ItemForm{Name: typedName, Quantity: typedQuantity}, page.Members[owner], itemID)
	if len(errs) > 0 {
		app.refuseItemSave(w, r, page, errs)
		return
	}
	_, err = app.queries.UpdateItem(r.Context(), db.UpdateItemParams{
		Name:     typedName,
		NameKey:  itemKey(typedName),
		Quantity: quantity,
		ID:       itemID,
		TripID:   tripID,
	})
	// A unique violation is the other parent adding the same name a moment
	// ago: refuse it like any duplicate. Unlike adding, it is not a
	// double tap, since this save did not happen.
	if db.IsUniqueViolation(err) {
		app.refuseItemSave(w, r, page, []string{alreadyHasItem(page.Members[owner].Name, typedName)})
		return
	}
	if errors.Is(err, sql.ErrNoRows) {
		app.itemNotFound(w, r)
		return
	}
	if err != nil {
		app.serverError(w, "update item", err)
		return
	}
	http.Redirect(w, r, view.ItemURL(tripID, itemID), http.StatusSeeOther)
}

// refuseItemSave answers a refused save with the trip page, the row still
// open with what was typed.
func (app *application) refuseItemSave(w http.ResponseWriter, r *http.Request, page view.TripPageData, errs []string) {
	page.Edit.Errors = errs
	app.render(w, r, http.StatusUnprocessableEntity, view.TripPage(page))
}

// findItem is the index of the member who owns item id, and the item.
func findItem(members []view.MemberItems, id int64) (int, view.Item) {
	for i, m := range members {
		for _, it := range m.Items {
			if it.ID == id {
				return i, it
			}
		}
	}
	return -1, view.Item{}
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

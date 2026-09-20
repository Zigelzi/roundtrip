package main

import (
	"fmt"
	"html"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/Zigelzi/roundtrip/internal/db"
)

// Milestone 01 (spec/milestones/01-packing-list.md), item scenarios S9–S14.

// S9: Parent adds an item to a member.
func TestAddItem(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	parent := memberID(t, database, "Parent 1")

	rec := addItem(t, app, tripID, parent, "Underwear", "4")

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want 303 back to the trip", rec.Code)
	}
	if loc, want := rec.Header().Get("Location"), fmt.Sprintf("/trips/%d#member-%d", tripID, parent); loc != want {
		t.Errorf("redirect to %q, want %q (back to Parent 1's section)", loc, want)
	}
	section := memberSection(t, get(t, app, tripURLFor(tripID)), "Parent 1")
	if !strings.Contains(section, "Underwear") || !strings.Contains(section, "4 pcs") {
		t.Errorf("Parent 1's section does not list Underwear, 4 pcs:\n%s", section)
	}
}

// S10: Earlier item names are suggested.
func TestEarlierItemNamesSuggested(t *testing.T) {
	app, database := newTestApp(t)
	earlier := insertTrip(t, database, "Levi", "2026-03-01", 5)
	insertItem(t, database, earlier, "Parent 2", "Sunscreen", 1)
	insertItem(t, database, earlier, "Parent 1", "sunscreen", 1)
	newTrip := insertTrip(t, database, "Parainen", "2026-09-20", 3)

	body := get(t, app, tripURLFor(newTrip))

	if !strings.Contains(body, `<option value="Sunscreen">`) {
		t.Errorf("trip page does not suggest the earlier item name Sunscreen")
	}
	if n := strings.Count(strings.ToLower(body), `<option value="sunscreen">`); n != 1 {
		t.Errorf("Sunscreen is suggested %d times, want once regardless of case", n)
	}
	if name := inputTag(t, memberSection(t, body, "Parent 1"), "name"); !strings.Contains(name, `list="item-names"`) {
		t.Errorf("item name field is not linked to the suggestions: %s", name)
	}
}

// S11: The same item cannot be added twice to a member.
func TestAddDuplicateItemRejected(t *testing.T) {
	tests := []struct{ existing, added string }{
		{"Underwear", "underwear"},
		{"Ämpäri", "ämpäri"},         // case-insensitive beyond ASCII
		{"Underwear", " Underwear "}, // surrounding spaces don't make it new
	}
	for _, tt := range tests {
		t.Run(tt.added, func(t *testing.T) {
			app, database := newTestApp(t)
			tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
			insertItem(t, database, tripID, "Parent 1", tt.existing, 4)

			rec := addItem(t, app, tripID, memberID(t, database, "Parent 1"), tt.added, "1")

			if rec.Code != http.StatusUnprocessableEntity {
				t.Errorf("status = %d, want 422", rec.Code)
			}
			if n := count(t, database, "item"); n != 1 {
				t.Errorf("%d items saved, want 1", n)
			}
			section := memberSection(t, rec.Body.String(), "Parent 1")
			if !strings.Contains(section, "Parent 1 already has") {
				t.Errorf("Parent 1's section does not say Parent 1 already has it:\n%s", section)
			}
		})
	}
}

// S12: Different members can have the same item.
func TestSameItemForDifferentMembers(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	insertItem(t, database, tripID, "Parent 1", "Underwear", 4)

	rec := addItem(t, app, tripID, memberID(t, database, "Parent 2"), "Underwear", "4")

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want 303", rec.Code)
	}
	body := get(t, app, tripURLFor(tripID))
	for _, member := range []string{"Parent 1", "Parent 2"} {
		if !strings.Contains(memberSection(t, body, member), "Underwear") {
			t.Errorf("%s's section does not list Underwear", member)
		}
	}
}

// S13: Invalid item input is rejected.
func TestAddItemRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name, itemName, quantity, wantMessage string
	}{
		{"empty name", "  ", "1", "Enter an item name"},
		{"quantity below 1", "Socks", "0", "Quantity must be a whole number, at least 1"},
		{"quantity not a number", "Socks", "many", "Quantity must be a whole number, at least 1"},
		{"quantity not whole", "Socks", "1.5", "Quantity must be a whole number, at least 1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app, database := newTestApp(t)
			tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)

			rec := addItem(t, app, tripID, memberID(t, database, "Parent 1"), tt.itemName, tt.quantity)

			if rec.Code != http.StatusUnprocessableEntity {
				t.Errorf("status = %d, want 422", rec.Code)
			}
			if n := count(t, database, "item"); n != 0 {
				t.Errorf("%d items saved, want 0", n)
			}
			section := memberSection(t, rec.Body.String(), "Parent 1")
			if !strings.Contains(section, html.EscapeString(tt.wantMessage)) {
				t.Errorf("Parent 1's section does not explain %q", tt.wantMessage)
			}
			// The form posts to a URL ending in Parent 1's anchor, so the page
			// showing the error scrolls to that section rather than the top.
			if want := fmt.Sprintf(`#member-%d"`, memberID(t, database, "Parent 1")); !strings.Contains(section, want) {
				t.Errorf("add form action does not keep the section anchor %s", want)
			}
		})
	}
}

// Adding to a trip or member that doesn't exist is a 404, not a server error.
func TestAddItemToMissingTripOrMember(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)

	for _, path := range []string{
		fmt.Sprintf("/trips/999/members/%d/items", memberID(t, database, "Parent 1")),
		fmt.Sprintf("/trips/%d/members/999/items", tripID),
	} {
		rec := send(t, app, http.MethodPost, path, url.Values{"name": {"Socks"}, "quantity": {"1"}})
		if rec.Code != http.StatusNotFound {
			t.Errorf("POST %s: status = %d, want 404", path, rec.Code)
		}
	}
	if n := count(t, database, "item"); n != 0 {
		t.Errorf("%d items saved, want 0", n)
	}
}

// The database itself refuses a second copy of the same item for a member
// (e.g. a double-tapped Add racing past the handler's check).
func TestDatabaseRefusesDuplicateItem(t *testing.T) {
	_, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	insertItem(t, database, tripID, "Parent 1", "Ämpäri", 1)

	_, err := database.Exec("INSERT INTO item (trip_id, family_member_id, name, name_key, quantity) VALUES (?, ?, ?, ?, ?)",
		tripID, memberID(t, database, "Parent 1"), "ämpäri", "ämpäri", 1)

	if !db.IsUniqueViolation(err) {
		t.Errorf("second Ämpäri for Parent 1: err = %v, want a unique-constraint violation", err)
	}
}

// S14: Parent removes an item from a member.
func TestRemoveItem(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	parentsItem := insertItem(t, database, tripID, "Parent 1", "Underwear", 4)
	insertItem(t, database, tripID, "Parent 2", "Underwear", 4)

	rec := send(t, app, http.MethodPost, fmt.Sprintf("/trips/%d/items/%d/delete", tripID, parentsItem), nil)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want 303", rec.Code)
	}
	if loc, want := rec.Header().Get("Location"), fmt.Sprintf("/trips/%d#member-%d", tripID, memberID(t, database, "Parent 1")); loc != want {
		t.Errorf("redirect to %q, want %q (back to Parent 1's section)", loc, want)
	}
	body := get(t, app, tripURLFor(tripID))
	if strings.Contains(memberSection(t, body, "Parent 1"), "Underwear") {
		t.Errorf("Parent 1's section still lists Underwear")
	}
	if !strings.Contains(memberSection(t, body, "Parent 2"), "Underwear") {
		t.Errorf("Parent 2's Underwear was removed too")
	}
}

// Removing an item through another trip's URL must not delete it.
func TestRemoveItemOnlyFromItsOwnTrip(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	other := insertTrip(t, database, "Levi", "2026-12-01", 5)
	item := insertItem(t, database, tripID, "Parent 1", "Underwear", 4)

	send(t, app, http.MethodPost, fmt.Sprintf("/trips/%d/items/%d/delete", other, item), nil)

	if n := count(t, database, "item"); n != 1 {
		t.Errorf("item was deleted through another trip's URL")
	}
}

func addItem(t *testing.T, app *application, tripID, memberID int64, name, quantity string) *httptest.ResponseRecorder {
	t.Helper()
	return send(t, app, http.MethodPost, fmt.Sprintf("/trips/%d/members/%d/items", tripID, memberID),
		url.Values{"name": {name}, "quantity": {quantity}})
}

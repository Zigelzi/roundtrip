package main

import (
	"database/sql"
	"html"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// Milestone 15 (spec/milestones/15-edit-item-name.md): editing an item's name
// from the trip page's edit row.
//
// Page contract these tests rely on:
//   - The open row (?edit=ITEM_ID) has an input aria-label="Name of NAME"
//     named "name" holding the item's name, next to the quantity input from
//     milestone 04, in one form that posts to the item's save URL.
//   - A refused save is a 422 of the whole trip page with the row still open
//     and the typed values in its fields.

func saveItem(t *testing.T, app *application, tripID, itemID int64, name, quantity string) *httptest.ResponseRecorder {
	t.Helper()
	return send(t, app, http.MethodPost, quantityURL(tripID, itemID), url.Values{"name": {name}, "quantity": {quantity}})
}

func itemRow(t *testing.T, database *sql.DB, itemID int64) (name string, quantity int) {
	t.Helper()
	if err := database.QueryRow("SELECT name, quantity FROM item WHERE id = ?", itemID).Scan(&name, &quantity); err != nil {
		t.Fatalf("read item %d: %v", itemID, err)
	}
	return name, quantity
}

// S1: the open row shows the name to edit.
func TestOpenRowShowsName(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	item := insertItem(t, database, tripID, "Parent 1", "Sunscren", 2)

	section := memberSection(t, get(t, app, editURL(tripID, item)), "Parent 1")

	if tag := inputTagLabelled(t, section, "Name of Sunscren"); !strings.Contains(tag, `value="Sunscren"`) {
		t.Errorf("name field does not hold the item's name: %s", tag)
	}
	if tag := inputTagLabelled(t, section, "New quantity for Sunscren"); !strings.Contains(tag, `value="2"`) {
		t.Errorf("quantity field does not start at 2: %s", tag)
	}
}

// S2, S3, S6: renaming saves the new name and keeps the quantity.
func TestRenameItem(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	typo := insertItem(t, database, tripID, "Parent 1", "Sunscren", 2)
	lower := insertItem(t, database, tripID, "Parent 1", "hat", 1)
	insertItem(t, database, tripID, "Parent 2", "Socks", 1)
	other := insertItem(t, database, tripID, "Parent 2", "Sunscren", 1)

	for _, c := range []struct {
		id   int64
		to   string
		want string
	}{
		{typo, "Sunscreen", "Sunscreen"},
		{lower, "Hat", "Hat"},
		// The same name on another person is fine.
		{other, "Sunscreen", "Sunscreen"},
	} {
		rec := saveItem(t, app, tripID, c.id, c.to, "2")
		if rec.Code != http.StatusSeeOther {
			t.Fatalf("rename to %q: status = %d, want 303", c.to, rec.Code)
		}
		if loc := rec.Header().Get("Location"); !strings.HasPrefix(loc, tripURLFor(tripID)) {
			t.Errorf("rename redirects to %q, want the trip page", loc)
		}
		if name, _ := itemRow(t, database, c.id); name != c.want {
			t.Errorf("name = %q, want %q", name, c.want)
		}
	}

	if _, q := itemRow(t, database, typo); q != 2 {
		t.Errorf("quantity = %d, want it saved as 2", q)
	}
	section := memberSection(t, get(t, app, tripURLFor(tripID)), "Parent 1")
	if !strings.Contains(section, "Sunscreen") || strings.Contains(section, "Sunscren") {
		t.Errorf("Parent 1's list does not show the fixed name only")
	}
}

// S4, S5, S10, S11: refused saves change nothing and keep the row open.
func TestRefusedRename(t *testing.T) {
	for _, c := range []struct {
		name, typed, quantity string
		wantMessages          []string
		wantNameField         string
	}{
		{"duplicate", "sunscreen", "2", []string{"Parent 1 already has Sunscreen."}, "sunscreen"},
		{"empty", "", "2", []string{"Enter an item name."}, ""},
		{"bad quantity keeps typed name", "Sunscreen2", "0", []string{"Quantity must be a whole number, at least 1."}, "Sunscreen2"},
		{"both", "sunscreen", "0", []string{"Parent 1 already has Sunscreen.", "Quantity must be a whole number, at least 1."}, "sunscreen"},
	} {
		t.Run(c.name, func(t *testing.T) {
			app, database := newTestApp(t)
			tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
			insertItem(t, database, tripID, "Parent 1", "Sunscreen", 1)
			typo := insertItem(t, database, tripID, "Parent 1", "Sunscren", 2)

			rec := saveItem(t, app, tripID, typo, c.typed, c.quantity)

			if rec.Code != http.StatusUnprocessableEntity {
				t.Fatalf("status = %d, want 422", rec.Code)
			}
			body := rec.Body.String()
			for _, m := range c.wantMessages {
				if !strings.Contains(body, html.EscapeString(m)) {
					t.Errorf("reply does not say %q", m)
				}
			}
			tag := inputTagLabelled(t, memberSection(t, body, "Parent 1"), "Name of Sunscren")
			if !strings.Contains(tag, `value="`+c.wantNameField+`"`) {
				t.Errorf("name field should hold %q: %s", c.wantNameField, tag)
			}
			if name, q := itemRow(t, database, typo); name != "Sunscren" || q != 2 {
				t.Errorf("item = %q x%d after a refused save, want Sunscren x2", name, q)
			}
		})
	}
}

// S7: leaving the name alone and changing only the quantity still works.
func TestSaveWithUnchangedName(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	item := insertItem(t, database, tripID, "Parent 1", "Sunscreen", 2)

	if rec := saveItem(t, app, tripID, item, "Sunscreen", "3"); rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want 303", rec.Code)
	}
	if name, q := itemRow(t, database, item); name != "Sunscreen" || q != 3 {
		t.Errorf("item = %q x%d, want Sunscreen x3", name, q)
	}
}

// S8: a rename does not move the item in the packing lifecycle.
func TestRenameKeepsStatus(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	packed := insertItem(t, database, tripID, "Parent 1", "Sunscren", 1)
	prepared := insertItem(t, database, tripID, "Parent 1", "Hat", 1)
	packItem(t, database, packed)
	prepareItem(t, database, prepared)

	saveItem(t, app, tripID, packed, "Sunscreen", "1")
	saveItem(t, app, tripID, prepared, "Cap", "1")

	if s := itemStatus(t, database, packed); s != "packed" {
		t.Errorf("Sunscreen status = %q, want packed", s)
	}
	if s := itemStatus(t, database, prepared); s != "prepared" {
		t.Errorf("Cap status = %q, want prepared", s)
	}
	if body := get(t, app, packURLFor(tripID)); !strings.Contains(body, "Sunscreen") || strings.Contains(body, "Sunscren") {
		t.Errorf("packing page does not show the new name only")
	}
}

// S9: the item is gone: not found, and nothing is created.
func TestRenameRemovedItem(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	item := insertItem(t, database, tripID, "Parent 1", "Sunscren", 1)
	database.Exec("DELETE FROM item WHERE id = ?", item)

	for _, name := range []string{"Sunscreen", ""} {
		if rec := saveItem(t, app, tripID, item, name, "1"); rec.Code != http.StatusNotFound {
			t.Errorf("save with name %q: status = %d, want 404", name, rec.Code)
		}
	}
	if n := count(t, database, "item"); n != 0 {
		t.Errorf("%d items exist, want none", n)
	}
}

// S12: another parent added the same name a moment earlier.
func TestRenameToNameAddedMeanwhile(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	typo := insertItem(t, database, tripID, "Parent 1", "Sunscren", 1)
	get(t, app, editURL(tripID, typo)) // the row is open on the first phone
	insertItem(t, database, tripID, "Parent 1", "Sunscreen", 1)

	rec := saveItem(t, app, tripID, typo, "Sunscreen", "1")

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Parent 1 already has Sunscreen.") {
		t.Errorf("reply does not say Parent 1 already has Sunscreen")
	}
	if name, _ := itemRow(t, database, typo); name != "Sunscren" {
		t.Errorf("name = %q, want it kept", name)
	}
}

// S13: a save with no name field (a page from before this change) keeps the name.
func TestSaveWithoutNameKeepsName(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	item := insertItem(t, database, tripID, "Parent 1", "Sunscreen", 2)

	rec := send(t, app, http.MethodPost, quantityURL(tripID, item), url.Values{"quantity": {"3"}})

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want 303", rec.Code)
	}
	if name, q := itemRow(t, database, item); name != "Sunscreen" || q != 3 {
		t.Errorf("item = %q x%d, want Sunscreen x3", name, q)
	}
}

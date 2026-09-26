package main

import (
	"fmt"
	"html"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// Milestone 04 (spec/milestones/04-trip-basics.md), slice 3: changing an
// item's quantity, S8 to S10.
//
// Page contract these tests rely on:
//   - Each item row has a <button aria-label="Change quantity of NAME"> that
//     opens the row for editing, next to the existing "Remove NAME" button.
//   - The trip page with ?edit=ITEM_ID shows that row in edit mode: an input
//     aria-label="New quantity for NAME" named "quantity", in a form that
//     posts to /trips/TRIP/items/ITEM/quantity, a Save button and a Cancel
//     link back to the plain trip page at that row (#item-ITEM). Every
//     other row's Change and Remove buttons are disabled; the add forms are
//     not.
//   - Saving without htmx redirects (303) back to the trip page.

func quantityURL(tripID, itemID int64) string {
	return fmt.Sprintf("/trips/%d/items/%d/quantity", tripID, itemID)
}

func editURL(tripID, itemID int64) string {
	return fmt.Sprintf("/trips/%d?edit=%d", tripID, itemID)
}

func saveQuantity(t *testing.T, app *application, tripID, itemID int64, quantity string) *httptest.ResponseRecorder {
	t.Helper()
	return send(t, app, http.MethodPost, quantityURL(tripID, itemID), url.Values{"quantity": {quantity}})
}

// S8: Parent changes an item's quantity while planning.
func TestChangeItemQuantity(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	underwear := insertItem(t, database, tripID, "Parent 1", "Underwear", 4)
	book := insertItem(t, database, tripID, "Parent 1", "Book", 1)
	toothbrush := insertItem(t, database, tripID, "Parent 1", "Toothbrush", 1)
	packItem(t, database, toothbrush)

	// The row opens for editing with its current quantity.
	section := memberSection(t, get(t, app, editURL(tripID, underwear)), "Parent 1")
	if tag := inputTagLabelled(t, section, "New quantity for Underwear"); !strings.Contains(tag, `value="4"`) {
		t.Errorf("edit field does not start at 4: %s", tag)
	}
	if !strings.Contains(section, fmt.Sprintf(`action="%s"`, quantityURL(tripID, underwear))) {
		t.Errorf("edit form does not post to %s", quantityURL(tripID, underwear))
	}

	for _, c := range []struct {
		id   int64
		name string
		to   string
	}{
		{underwear, "Underwear", "5"},
		{book, "Book", "2"},
		{toothbrush, "Toothbrush", "2"},
	} {
		rec := saveQuantity(t, app, tripID, c.id, c.to)
		if rec.Code != http.StatusSeeOther {
			t.Fatalf("save %s: status = %d, want 303", c.name, rec.Code)
		}
		if loc := rec.Header().Get("Location"); !strings.HasPrefix(loc, tripURLFor(tripID)) {
			t.Errorf("save %s redirects to %q, want the trip page", c.name, loc)
		}
	}

	for _, c := range []struct {
		id   int64
		want int
	}{{underwear, 5}, {book, 2}, {toothbrush, 2}} {
		var q int
		database.QueryRow("SELECT quantity FROM item WHERE id = ?", c.id).Scan(&q)
		if q != c.want {
			t.Errorf("item %d quantity = %d, want %d", c.id, q, c.want)
		}
	}
	section = memberSection(t, get(t, app, tripURLFor(tripID)), "Parent 1")
	for _, want := range []string{"5 pcs", "2 pcs"} {
		if !strings.Contains(section, want) {
			t.Errorf("trip page does not show %q", want)
		}
	}
	if s := itemStatus(t, database, toothbrush); s != "packed" {
		t.Errorf("Toothbrush status = %q after changing its quantity, want packed", s)
	}
}

// Changing a quantity through another trip's URL must not touch the item.
func TestChangeQuantityOnlyOnItsOwnTrip(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	other := insertTrip(t, database, "Levi", "2026-12-20", 5)
	item := insertItem(t, database, tripID, "Parent 1", "Underwear", 4)

	if rec := saveQuantity(t, app, other, item, "9"); rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
	if rec := saveQuantity(t, app, other, item, "0"); rec.Code != http.StatusNotFound {
		t.Errorf("invalid quantity through another trip: status = %d, want 404", rec.Code)
	}

	var q int
	database.QueryRow("SELECT quantity FROM item WHERE id = ?", item).Scan(&q)
	if q != 4 {
		t.Errorf("quantity = %d after a save through another trip's URL, want 4", q)
	}
}

// S9: An invalid quantity is refused.
func TestInvalidQuantityIsRefused(t *testing.T) {
	for _, input := range []string{"0", "-1", "2.5", "abc"} {
		t.Run(input, func(t *testing.T) {
			app, database := newTestApp(t)
			tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
			underwear := insertItem(t, database, tripID, "Parent 1", "Underwear", 4)

			rec := saveQuantity(t, app, tripID, underwear, input)

			if rec.Code != http.StatusUnprocessableEntity {
				t.Errorf("status = %d, want 422", rec.Code)
			}
			body := rec.Body.String()
			if want := html.EscapeString("Quantity must be a whole number, at least 1."); !strings.Contains(body, want) {
				t.Errorf("reply does not explain the problem")
			}
			// The row stays open so the parent can correct it.
			inputTagLabelled(t, body, "New quantity for Underwear")
			var q int
			database.QueryRow("SELECT quantity FROM item WHERE id = ?", underwear).Scan(&q)
			if q != 4 {
				t.Errorf("quantity = %d, want it kept at 4", q)
			}
		})
	}
}

// S10: Only one item is edited at a time.
func TestOnlyOneItemIsEditedAtATime(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	underwear := insertItem(t, database, tripID, "Parent 1", "Underwear", 4)
	insertItem(t, database, tripID, "Parent 1", "Socks", 4)
	insertItem(t, database, tripID, "Parent 2", "Underwear", 4)

	body := get(t, app, editURL(tripID, underwear))
	p1 := memberSection(t, body, "Parent 1")
	p2 := memberSection(t, body, "Parent 2")

	// Other rows, in the same section and in another, are locked.
	for _, c := range []struct{ section, label string }{
		{p1, "Change quantity of Socks"},
		{p1, "Remove Socks"},
		{p2, "Change quantity of Underwear"},
		{p2, "Remove Underwear"},
	} {
		if tag := buttonTag(t, c.section, c.label); !strings.Contains(tag, "disabled") {
			t.Errorf("%q is not disabled while another row is edited: %s", c.label, tag)
		}
	}
	// Adding a new item still works (T2).
	if strings.Contains(submitTag(t, p2), "disabled") {
		t.Errorf("Parent 2's Add button is disabled while a row is edited")
	}
	// Cancel goes back to the plain trip page (no ?edit), at the same row.
	if !strings.Contains(p1, fmt.Sprintf(`href="%s#item-%d"`, tripURLFor(tripID), underwear)) || !strings.Contains(p1, "Cancel") {
		t.Errorf("edit row has no Cancel link back to the trip page")
	}
	after := get(t, app, tripURLFor(tripID))
	for _, c := range []struct{ member, label string }{
		{"Parent 1", "Change quantity of Underwear"},
		{"Parent 1", "Change quantity of Socks"},
		{"Parent 1", "Remove Socks"},
		{"Parent 2", "Remove Underwear"},
	} {
		if tag := buttonTag(t, memberSection(t, after, c.member), c.label); strings.Contains(tag, "disabled") {
			t.Errorf("%q is still disabled after leaving edit mode: %s", c.label, tag)
		}
	}
}

// inputTagLabelled returns the <input ...> tag with the given aria-label.
func inputTagLabelled(t *testing.T, body, label string) string {
	t.Helper()
	i := strings.Index(body, `aria-label="`+label+`"`)
	if i < 0 {
		t.Fatalf("no input labelled %q", label)
	}
	start := strings.LastIndex(body[:i], "<input")
	end := strings.Index(body[i:], ">")
	if start < 0 || end < 0 {
		t.Fatalf("%q is not an <input> tag", label)
	}
	return body[start : i+end+1]
}

// submitTag returns the add form's submit button in a member section.
func submitTag(t *testing.T, section string) string {
	t.Helper()
	i := strings.Index(section, `>Add</button>`)
	if i < 0 {
		t.Fatalf("section has no Add button")
	}
	start := strings.LastIndex(section[:i], "<button")
	return section[start : i+1]
}

// S10 (automated part): opening, saving and cancelling swap every member's
// lists in place rather than loading a new page, so the view does not move.
// Whether it really stays put on a phone is checked by hand.
func TestEditingSwapsListsInPlace(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	underwear := insertItem(t, database, tripID, "Parent 1", "Underwear", 4)

	page := get(t, app, tripURLFor(tripID))
	if !strings.Contains(page, `id="trip-items"`) {
		t.Fatalf("trip page has no #trip-items container to swap")
	}
	open := enclosingTag(t, page, `aria-label="Change quantity of Underwear"`, "<form")
	assertSwapsTripItems(t, "Change quantity form", open)
	if !strings.Contains(open, "hx-get=") {
		t.Errorf("Change quantity form does not load edit mode with htmx: %s", open)
	}

	edit := get(t, app, editURL(tripID, underwear))
	save := enclosingTag(t, edit, `aria-label="New quantity for Underwear"`, "<form")
	assertSwapsTripItems(t, "Save form", save)
	if want := fmt.Sprintf(`hx-post="%s"`, quantityURL(tripID, underwear)); !strings.Contains(save, want) {
		t.Errorf("Save form does not post with htmx (%s): %s", want, save)
	}
	cancel := enclosingTag(t, edit, ">Cancel</a>", "<a")
	assertSwapsTripItems(t, "Cancel link", cancel)
	if !strings.Contains(cancel, "hx-get=") {
		t.Errorf("Cancel does not go back with htmx: %s", cancel)
	}
}

// enclosingTag returns the opening tag (starting with open, e.g. "<form")
// that comes last before marker in body.
func enclosingTag(t *testing.T, body, marker, open string) string {
	t.Helper()
	i := strings.Index(body, marker)
	if i < 0 {
		t.Fatalf("page has no %s", marker)
	}
	start := strings.LastIndex(body[:i], open)
	if start < 0 {
		t.Fatalf("no %s before %s", open, marker)
	}
	end := strings.Index(body[start:], ">")
	return body[start : start+end+1]
}

func assertSwapsTripItems(t *testing.T, what, tag string) {
	t.Helper()
	for _, want := range []string{`hx-target="#trip-items"`, `hx-select="#trip-items"`} {
		if !strings.Contains(tag, want) {
			t.Errorf("%s lacks %s: %s", what, want, tag)
		}
	}
}

// S10: adding an item while a row is open keeps that row open and the other
// rows locked (the add form carries ?edit=).
func TestAddWhileEditingKeepsEditMode(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	underwear := insertItem(t, database, tripID, "Parent 1", "Underwear", 4)
	insertItem(t, database, tripID, "Parent 1", "Socks", 4)
	parent := memberID(t, database, "Parent 1")

	rec := sendHTMX(t, app, fmt.Sprintf("/trips/%d/members/%d/items?edit=%d", tripID, parent, underwear),
		url.Values{"name": {"Book"}, "quantity": {"1"}})

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Book") {
		t.Errorf("reply does not list the added Book")
	}
	inputTagLabelled(t, body, "New quantity for Underwear")
	if tag := buttonTag(t, body, "Remove Socks"); !strings.Contains(tag, "disabled") {
		t.Errorf("Socks is unlocked after adding while Underwear is open: %s", tag)
	}
}

// An edit id that is unknown or belongs to another trip shows the normal
// page: no row open, nothing locked.
func TestEditParamForAnotherItemIsIgnored(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	other := insertTrip(t, database, "Levi", "2026-12-20", 5)
	insertItem(t, database, tripID, "Parent 1", "Socks", 4)
	foreign := insertItem(t, database, other, "Parent 1", "Underwear", 4)

	for _, id := range []int64{foreign, 999999} {
		body := get(t, app, editURL(tripID, id))
		if strings.Contains(body, "New quantity for") {
			t.Errorf("?edit=%d opened a row on the wrong trip", id)
		}
		if tag := buttonTag(t, body, "Remove Socks"); strings.Contains(tag, "disabled") {
			t.Errorf("?edit=%d locked the page: %s", id, tag)
		}
	}
}

package main

import (
	"net/http"
	"strings"
	"testing"
)

// Milestone 08 (spec/milestones/08-gather-and-pack.md), slice 2, scenarios
// S3, S8 and S9: an item goes planned, then prepared, then packed, and each
// tap names its step. A step only moves an item that is in the state the
// step starts from (Scope 6): prepare only a planned item, pack only a
// prepared one, unpack only a packed one. Anything else saves nothing and
// answers with the owner's list, packed part and progress as they really
// are, so a parent whose page is out of date never moves an item backwards
// or two steps at once. S4, S5 and S7 are the rewritten 02/S2, 02/S3 and
// 03/S3 tests in pack_toggle_test.go and family_pack_test.go.

// S3: Parent prepares an item.
func TestPrepareItemInPlace(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	child1 := familyMembers[2]
	id := memberID(t, database, child1)
	swimsuit := insertItem(t, database, tripID, child1, "Swimsuit", 1)
	towel := insertItem(t, database, tripID, child1, "Towel", 1)
	insertItem(t, database, tripID, child1, "Sun hat", 1)

	// Before the tap every row is planned: it offers to prepare the item,
	// and nothing says "Prepared".
	section := memberSection(t, get(t, app, packURLFor(tripID)), child1)
	if !strings.Contains(section, `action="`+prepareItemURL(tripID, id, swimsuit)+`"`) {
		t.Errorf("a planned row's tap does not prepare the item:\n%s", section)
	}
	if !strings.Contains(section, `aria-label="Prepare Swimsuit"`) {
		t.Errorf("a planned row is not labelled \"Prepare Swimsuit\":\n%s", section)
	}
	if strings.Contains(section, "Prepared") {
		t.Errorf("a planned item says it is prepared:\n%s", section)
	}

	rec := sendHTMX(t, app, prepareItemURL(tripID, id, swimsuit), nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	toPack := packFragment(t, body, toPackListID(id))
	assertInOrder(t, toPack, "Swimsuit", "Towel", "Sun hat")

	row := packRow(t, toPack, "Swimsuit")
	if !strings.Contains(row, "Prepared") {
		t.Errorf("the prepared row does not say \"Prepared\":\n%s", row)
	}
	if !strings.Contains(row, `aria-label="Pack Swimsuit"`) || strings.Contains(row, "Prepare Swimsuit") {
		t.Errorf("the prepared row does not offer to pack instead of prepare:\n%s", row)
	}
	if !strings.Contains(row, `action="`+packItemURL(tripID, id, swimsuit)+`"`) {
		t.Errorf("the prepared row's tap does not pack the item:\n%s", row)
	}
	for _, name := range []string{"Towel", "Sun hat"} {
		if planned := packRow(t, toPack, name); strings.Contains(planned, "Prepared") {
			t.Errorf("planned %q says it is prepared:\n%s", name, planned)
		}
	}
	if planned := packRow(t, toPack, "Towel"); !strings.Contains(planned, `action="`+prepareItemURL(tripID, id, towel)+`"`) {
		t.Errorf("the other planned rows no longer prepare:\n%s", planned)
	}

	if progress := packFragment(t, body, "pack-progress"); !strings.Contains(progress, "0 of 3 packed") {
		t.Errorf("a prepared item is counted as packed:\n%s", progress)
	}
	if got := itemStatus(t, database, swimsuit); got != "prepared" {
		t.Errorf("stored status = %q, want %q", got, "prepared")
	}

	// Still prepared on the next visit.
	page := memberSection(t, get(t, app, packURLFor(tripID)), child1)
	if row := packRow(t, page, "Swimsuit"); !strings.Contains(row, "Prepared") {
		t.Errorf("the prepared mark did not survive reloading the page:\n%s", row)
	}
}

// S8: Parent 1's page still shows "Swimsuit" as planned, but Parent 2 has
// prepared it since. Parent 1's tap sends prepare: the item stays prepared,
// not packed, and the reply brings the list up to date.
func TestPrepareAnItemAlreadyPrepared(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	child1 := familyMembers[2]
	id := memberID(t, database, child1)
	swimsuit := insertItem(t, database, tripID, child1, "Swimsuit", 1)
	insertItem(t, database, tripID, child1, "Towel", 1)
	prepareItem(t, database, swimsuit)

	rec := sendHTMX(t, app, prepareItemURL(tripID, id, swimsuit), nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if got := itemStatus(t, database, swimsuit); got != "prepared" {
		t.Errorf("stored status = %q, want %q", got, "prepared")
	}
	body := rec.Body.String()
	row := packRow(t, packFragment(t, body, toPackListID(id)), "Swimsuit")
	if !strings.Contains(row, "Prepared") || !strings.Contains(row, `aria-label="Pack Swimsuit"`) {
		t.Errorf("the list did not come back up to date with Swimsuit prepared:\n%s", row)
	}
	if progress := packFragment(t, body, "pack-progress"); !strings.Contains(progress, "0 of 2 packed") {
		t.Errorf("progress does not say 0 of 2 packed:\n%s", progress)
	}
}

// S9: Parent 1's page still shows "Swimsuit" as planned, but Parent 2 has
// prepared and packed it since. Parent 1's tap sends prepare: the item stays
// packed. Parent 1's page had no packed part for Child 1, so the reply must
// carry the whole part, not only a count and a list to swap into it.
func TestPrepareAnItemAlreadyPacked(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	child1 := familyMembers[2]
	id := memberID(t, database, child1)
	swimsuit := insertItem(t, database, tripID, child1, "Swimsuit", 1)
	insertItem(t, database, tripID, child1, "Towel", 1)
	insertItem(t, database, tripID, child1, "Sun hat", 1)
	packItem(t, database, swimsuit)

	rec := sendHTMX(t, app, prepareItemURL(tripID, id, swimsuit), nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if got := itemStatus(t, database, swimsuit); got != "packed" {
		t.Errorf("stored status = %q, want %q", got, "packed")
	}
	body := rec.Body.String()
	if toPack := packFragment(t, body, toPackListID(id)); strings.Contains(toPack, "Swimsuit") {
		t.Errorf("the packed item is back in the list of things to pack:\n%s", toPack)
	}
	part := packedPart(t, body, id)
	if !strings.Contains(part, "<details") || !strings.Contains(part, "Packed (1)") || !strings.Contains(part, "Swimsuit") {
		t.Errorf("the reply does not bring Child 1's packed part, holding Swimsuit:\n%s", part)
	}
	if progress := packFragment(t, body, "pack-progress"); !strings.Contains(progress, "1 of 3 packed") {
		t.Errorf("progress does not say 1 of 3 packed:\n%s", progress)
	}
}

// Scope 6, the rest of the wrong-state pairs: a step sent for an item in
// another state saves nothing and answers with the list as it really is.
// Prepare on a prepared or packed item is S8 and S9 above.
func TestStepsOnlyMoveItemsFromTheirStartingState(t *testing.T) {
	cases := []struct {
		name   string
		status string // the item's stored state when the tap arrives
		url    func(tripID, memberID, itemID int64) string
		label  string // the row's spoken label in the reply
	}{
		{"pack a planned item", "planned", packItemURL, "Prepare Swimsuit"},
		{"unpack a prepared item", "prepared", unpackItemURL, "Pack Swimsuit"},
		{"unpack a planned item", "planned", unpackItemURL, "Prepare Swimsuit"},
		{"take back a planned item", "planned", unprepareItemURL, "Prepare Swimsuit"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			app, database := newTestApp(t)
			tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
			child1 := familyMembers[2]
			id := memberID(t, database, child1)
			swimsuit := insertItem(t, database, tripID, child1, "Swimsuit", 1)
			insertItem(t, database, tripID, child1, "Towel", 1)
			if c.status == "prepared" {
				prepareItem(t, database, swimsuit)
			}

			rec := sendHTMX(t, app, c.url(tripID, id, swimsuit), nil)

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200", rec.Code)
			}
			if got := itemStatus(t, database, swimsuit); got != c.status {
				t.Errorf("stored status = %q, want it left at %q", got, c.status)
			}
			body := rec.Body.String()
			if row := packRow(t, packFragment(t, body, toPackListID(id)), "Swimsuit"); !strings.Contains(row, `aria-label="`+c.label+`"`) {
				t.Errorf("the reply does not show Swimsuit as it really is (%q):\n%s", c.label, row)
			}
			if progress := packFragment(t, body, "pack-progress"); !strings.Contains(progress, "0 of 2 packed") {
				t.Errorf("progress does not say 0 of 2 packed:\n%s", progress)
			}
		})
	}
}

// packRow returns the list row (<li>) in html that holds name.
func packRow(t *testing.T, html, name string) string {
	t.Helper()
	for _, row := range strings.Split(html, "<li")[1:] {
		if strings.Contains(row, name) {
			return "<li" + row
		}
	}
	t.Fatalf("no row for %q in:\n%s", name, html)
	return ""
}

// assertInOrder checks that names appear in html in the given order.
func assertInOrder(t *testing.T, html string, names ...string) {
	t.Helper()
	last := -1
	for _, name := range names {
		at := strings.Index(html, name)
		if at < 0 {
			t.Errorf("%q is missing from the list:\n%s", name, html)
			return
		}
		if at < last {
			t.Errorf("the list is out of order, want %v:\n%s", names, html)
			return
		}
		last = at
	}
}

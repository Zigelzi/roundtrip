package main

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
)

// Milestone 08 (spec/milestones/08-gather-and-pack.md), slice 3, scenarios
// S6 and S10: a prepared item can be taken back to planned with a back
// button at the end of its row. Like every other step it only moves an item
// from the state it starts from (Scope 6): take back only a prepared item.
// Taking back a planned item is a row in the wrong-state table in
// prepare_test.go.

// S6: Parent takes a prepared item back to planned.
func TestUnprepareItemInPlace(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	child1 := familyMembers[2]
	id := memberID(t, database, child1)
	swimsuit := insertItem(t, database, tripID, child1, "Swimsuit", 1)
	insertItem(t, database, tripID, child1, "Towel", 1)
	insertItem(t, database, tripID, child1, "Sun hat", 1)
	goggles := insertItem(t, database, tripID, child1, "Goggles", 1)
	prepareItem(t, database, swimsuit)
	packItem(t, database, goggles)

	// Only the prepared row has a back button. It is a form of its own
	// beside the row's tap, serialised like every other tap.
	page := get(t, app, packURLFor(tripID))
	section := memberSection(t, page, child1)
	toPack := packFragment(t, page, toPackListID(id))
	row := packRow(t, toPack, "Swimsuit")
	back := `action="` + unprepareItemURL(tripID, id, swimsuit) + `"`
	if !strings.Contains(row, back) {
		t.Errorf("the prepared row has no back button:\n%s", row)
	}
	if !strings.Contains(row, `aria-label="Back to planned: Swimsuit"`) {
		t.Errorf("the back button is not labelled \"Back to planned: Swimsuit\":\n%s", row)
	}
	if strings.Count(row, `hx-sync="body:queue all"`) < 2 {
		t.Errorf("the back button does not serialise its tap with the others:\n%s", row)
	}
	if n := strings.Count(section, `aria-label="Back to planned:`); n != 1 {
		t.Errorf("%d back buttons in Child 1's section, want 1 (planned and packed rows have none):\n%s", n, section)
	}

	// On a page narrowed to Child 1 the back button keeps the narrowing.
	narrowed := get(t, app, narrowedPackURL(tripID, id))
	if want := fmt.Sprintf(`action="%s?member=%d"`, unprepareItemURL(tripID, id, swimsuit), id); !strings.Contains(narrowed, want) {
		t.Errorf("the back button on the narrowed page does not keep the narrowing (%s)", want)
	}

	rec := sendHTMX(t, app, unprepareItemURL(tripID, id, swimsuit), partShown)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if got := itemStatus(t, database, swimsuit); got != "planned" {
		t.Errorf("stored status = %q, want %q", got, "planned")
	}
	body := rec.Body.String()
	toPack = packFragment(t, body, toPackListID(id))
	assertInOrder(t, toPack, "Swimsuit", "Towel", "Sun hat")
	row = packRow(t, toPack, "Swimsuit")
	if strings.Contains(row, "Prepared") || strings.Contains(row, "Back to planned") {
		t.Errorf("the row still shows Swimsuit as prepared:\n%s", row)
	}
	if !strings.Contains(row, `aria-label="Prepare Swimsuit"`) {
		t.Errorf("the row does not offer to prepare Swimsuit again:\n%s", row)
	}
	if progress := packFragment(t, body, "pack-progress"); !strings.Contains(progress, "1 of 4 packed") {
		t.Errorf("progress does not say 1 of 4 packed:\n%s", progress)
	}
	// Swapping the <details> would close an open packed part.
	if strings.Contains(body, "<details") {
		t.Errorf("reply contains <details>, which would collapse the packed part:\n%s", body)
	}
}

// S10: Parent 1's page still shows "Swimsuit" as prepared, but Parent 2 has
// packed it since. Parent 1's back tap leaves it packed, and the reply
// carries the whole packed part, as in S9.
func TestUnprepareAnItemAlreadyPacked(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	child1 := familyMembers[2]
	id := memberID(t, database, child1)
	swimsuit := insertItem(t, database, tripID, child1, "Swimsuit", 1)
	insertItem(t, database, tripID, child1, "Towel", 1)
	packItem(t, database, swimsuit)

	rec := sendHTMX(t, app, unprepareItemURL(tripID, id, swimsuit), nil)

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
	if progress := packFragment(t, body, "pack-progress"); !strings.Contains(progress, "1 of 2 packed") {
		t.Errorf("progress does not say 1 of 2 packed:\n%s", progress)
	}
}

// 02/S10 for the back button: an item the other parent has removed does not
// come back, and the rest of the list does.
func TestUnprepareItemThatIsGone(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	first := memberID(t, database, familyMembers[0])
	removed := insertItem(t, database, tripID, familyMembers[0], "Toothbrush", 1)
	insertItem(t, database, tripID, familyMembers[0], "Underwear", 4)
	if _, err := database.Exec("DELETE FROM item WHERE id = ?", removed); err != nil {
		t.Fatalf("remove item: %v", err)
	}

	rec := sendHTMX(t, app, unprepareItemURL(tripID, first, removed), nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	toPack := packFragment(t, rec.Body.String(), toPackListID(first))
	if strings.Contains(toPack, "Toothbrush") {
		t.Errorf("the removed item came back:\n%s", toPack)
	}
	if !strings.Contains(toPack, "Underwear") {
		t.Errorf("the rest of the list is missing:\n%s", toPack)
	}
}

// An item on another trip cannot be taken back through this trip. It is
// stored prepared, so a step that skipped the trip check would move it.
func TestUnprepareItemFromAnotherTrip(t *testing.T) {
	app, database := newTestApp(t)
	tripA := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	tripB := insertTrip(t, database, "Turku", "2026-10-01", 2)
	first := memberID(t, database, familyMembers[0])
	itemOnB := insertItem(t, database, tripB, familyMembers[0], "Toothbrush", 1)
	prepareItem(t, database, itemOnB)

	sendHTMX(t, app, unprepareItemURL(tripA, first, itemOnB), nil)

	if got := itemStatus(t, database, itemOnB); got != "prepared" {
		t.Errorf("item on another trip was moved through this trip: status = %q", got)
	}
}

func unprepareItemURL(tripID, memberID, itemID int64) string {
	return fmt.Sprintf("/trips/%d/members/%d/items/%d/unprepare", tripID, memberID, itemID)
}

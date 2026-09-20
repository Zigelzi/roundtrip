package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"testing"
)

// Milestone 02 (spec/milestones/02-pack-items.md), scenarios S2, S3 and S10:
// ticking an item packed and back again, in place.

// S2: Parent marks an item as packed.
func TestPackItemInPlace(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	first := memberID(t, database, familyMembers[0])
	toothbrush := insertItem(t, database, tripID, familyMembers[0], "Toothbrush", 1)
	insertItem(t, database, tripID, familyMembers[0], "Underwear", 4)
	insertItem(t, database, tripID, familyMembers[1], "Raincoat", 1)

	rec := sendHTMX(t, app, packItemURL(tripID, first, toothbrush), nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	toPack := packFragment(t, body, toPackListID(first))
	if strings.Contains(toPack, "Toothbrush") {
		t.Errorf("packed item is still in the list of things to pack:\n%s", toPack)
	}
	if !strings.Contains(toPack, "Underwear") {
		t.Errorf("the member's other item vanished from the list:\n%s", toPack)
	}
	if packed := packFragment(t, body, "packed-list"); !strings.Contains(packed, "Toothbrush") {
		t.Errorf("packed item is not in the packed section:\n%s", packed)
	}
	if progress := packFragment(t, body, "pack-progress"); !strings.Contains(progress, "1 of 3 packed") {
		t.Errorf("progress does not say 1 of 3 packed:\n%s", progress)
	}
	if got := itemStatus(t, database, toothbrush); got != "packed" {
		t.Errorf("stored status = %q, want %q", got, "packed")
	}

	// Still packed on the next visit.
	if page := get(t, app, packURLFor(tripID)); !strings.Contains(page, "1 of 3 packed") {
		t.Errorf("the tick did not survive reloading the page")
	}
}

// S3: Parent unpacks an item.
func TestUnpackItemInPlace(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	first := memberID(t, database, familyMembers[0])
	toothbrush := insertItem(t, database, tripID, familyMembers[0], "Toothbrush", 1)
	insertItem(t, database, tripID, familyMembers[0], "Underwear", 4)
	insertItem(t, database, tripID, familyMembers[1], "Raincoat", 1)
	packItem(t, database, toothbrush)

	rec := sendHTMX(t, app, unpackItemURL(tripID, first, toothbrush), nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	if toPack := packFragment(t, body, toPackListID(first)); !strings.Contains(toPack, "Toothbrush") {
		t.Errorf("unpacked item is not back in the list of things to pack:\n%s", toPack)
	}
	if packed := packFragment(t, body, "packed-list"); strings.Contains(packed, "Toothbrush") {
		t.Errorf("unpacked item is still in the packed section:\n%s", packed)
	}
	if progress := packFragment(t, body, "pack-progress"); !strings.Contains(progress, "0 of 3 packed") {
		t.Errorf("progress does not say 0 of 3 packed:\n%s", progress)
	}
	if got := itemStatus(t, database, toothbrush); got != "planned" {
		t.Errorf("stored status = %q, want %q", got, "planned")
	}
}

// S3 (automated part of "the packed section is still open"): a reply must
// never carry the <details> element itself. Swapping it would reset the
// open/closed state, so the parent would have to reopen it after every tap.
// Only what is inside it is swapped.
func TestPackRepliesNeverSwapThePackedSectionItself(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	first := memberID(t, database, familyMembers[0])
	item := insertItem(t, database, tripID, familyMembers[0], "Toothbrush", 1)

	for _, url := range []string{packItemURL(tripID, first, item), unpackItemURL(tripID, first, item)} {
		body := sendHTMX(t, app, url, nil).Body.String()
		if strings.Contains(body, "<details") {
			t.Errorf("POST %s: reply contains <details>, which would collapse the packed section:\n%s", url, body)
		}
	}
}

// S10: Parent taps an item the other parent has already removed.
func TestPackItemThatIsGone(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	first := memberID(t, database, familyMembers[0])
	removed := insertItem(t, database, tripID, familyMembers[0], "Toothbrush", 1)
	insertItem(t, database, tripID, familyMembers[0], "Underwear", 4)
	if _, err := database.Exec("DELETE FROM item WHERE id = ?", removed); err != nil {
		t.Fatalf("remove item: %v", err)
	}

	rec := sendHTMX(t, app, packItemURL(tripID, first, removed), nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	toPack := packFragment(t, rec.Body.String(), toPackListID(first))
	if strings.Contains(toPack, "Toothbrush") {
		t.Errorf("the removed item came back:\n%s", toPack)
	}
	if !strings.Contains(toPack, "Underwear") {
		t.Errorf("the parent cannot carry on packing; the rest of the list is missing:\n%s", toPack)
	}
	if n := countPacked(t, database); n != 0 {
		t.Errorf("%d items packed, want 0 -- nothing should have been saved", n)
	}
}

// S10: an item belonging to another trip cannot be packed through this trip.
func TestPackItemFromAnotherTrip(t *testing.T) {
	app, database := newTestApp(t)
	tripA := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	tripB := insertTrip(t, database, "Turku", "2026-10-01", 2)
	first := memberID(t, database, familyMembers[0])
	itemOnB := insertItem(t, database, tripB, familyMembers[0], "Toothbrush", 1)

	sendHTMX(t, app, packItemURL(tripA, first, itemOnB), nil)

	if got := itemStatus(t, database, itemOnB); got != "planned" {
		t.Errorf("item on another trip was packed through this trip: status = %q", got)
	}
}

// S2: without JavaScript the tap is a plain form post, answered with a
// redirect back to the packing page.
func TestPackItemWithoutHTMXRedirects(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	first := memberID(t, database, familyMembers[0])
	item := insertItem(t, database, tripID, familyMembers[0], "Toothbrush", 1)

	rec := send(t, app, http.MethodPost, packItemURL(tripID, first, item), nil)

	if rec.Code != http.StatusSeeOther {
		t.Errorf("status = %d, want 303", rec.Code)
	}
	if got, want := rec.Header().Get("Location"), packURLFor(tripID); got != want {
		t.Errorf("Location = %q, want %q", got, want)
	}
	if got := itemStatus(t, database, item); got != "packed" {
		t.Errorf("stored status = %q, want %q", got, "packed")
	}
}

// Taps are serialised: every reply is a full snapshot of the member's list,
// the packed section and the progress, so two replies arriving out of order
// would show the older one and visually un-tick an item. Packing 120 items
// in a row is exactly the pattern that overlaps requests.
func TestPackRowsSerialiseTaps(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	insertItem(t, database, tripID, familyMembers[0], "Underwear", 4)

	body := get(t, app, packURLFor(tripID))

	if !strings.Contains(body, `hx-sync="body:queue all"`) {
		t.Errorf("rows do not serialise their requests; taps can arrive out of order:\n%s", body)
	}
}

// The list re-rendered after a tap is the item's real owner's, whatever the
// URL claims the member is. Re-rendering someone else's list would leave the
// changed one stale on screen.
func TestPackItemRerendersTheItemsOwner(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	owner := memberID(t, database, familyMembers[0])
	other := memberID(t, database, familyMembers[1])
	item := insertItem(t, database, tripID, familyMembers[0], "Underwear", 4)

	rec := sendHTMX(t, app, packItemURL(tripID, other, item), nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if toPack := packFragment(t, rec.Body.String(), toPackListID(owner)); strings.Contains(toPack, "Underwear") {
		t.Errorf("the owner's list was not the one re-rendered:\n%s", toPack)
	}
}

// A member id that is not on the trip writes nothing and is a bad request,
// not a server fault.
func TestPackItemWithUnknownMemberSavesNothing(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)

	rec := sendHTMX(t, app, packItemURL(tripID, 999, 999), nil)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
	if n := countPacked(t, database); n != 0 {
		t.Errorf("%d items packed, want 0", n)
	}
}

func packItemURL(tripID, memberID, itemID int64) string {
	return fmt.Sprintf("/trips/%d/members/%d/items/%d/pack", tripID, memberID, itemID)
}

func unpackItemURL(tripID, memberID, itemID int64) string {
	return fmt.Sprintf("/trips/%d/members/%d/items/%d/unpack", tripID, memberID, itemID)
}

func toPackListID(memberID int64) string {
	return fmt.Sprintf("topack-member-%d", memberID)
}

func itemStatus(t *testing.T, database *sql.DB, itemID int64) string {
	t.Helper()
	var status string
	if err := database.QueryRow("SELECT status FROM item WHERE id = ?", itemID).Scan(&status); err != nil {
		t.Fatalf("status of item %d: %v", itemID, err)
	}
	return status
}

func countPacked(t *testing.T, database *sql.DB) int {
	t.Helper()
	var n int
	if err := database.QueryRow("SELECT count(*) FROM item WHERE status = 'packed'").Scan(&n); err != nil {
		t.Fatalf("count packed: %v", err)
	}
	return n
}

// packFragment returns one swap target out of an htmx reply. A tap updates
// three places at once (the member's list, the packed section and the
// progress), so a reply holds several fragments as top-level siblings; each
// one runs until the next begins.
func packFragment(t *testing.T, body, id string) string {
	t.Helper()
	var starts []int
	for _, want := range []string{toPackListID(1), toPackListID(2), toPackListID(3), toPackListID(4), "packed-list", "pack-progress"} {
		if i := strings.Index(body, `<div id="`+want+`"`); i >= 0 {
			starts = append(starts, i)
		}
	}
	sort.Ints(starts)

	from := strings.Index(body, `<div id="`+id+`"`)
	if from < 0 {
		t.Fatalf("reply has no fragment %q:\n%s", id, body)
	}
	for _, s := range starts {
		if s > from {
			return body[from:s]
		}
	}
	return body[from:]
}

// S1/S3: the packed section says how many items it holds, so it is obvious
// there is something inside to open. The count is swapped on every tap.
func TestPackedSectionShowsItsCount(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	first := memberID(t, database, familyMembers[0])
	item := insertItem(t, database, tripID, familyMembers[0], "Toothbrush", 1)
	insertItem(t, database, tripID, familyMembers[0], "Underwear", 4)

	if body := get(t, app, packURLFor(tripID)); !strings.Contains(body, "Packed (0)") {
		t.Errorf("the packed section does not say how many items it holds:\n%s", body)
	}

	body := sendHTMX(t, app, packItemURL(tripID, first, item), nil).Body.String()

	if !strings.Contains(body, `id="packed-summary"`) {
		t.Errorf("a tap does not update the packed section's count:\n%s", body)
	}
	if !strings.Contains(body, "Packed (1)") {
		t.Errorf("the packed count did not go up after a tap:\n%s", body)
	}
}

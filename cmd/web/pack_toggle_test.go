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
// ticking an item packed and back again, in place. Milestone 08
// (spec/milestones/08-gather-and-pack.md) replaced S2 and S3: slice 1 moved
// the packed items from one shared section into each member's own packed
// part, and slice 2 put prepared between planned and packed, so only a
// prepared item is packed and unpacking returns it to prepared. The tests
// for S2 and S3 below check 08/S4 and 08/S5 now.

// 08/S4: Parent packs a prepared item.
func TestPackItemInPlace(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	first := memberID(t, database, familyMembers[0])
	toothbrush := insertItem(t, database, tripID, familyMembers[0], "Toothbrush", 1)
	prepareItem(t, database, toothbrush)
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
	if packed := packedPart(t, body, first); !strings.Contains(packed, "Toothbrush") {
		t.Errorf("packed item is not in the member's packed part:\n%s", packed)
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

// 08/S5: Parent unpacks an item. It goes back to prepared, not planned, and
// back to its usual place among the member's planned and prepared items
// (they are listed in the order they were added). Two items are packed, so
// the packed part is still there afterwards with one item left in it.
func TestUnpackItemInPlace(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	child1 := familyMembers[2]
	id := memberID(t, database, child1)
	swimsuit := insertItem(t, database, tripID, child1, "Swimsuit", 1)
	towel := insertItem(t, database, tripID, child1, "Towel", 1)
	insertItem(t, database, tripID, child1, "Sun hat", 1)
	prepareItem(t, database, insertItem(t, database, tripID, child1, "Shorts", 2))
	packItem(t, database, swimsuit)
	packItem(t, database, towel)

	rec := sendHTMX(t, app, unpackItemURL(tripID, id, swimsuit), nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	toPack := packFragment(t, body, toPackListID(id))
	assertInOrder(t, toPack, "Swimsuit", "Sun hat", "Shorts")
	row := packRow(t, toPack, "Swimsuit")
	if !strings.Contains(row, `aria-label="Pack Swimsuit"`) || !strings.Contains(row, "Prepared") {
		t.Errorf("the unpacked item does not show as prepared, ready to pack:\n%s", row)
	}
	packed := packedPart(t, body, id)
	if strings.Contains(packed, "Swimsuit") {
		t.Errorf("unpacked item is still in the member's packed part:\n%s", packed)
	}
	if !strings.Contains(packed, "Towel") {
		t.Errorf("the other packed item vanished from the packed part:\n%s", packed)
	}
	if !strings.Contains(packed, "Packed (1)") {
		t.Errorf("the packed part's count did not go down to 1:\n%s", packed)
	}
	if progress := packFragment(t, body, "pack-progress"); !strings.Contains(progress, "1 of 4 packed") {
		t.Errorf("progress does not say 1 of 4 packed:\n%s", progress)
	}
	if got := itemStatus(t, database, swimsuit); got != "prepared" {
		t.Errorf("stored status = %q, want %q", got, "prepared")
	}
}

// S3, and 08/S5's "the packed part is still open": while a member's packed
// part exists, a reply must never carry its <details> element. Swapping it
// would reset the open/closed state, so the parent would have to reopen it
// after every tap. Only the count and the list inside it are swapped.
func TestPackRepliesNeverSwapAnExistingPackedPart(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	first := memberID(t, database, familyMembers[0])
	packItem(t, database, insertItem(t, database, tripID, familyMembers[0], "Pyjamas", 1))
	packItem(t, database, insertItem(t, database, tripID, familyMembers[0], "Socks", 1))
	toothbrush := insertItem(t, database, tripID, familyMembers[0], "Toothbrush", 1)
	prepareItem(t, database, toothbrush)
	underwear := insertItem(t, database, tripID, familyMembers[0], "Underwear", 4)
	packItem(t, database, underwear)

	// Packing one more item, then unpacking one: the part exists before and
	// after both taps.
	for _, url := range []string{packItemURL(tripID, first, toothbrush), unpackItemURL(tripID, first, underwear)} {
		body := sendHTMX(t, app, url, partShown).Body.String()
		if strings.Contains(body, "<details") {
			t.Errorf("POST %s: reply contains <details>, which would collapse the packed part:\n%s", url, body)
		}
		if !strings.Contains(body, `id="`+packedSummaryID(first)+`"`) {
			t.Errorf("POST %s: reply does not update the packed part's count:\n%s", url, body)
		}
	}
}

// 08/S1 (R4): a member with nothing packed has no packed part, so the part
// first appears with the first packed item. It did not exist before, so
// sending it whole, closed, collapses nothing.
func TestPackedPartAppearsWithTheFirstPackedItem(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	first := memberID(t, database, familyMembers[0])
	toothbrush := insertItem(t, database, tripID, familyMembers[0], "Toothbrush", 1)
	prepareItem(t, database, toothbrush)
	insertItem(t, database, tripID, familyMembers[0], "Underwear", 4)

	body := sendHTMX(t, app, packItemURL(tripID, first, toothbrush), nil).Body.String()

	part := packedPart(t, body, first)
	if !strings.Contains(part, "<details") {
		t.Errorf("the member's packed part did not appear with the first packed item:\n%s", part)
	}
	if detailsIsOpen(part) {
		t.Errorf("a new packed part should start closed:\n%s", part)
	}
	if !strings.Contains(part, "Packed (1)") || !strings.Contains(part, "Toothbrush") {
		t.Errorf("the new packed part does not hold the packed item with its count:\n%s", part)
	}
}

// 08/S1 (R4): unpacking the last packed item removes the member's packed
// part again, rather than leaving an empty "Packed (0)".
func TestPackedPartGoesWithTheLastUnpackedItem(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	first := memberID(t, database, familyMembers[0])
	toothbrush := insertItem(t, database, tripID, familyMembers[0], "Toothbrush", 1)
	insertItem(t, database, tripID, familyMembers[0], "Underwear", 4)
	packItem(t, database, toothbrush)

	body := sendHTMX(t, app, unpackItemURL(tripID, first, toothbrush), nil).Body.String()

	part := packFragment(t, body, packedPartID(first))
	if strings.Contains(part, "<details") || strings.Contains(part, "Packed (") {
		t.Errorf("the packed part is still there after its last item was unpacked:\n%s", part)
	}
}

// stepURLs are the addresses that move an item forward, for the tests below
// that hold for every step. Unpack is left out: it starts from packed, and
// these tests start from planned.
var stepURLs = map[string]func(tripID, memberID, itemID int64) string{
	"prepare": prepareItemURL,
	"pack":    packItemURL,
}

// S10: Parent taps an item the other parent has already removed.
func TestPackItemThatIsGone(t *testing.T) {
	for step, stepURL := range stepURLs {
		t.Run(step, func(t *testing.T) {
			app, database := newTestApp(t)
			tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
			first := memberID(t, database, familyMembers[0])
			removed := insertItem(t, database, tripID, familyMembers[0], "Toothbrush", 1)
			insertItem(t, database, tripID, familyMembers[0], "Underwear", 4)
			if _, err := database.Exec("DELETE FROM item WHERE id = ?", removed); err != nil {
				t.Fatalf("remove item: %v", err)
			}

			rec := sendHTMX(t, app, stepURL(tripID, first, removed), nil)

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
			if n := countMoved(t, database); n != 0 {
				t.Errorf("%d items moved past planned, want 0 -- nothing should have been saved", n)
			}
		})
	}
}

// S10: an item belonging to another trip cannot be moved through this trip.
func TestPackItemFromAnotherTrip(t *testing.T) {
	for step, stepURL := range stepURLs {
		t.Run(step, func(t *testing.T) {
			app, database := newTestApp(t)
			tripA := insertTrip(t, database, "Parainen", "2026-09-20", 3)
			tripB := insertTrip(t, database, "Turku", "2026-10-01", 2)
			first := memberID(t, database, familyMembers[0])
			itemOnB := insertItem(t, database, tripB, familyMembers[0], "Toothbrush", 1)

			sendHTMX(t, app, stepURL(tripA, first, itemOnB), nil)

			if got := itemStatus(t, database, itemOnB); got != "planned" {
				t.Errorf("item on another trip was moved through this trip: status = %q", got)
			}
		})
	}
}

// S2: without JavaScript the tap is a plain form post, answered with a
// redirect back to the packing page. Both forward steps and taking a
// prepared item back work this way.
func TestPackItemWithoutHTMXRedirects(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	first := memberID(t, database, familyMembers[0])
	item := insertItem(t, database, tripID, familyMembers[0], "Toothbrush", 1)

	for _, step := range []struct{ url, want string }{
		{prepareItemURL(tripID, first, item), "prepared"},
		{unprepareItemURL(tripID, first, item), "planned"},
		{prepareItemURL(tripID, first, item), "prepared"},
		{packItemURL(tripID, first, item), "packed"},
	} {
		rec := send(t, app, http.MethodPost, step.url, nil)

		if rec.Code != http.StatusSeeOther {
			t.Errorf("POST %s: status = %d, want 303", step.url, rec.Code)
		}
		if got, want := rec.Header().Get("Location"), packURLFor(tripID); got != want {
			t.Errorf("POST %s: Location = %q, want %q", step.url, got, want)
		}
		if got := itemStatus(t, database, item); got != step.want {
			t.Errorf("POST %s: stored status = %q, want %q", step.url, got, step.want)
		}
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
	prepareItem(t, database, item)

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

	for _, url := range []string{prepareItemURL(tripID, 999, 999), packItemURL(tripID, 999, 999), unprepareItemURL(tripID, 999, 999)} {
		rec := sendHTMX(t, app, url, nil)

		if rec.Code != http.StatusNotFound {
			t.Errorf("POST %s: status = %d, want 404", url, rec.Code)
		}
	}
	if n := countMoved(t, database); n != 0 {
		t.Errorf("%d items moved past planned, want 0", n)
	}
}

func prepareItemURL(tripID, memberID, itemID int64) string {
	return fmt.Sprintf("/trips/%d/members/%d/items/%d/prepare", tripID, memberID, itemID)
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

// countMoved counts the items that are no longer planned: prepared or
// packed. A tap that should save nothing leaves it at zero.
func countMoved(t *testing.T, database *sql.DB) int {
	t.Helper()
	var n int
	if err := database.QueryRow("SELECT count(*) FROM item WHERE status <> 'planned'").Scan(&n); err != nil {
		t.Fatalf("count moved: %v", err)
	}
	return n
}

// packFragment returns one swap target out of an htmx reply. A tap updates
// several places at once (the member's list, their packed part and the
// progress), so a reply holds several fragments as top-level siblings; each
// one runs until the next begins. A member's whole packed part, sent when it
// appears, is cut at the count and list nested inside it; packedPart puts
// the pieces back together.
func packFragment(t *testing.T, body, id string) string {
	t.Helper()
	var starts []int
	for _, prefix := range []string{"topack-member-", "packed-member-", "packed-summary-member-", "packed-list-member-", "pack-progress"} {
		for off := 0; ; {
			i := strings.Index(body[off:], `id="`+prefix)
			if i < 0 {
				break
			}
			starts = append(starts, strings.LastIndex(body[:off+i], "<"))
			off += i + 1
		}
	}
	sort.Ints(starts)

	at := strings.Index(body, `id="`+id+`"`)
	if at < 0 {
		t.Fatalf("reply has no fragment %q:\n%s", id, body)
	}
	from := strings.LastIndex(body[:at], "<")
	for _, s := range starts {
		if s > from {
			return body[from:s]
		}
	}
	return body[from:]
}

// packedPart is everything a reply carries for one member's packed part:
// the whole part when it appears or goes, otherwise its count and its list.
func packedPart(t *testing.T, body string, memberID int64) string {
	t.Helper()
	var part string
	for _, id := range []string{packedPartID(memberID), packedSummaryID(memberID), packedListID(memberID)} {
		if strings.Contains(body, `id="`+id+`"`) {
			part += packFragment(t, body, id)
		}
	}
	if part == "" {
		t.Fatalf("reply carries nothing of member %d's packed part:\n%s", memberID, body)
	}
	return part
}

// The member's packed part: a wrapper that is always on the page, the
// collapsible part inside it (only when something is packed), and inside
// that the count and the list, which are what a tap swaps.
func packedPartID(memberID int64) string {
	return fmt.Sprintf("packed-member-%d", memberID)
}

func packedSummaryID(memberID int64) string {
	return fmt.Sprintf("packed-summary-member-%d", memberID)
}

func packedListID(memberID int64) string {
	return fmt.Sprintf("packed-list-member-%d", memberID)
}

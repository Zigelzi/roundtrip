package main

import (
	"html"
	"net/http"
	"strings"
	"testing"
)

// Milestone 11 (spec/milestones/11-in-the-bag.md): each person's packed
// part moves from the end of their list to right under their heading, and
// previews the names of what is packed while it is closed.
//
// The page contract these tests state:
//   - In a person's section, their heading block (pack-head-member-N) comes
//     first, then their packed part's wrapper (packed-member-N), then the
//     list still to pack (topack-member-N). The wrapper is there even when
//     nothing is packed, so the part appears in its place on the first tap.
//   - The packed part's summary (packed-summary-member-N) holds the title
//     "Packed (n)", the preview, and the hint "Tap an item to take it out".
//     The preview is the element marked data-bag-preview; its text is the
//     packed item names, in list order, joined with ", ".

const bagHint = "Tap an item to take it out"

// assertIDsInOrder checks that the elements with the given ids appear in
// html in that order.
func assertIDsInOrder(t *testing.T, html string, ids ...string) {
	t.Helper()
	last := -1
	for _, id := range ids {
		at := strings.Index(html, `id="`+id+`"`)
		if at < 0 {
			t.Errorf("no element with id %q in:\n%s", id, html)
			return
		}
		if at < last {
			t.Errorf("the section is out of order, want %v:\n%s", ids, html)
			return
		}
		last = at
	}
}

// bagPreview is the text of the packed part's preview in a piece of HTML.
func bagPreview(t *testing.T, s string) string {
	t.Helper()
	i := strings.Index(s, "data-bag-preview")
	if i < 0 {
		t.Fatalf("no preview (data-bag-preview) in:\n%s", s)
	}
	start := i + strings.Index(s[i:], ">") + 1
	end := start + strings.Index(s[start:], "<")
	return strings.TrimSpace(html.UnescapeString(s[start:end]))
}

// S1: The packed part sits right under a person's heading.
func TestPackedPartSitsUnderTheHeading(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	parent1 := memberID(t, database, familyMembers[0])
	parent2 := memberID(t, database, familyMembers[1])
	packItem(t, database, insertItem(t, database, tripID, familyMembers[0], "Socks", 1))
	insertItem(t, database, tripID, familyMembers[0], "Hat", 1)
	insertItem(t, database, tripID, familyMembers[0], "Towel", 1)
	insertItem(t, database, tripID, familyMembers[1], "Raincoat", 1)

	page := get(t, app, packURLFor(tripID))

	section := memberSection(t, page, familyMembers[0])
	assertIDsInOrder(t, section, packHeadID(parent1), packedPartID(parent1), toPackListID(parent1))
	assertInOrder(t, section, "Packed (1)", "Hat", "Towel")
	if part := packFragment(t, page, packedPartID(parent1)); detailsIsOpen(part) {
		t.Errorf("the packed part should start closed:\n%s", part)
	}

	// Nothing packed: no part shows, but its place under the heading is kept.
	other := memberSection(t, page, familyMembers[1])
	if strings.Contains(other, "<details") || strings.Contains(other, "Packed (") {
		t.Errorf("Parent 2 has nothing packed but shows a packed part:\n%s", other)
	}
	assertIDsInOrder(t, other, packHeadID(parent2), packedPartID(parent2), toPackListID(parent2))
}

// S2: The packed part appears in its new place on the first tap.
func TestPackedPartAppearsUnderTheHeadingOnTheFirstTap(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	parent1 := memberID(t, database, familyMembers[0])
	hat := insertItem(t, database, tripID, familyMembers[0], "Hat", 1)
	prepareItem(t, database, hat)
	insertItem(t, database, tripID, familyMembers[0], "Towel", 1)

	rec := sendHTMX(t, app, packItemURL(tripID, parent1, hat), partNotShown)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	// The whole part is sent into its wrapper, which sits under the heading.
	if wrapper := packFragment(t, body, packedPartID(parent1)); !strings.Contains(wrapper, "hx-swap-oob") {
		t.Errorf("the packed part is not swapped into its place:\n%s", wrapper)
	}
	part := packedPart(t, body, parent1)
	if !strings.Contains(part, "<details") || !strings.Contains(part, "Packed (1)") || !strings.Contains(part, "Hat") {
		t.Errorf("the packed part did not appear with Hat in it:\n%s", part)
	}
	if toPack := packFragment(t, body, toPackListID(parent1)); strings.Contains(toPack, "Hat") {
		t.Errorf("Hat is still among the things to pack:\n%s", toPack)
	}

	// And on the next visit it is under the heading.
	section := memberSection(t, get(t, app, packURLFor(tripID)), familyMembers[0])
	assertIDsInOrder(t, section, packHeadID(parent1), packedPartID(parent1), toPackListID(parent1))
}

// S3: Opened, the packed part says how to take an item out.
func TestOpenedPackedPartSaysHowToTakeAnItemOut(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	parent1 := memberID(t, database, familyMembers[0])
	packItem(t, database, insertItem(t, database, tripID, familyMembers[0], "Socks", 1))
	insertItem(t, database, tripID, familyMembers[0], "Hat", 1)

	page := get(t, app, packURLFor(tripID))

	summary := packFragment(t, page, packedSummaryID(parent1))
	if !strings.Contains(summary, bagHint) {
		t.Errorf("the packed part does not say %q:\n%s", bagHint, summary)
	}
	// Q2: the chevron, the preview and the hint replace "Show" and "Hide".
	if strings.Contains(summary, ">Show<") || strings.Contains(summary, ">Hide<") {
		t.Errorf("the packed part still says Show or Hide:\n%s", summary)
	}
	row := packRow(t, packFragment(t, page, packedListID(parent1)), "Socks")
	if !strings.Contains(row, `aria-label="Unpack Socks"`) || !strings.Contains(row, "line-through") {
		t.Errorf("Socks is not listed as packed, struck through, ready to take out:\n%s", row)
	}
}

// S4: Taking an item out from under the heading returns it to prepared.
// 08/S5 from the part's new place: the existing 08 tests keep covering the
// details; this checks the scenario end to end.
func TestTakingAnItemOutFromUnderTheHeading(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	parent1 := memberID(t, database, familyMembers[0])
	socks := insertItem(t, database, tripID, familyMembers[0], "Socks", 1)
	insertItem(t, database, tripID, familyMembers[0], "Towel", 1)
	hat := insertItem(t, database, tripID, familyMembers[0], "Hat", 1)
	packItem(t, database, socks)
	packItem(t, database, hat)

	body := sendHTMX(t, app, unpackItemURL(tripID, parent1, socks), partShown).Body.String()

	toPack := packFragment(t, body, toPackListID(parent1))
	assertInOrder(t, toPack, "Socks", "Towel")
	if row := packRow(t, toPack, "Socks"); !strings.Contains(row, `aria-label="Pack Socks"`) || !strings.Contains(row, "Prepared") {
		t.Errorf("Socks is not back as prepared:\n%s", row)
	}
	if summary := packFragment(t, body, packedSummaryID(parent1)); !strings.Contains(summary, "Packed (1)") {
		t.Errorf("the packed part does not read Packed (1):\n%s", summary)
	}
	// Sending the <details> again would close the open part.
	if strings.Contains(body, "<details") {
		t.Errorf("reply contains <details>, which would close the packed part:\n%s", body)
	}

	body = sendHTMX(t, app, unpackItemURL(tripID, parent1, hat), partShown).Body.String()

	if part := packFragment(t, body, packedPartID(parent1)); strings.Contains(part, "<details") || strings.Contains(part, "Packed (") {
		t.Errorf("the packed part is still there after its last item was taken out:\n%s", part)
	}
}

// S7: A person who is all packed.
func TestAllPackedComesAfterThePackedPart(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	child1 := memberID(t, database, familyMembers[2])
	packItem(t, database, insertItem(t, database, tripID, familyMembers[2], "Swimsuit", 1))
	packItem(t, database, insertItem(t, database, tripID, familyMembers[2], "Towel", 1))

	page := get(t, app, packURLFor(tripID))

	section := memberSection(t, page, familyMembers[2])
	assertIDsInOrder(t, section, packHeadID(child1), packedPartID(child1), toPackListID(child1))
	part := packedPart(t, page, child1)
	if !strings.Contains(part, "Packed (2)") {
		t.Errorf("the packed part does not read Packed (2):\n%s", part)
	}
	if got := bagPreview(t, part); got != "Swimsuit, Towel" {
		t.Errorf("preview = %q, want %q", got, "Swimsuit, Towel")
	}
	// The note is in the list below the part, not inside the part.
	if strings.Contains(part, "All packed.") {
		t.Errorf("\"All packed.\" is inside the packed part:\n%s", part)
	}
	if toPack := packFragment(t, page, toPackListID(child1)); !strings.Contains(toPack, "All packed.") {
		t.Errorf("the list below the packed part does not say \"All packed.\":\n%s", toPack)
	}
}

// S5: The closed packed part shows what is in the bag.
func TestClosedPackedPartPreviewsItsNames(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	parent1 := memberID(t, database, familyMembers[0])
	packItem(t, database, insertItem(t, database, tripID, familyMembers[0], "Socks", 3))
	packItem(t, database, insertItem(t, database, tripID, familyMembers[0], "Hat", 1))
	insertItem(t, database, tripID, familyMembers[0], "Towel", 1)

	page := get(t, app, packURLFor(tripID))

	summary := packFragment(t, page, packedSummaryID(parent1))
	if !strings.Contains(summary, "Packed (2)") {
		t.Errorf("the packed part does not read Packed (2):\n%s", summary)
	}
	if got := bagPreview(t, summary); got != "Socks, Hat" {
		t.Errorf("preview = %q, want %q", got, "Socks, Hat")
	}
	if strings.Contains(summary, "pcs") {
		t.Errorf("the preview shows quantities:\n%s", summary)
	}
}

// S6: The preview follows every tap.
func TestBagPreviewFollowsEveryTap(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	parent1 := memberID(t, database, familyMembers[0])
	packItem(t, database, insertItem(t, database, tripID, familyMembers[0], "Socks", 3))
	hat := insertItem(t, database, tripID, familyMembers[0], "Hat", 1)
	packItem(t, database, hat)
	towel := insertItem(t, database, tripID, familyMembers[0], "Towel", 1)
	prepareItem(t, database, towel)

	body := sendHTMX(t, app, packItemURL(tripID, parent1, towel), partShown).Body.String()

	if got := bagPreview(t, packFragment(t, body, packedSummaryID(parent1))); got != "Socks, Hat, Towel" {
		t.Errorf("after packing Towel, preview = %q, want %q", got, "Socks, Hat, Towel")
	}

	body = sendHTMX(t, app, unpackItemURL(tripID, parent1, hat), partShown).Body.String()

	if got := bagPreview(t, packFragment(t, body, packedSummaryID(parent1))); got != "Socks, Towel" {
		t.Errorf("after taking Hat out, preview = %q, want %q", got, "Socks, Towel")
	}
}

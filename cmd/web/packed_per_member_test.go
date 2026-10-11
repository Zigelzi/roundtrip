package main

import (
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"testing"
)

// Milestone 08 (spec/milestones/08-gather-and-pack.md), slice 1, scenarios
// S1 and S2: packed items stay in their owner's section, in a collapsed
// packed part, instead of one shared Packed section at the end of the page.

// S1: Packed items are shown under each member.
func TestPackedItemsAreShownUnderEachMember(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	child1, child2 := familyMembers[2], familyMembers[3]
	packItem(t, database, insertItem(t, database, tripID, child1, "Swimsuit", 1))
	insertItem(t, database, tripID, child1, "Towel", 1)
	insertItem(t, database, tripID, child1, "Sun hat", 1)
	insertItem(t, database, tripID, child2, "Raincoat", 1)

	body := get(t, app, packURLFor(tripID))

	first := memberSection(t, body, child1)
	toPack := between(t, first, `id="`+toPackListID(memberID(t, database, child1))+`"`, "<details")
	for _, want := range []string{"Towel", "Sun hat"} {
		if !strings.Contains(toPack, want) {
			t.Errorf("Child 1's list does not show unpacked %q:\n%s", want, toPack)
		}
	}
	if strings.Contains(toPack, "Swimsuit") {
		t.Errorf("the packed item is still in Child 1's list of things to pack:\n%s", toPack)
	}
	i := strings.Index(first, "<details")
	if i < 0 {
		t.Fatalf("Child 1's section has no packed part:\n%s", first)
	}
	packed := first[i:]
	if !strings.Contains(packed, "Packed (1)") || !strings.Contains(packed, "Swimsuit") {
		t.Errorf("Child 1's section has no packed part holding the packed item with its count:\n%s", packed)
	}
	if detailsIsOpen(packed) {
		t.Errorf("Child 1's packed part should be closed when the page opens:\n%s", packed)
	}

	second := memberSection(t, body, child2)
	if strings.Contains(second, "<details") || strings.Contains(second, "Packed (") {
		t.Errorf("Child 2 has nothing packed, so their section should have no packed part:\n%s", second)
	}

	for _, gone := range []string{`id="packed"`, `id="packed-list"`, `id="packed-summary"`} {
		if strings.Contains(body, gone) {
			t.Errorf("the shared Packed section (%s) is still on the page", gone)
		}
	}
}

// S2: A fully packed member has all their items in their packed part.
func TestFullyPackedMemberHasAllItemsInTheirPackedPart(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	child1 := familyMembers[2]
	names := []string{"Swimsuit", "Towel", "Sun hat"}
	for _, name := range names {
		packItem(t, database, insertItem(t, database, tripID, child1, name, 1))
	}
	insertItem(t, database, tripID, familyMembers[0], "Underwear", 4)

	body := get(t, app, packURLFor(tripID))

	section := memberSection(t, body, child1)
	i := strings.Index(section, "<details")
	if i < 0 {
		t.Fatalf("Child 1's section has no packed part:\n%s", section)
	}
	packed := section[i:]
	for _, name := range names {
		if !strings.Contains(packed, name) {
			t.Errorf("Child 1's packed part is missing %q:\n%s", name, packed)
		}
	}
	if !strings.Contains(packed, "Packed (3)") {
		t.Errorf("Child 1's packed part does not count all three items:\n%s", packed)
	}
	if !strings.Contains(section[:i], "All packed") {
		t.Errorf("Child 1's section does not say Child 1 is packed:\n%s", section)
	}
}

// detailsIsOpen reports whether the first <details> tag in html carries the
// open attribute, which would show a packed part expanded.
func detailsIsOpen(html string) bool {
	i := strings.Index(html, "<details")
	if i < 0 {
		return false
	}
	tag := html[i : i+strings.Index(html[i:], ">")+1]
	return regexp.MustCompile(`\sopen(\s|=|>|/)`).MatchString(tag)
}

// between returns the part of s from the first from up to the next to, or
// to the end of s when to does not follow.
func between(t *testing.T, s, from, to string) string {
	t.Helper()
	i := strings.Index(s, from)
	if i < 0 {
		t.Fatalf("no %q in:\n%s", from, s)
	}
	if j := strings.Index(s[i:], to); j >= 0 {
		return s[i : i+j]
	}
	return s[i:]
}

// Code review, round 1: whether a reply sends a member's packed part whole
// follows what the page that sent the tap shows, not a guess from the
// database. Every tap form says whether its page shows the member's packed
// part. The part is sent whole unless the page shows it and the member still
// has something packed; then only the count and the list are, so an open
// part stays open.

// partShown and partNotShown are the field a tap form sends.
var (
	partShown    = url.Values{"packed_shown": {"1"}}
	partNotShown = url.Values{"packed_shown": {"0"}}
)

// Each tap form says whether its page shows the member's packed part.
func TestTapFormsSayWhetherThePackedPartIsShown(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	child1, child2 := familyMembers[2], familyMembers[3]
	insertItem(t, database, tripID, child1, "Swimsuit", 1)
	insertItem(t, database, tripID, child2, "Towel", 1)
	packItem(t, database, insertItem(t, database, tripID, child2, "Goggles", 1))

	page := get(t, app, packURLFor(tripID))

	notShown := `name="packed_shown" value="0"`
	shown := `name="packed_shown" value="1"`
	if s := memberSection(t, page, child1); !strings.Contains(s, notShown) || strings.Contains(s, shown) {
		t.Errorf("Child 1 has nothing packed, so every tap should say the part is not shown:\n%s", s)
	}
	s := memberSection(t, page, child2)
	if strings.Contains(s, notShown) {
		t.Errorf("Child 2 has a packed part, but a tap says it is not shown:\n%s", s)
	}
	if n := strings.Count(s, shown); n != 2 {
		t.Errorf("%d of Child 2's 2 taps (prepare Towel, unpack Goggles) say the part is shown:\n%s", n, s)
	}
}

// Parent 1's page shows nothing packed for Child 1, but Parent 2 has packed
// "Towel" since. Parent 1 prepares "Swimsuit": the tap works, and the reply
// brings Child 1's whole packed part, since a count and a list alone would
// have nowhere to go on Parent 1's page.
func TestTapFromAPageWithoutThePackedPartBringsItWhole(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	child1 := familyMembers[2]
	id := memberID(t, database, child1)
	swimsuit := insertItem(t, database, tripID, child1, "Swimsuit", 1)
	packItem(t, database, insertItem(t, database, tripID, child1, "Towel", 1))

	rec := sendHTMX(t, app, prepareItemURL(tripID, id, swimsuit), partNotShown)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	part := packedPart(t, rec.Body.String(), id)
	if !strings.Contains(part, "<details") || !strings.Contains(part, "Packed (1)") || !strings.Contains(part, "Towel") {
		t.Errorf("the reply does not bring Child 1's whole packed part, holding Towel:\n%s", part)
	}
}

// A repeated tap that changes nothing (the second of a double tap on the
// back button) keeps an open packed part open.
func TestRepeatedTapKeepsThePackedPartOpen(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	child1 := familyMembers[2]
	id := memberID(t, database, child1)
	swimsuit := insertItem(t, database, tripID, child1, "Swimsuit", 1)
	packItem(t, database, insertItem(t, database, tripID, child1, "Goggles", 1))

	body := sendHTMX(t, app, unprepareItemURL(tripID, id, swimsuit), partShown).Body.String()

	if strings.Contains(body, "<details") {
		t.Errorf("reply contains <details>, which would collapse the open packed part:\n%s", body)
	}
	if !strings.Contains(body, `id="`+packedSummaryID(id)+`"`) {
		t.Errorf("reply does not update the packed part's count:\n%s", body)
	}
}

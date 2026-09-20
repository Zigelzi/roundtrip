package main

import (
	"fmt"
	"strings"
	"testing"
)

// Milestone 02 (spec/milestones/02-pack-items.md), scenario S5: packing one
// person's things without the rest of the family's items in the way.

// S5: Parent packs one person's things only.
func TestPackPageNarrowedToOneMember(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	first := memberID(t, database, familyMembers[0])
	insertItem(t, database, tripID, familyMembers[0], "Underwear", 4)
	insertItem(t, database, tripID, familyMembers[0], "Toothbrush", 1)
	insertItem(t, database, tripID, familyMembers[1], "Raincoat", 1)

	body := get(t, app, narrowedPackURL(tripID, first))

	if section := memberSection(t, body, familyMembers[0]); !strings.Contains(section, "Underwear") {
		t.Errorf("the chosen member's items are missing:\n%s", section)
	}
	if strings.Contains(body, `data-member="`+familyMembers[1]+`"`) {
		t.Errorf("another member's section is still on the page")
	}
	if strings.Contains(body, "Raincoat") {
		t.Errorf("another member's items are still on the page")
	}
	if !strings.Contains(body, "0 of 2 packed") {
		t.Errorf("progress counts more than the chosen member's items")
	}
	if want := fmt.Sprintf(`href="%s"`, packURLFor(tripID)); !strings.Contains(body, want) {
		t.Errorf("no way back to seeing every member (%s)", want)
	}
}

// S5: the narrowing holds while packing — otherwise the first tap would
// throw the parent back to the whole family's list.
func TestPackingWhileNarrowedStaysNarrowed(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	first := memberID(t, database, familyMembers[0])
	underwear := insertItem(t, database, tripID, familyMembers[0], "Underwear", 4)
	insertItem(t, database, tripID, familyMembers[0], "Toothbrush", 1)
	insertItem(t, database, tripID, familyMembers[1], "Raincoat", 1)

	// The rows on a narrowed page carry the narrowing in their action.
	page := get(t, app, narrowedPackURL(tripID, first))
	action := fmt.Sprintf("%s?member=%d", packItemURL(tripID, first, underwear), first)
	if !strings.Contains(page, `action="`+action+`"`) {
		t.Fatalf("a row on the narrowed page does not keep the narrowing (%s)", action)
	}

	body := sendHTMX(t, app, action, nil).Body.String()

	if progress := packFragment(t, body, "pack-progress"); !strings.Contains(progress, "1 of 2 packed") {
		t.Errorf("after a tap the progress is no longer narrowed:\n%s", progress)
	}
	if packed := packFragment(t, body, "packed-list"); strings.Contains(packed, "Raincoat") {
		t.Errorf("after a tap another member's items appeared:\n%s", packed)
	}
}

// A filter naming nobody (a stale link, a hand-edited URL) must not render a
// page with no sections at all -- that reads as an empty trip.
func TestPackPageWithUnknownMemberShowsEveryone(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	insertItem(t, database, tripID, familyMembers[0], "Underwear", 4)

	body := get(t, app, narrowedPackURL(tripID, 999))

	if section := memberSection(t, body, familyMembers[0]); !strings.Contains(section, "Underwear") {
		t.Errorf("an unknown filter hid the trip's items:\n%s", section)
	}
	if !strings.Contains(body, "0 of 1 packed") {
		t.Errorf("an unknown filter left the progress counting nothing")
	}
}

func narrowedPackURL(tripID, memberID int64) string {
	return fmt.Sprintf("%s?member=%d", packURLFor(tripID), memberID)
}

// S1: the parent must be able to tell whose items they are looking at.
func TestPackPageMarksTheActiveFilter(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	first := memberID(t, database, familyMembers[0])
	insertItem(t, database, tripID, familyMembers[0], "Underwear", 4)

	narrowed := get(t, app, narrowedPackURL(tripID, first))
	if want := fmt.Sprintf(`<a href="%s?member=%d" aria-current="page"`, packURLFor(tripID), first); !strings.Contains(narrowed, want) {
		t.Errorf("the member being viewed is not marked as current (%s)", want)
	}
	if want := fmt.Sprintf(`<a href="%s" aria-current="page"`, packURLFor(tripID)); strings.Contains(narrowed, want) {
		t.Errorf("Everyone is marked current while narrowed to one member")
	}

	everyone := get(t, app, packURLFor(tripID))
	if want := fmt.Sprintf(`<a href="%s" aria-current="page"`, packURLFor(tripID)); !strings.Contains(everyone, want) {
		t.Errorf("Everyone is not marked as current on the unnarrowed page (%s)", want)
	}
}

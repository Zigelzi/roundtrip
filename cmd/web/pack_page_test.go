package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

// Milestone 02 (spec/milestones/02-pack-items.md), packing page scenarios
// S1, S6, S7, S8 and S9. Family members are referred to by position rather
// than by name: the real names are configuration (FAMILY_NAMES), not content.

// S1: Parent opens the packing page for a trip.
func TestPackPageListsItemsToPackPerMember(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	parent1, parent2 := familyMembers[0], familyMembers[1]
	insertItem(t, database, tripID, parent1, "Underwear", 4)
	insertItem(t, database, tripID, parent1, "Toothbrush", 1)
	insertItem(t, database, tripID, parent2, "Raincoat", 1)

	tripPage := get(t, app, tripURLFor(tripID))
	if want := fmt.Sprintf(`href="%s"`, packURLFor(tripID)); !strings.Contains(tripPage, want) {
		t.Fatalf("trip page has no link to the packing page (%s)", want)
	}

	body := get(t, app, packURLFor(tripID))

	first := memberSection(t, body, parent1)
	for _, want := range []string{"Underwear", "4 pcs", "Toothbrush", "1 pc"} {
		if !strings.Contains(first, want) {
			t.Errorf("first parent's section does not show %q:\n%s", want, first)
		}
	}
	if second := memberSection(t, body, parent2); !strings.Contains(second, "Raincoat") {
		t.Errorf("second parent's section does not show Raincoat:\n%s", second)
	}
	if !strings.Contains(body, "0 of 3 packed") {
		t.Errorf("page does not show that 0 of 3 items are packed")
	}
	for _, member := range familyMembers {
		want := fmt.Sprintf(`href="%s?member=%d"`, packURLFor(tripID), memberID(t, database, member))
		if !strings.Contains(body, want) {
			t.Errorf("page has no way to narrow to one member (%s)", want)
		}
	}
	if want := fmt.Sprintf(`href="%s"`, tripURLFor(tripID)); !strings.Contains(body, want) {
		t.Errorf("page has no link back to the trip page (%s)", want)
	}
}

// S6: Parent sees when everything is packed.
func TestPackPageSaysWhenEverythingIsPacked(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	for _, name := range []string{"Underwear", "Toothbrush", "Raincoat"} {
		packItem(t, database, insertItem(t, database, tripID, familyMembers[0], name, 1))
	}

	body := get(t, app, packURLFor(tripID))

	if !strings.Contains(body, "Everything is packed") {
		t.Errorf("page does not say everything is packed:\n%s", body)
	}
	if !strings.Contains(body, "3 of 3 packed") {
		t.Errorf("page does not show that 3 of 3 items are packed")
	}
}

// S7: Parent opens the packing page for a trip with no items.
func TestPackPageWithNoItems(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)

	body := get(t, app, packURLFor(tripID))

	if !strings.Contains(body, "Nothing to pack yet") {
		t.Errorf("page does not say there is nothing to pack yet:\n%s", body)
	}
	if want := fmt.Sprintf(`href="%s"`, tripURLFor(tripID)); !strings.Contains(body, want) {
		t.Errorf("page has no link to the trip page to add items (%s)", want)
	}
}

// S8: Parent opens the packing page for a trip that does not exist.
func TestPackPageForMissingTrip(t *testing.T) {
	app, _ := newTestApp(t)

	for _, path := range []string{"/trips/999/pack", "/trips/not-a-number/pack"} {
		rec := send(t, app, http.MethodGet, path, nil)

		if rec.Code != http.StatusNotFound {
			t.Errorf("GET %s: status = %d, want 404", path, rec.Code)
		}
		body := rec.Body.String()
		if !strings.Contains(body, "Trip not found") {
			t.Errorf("GET %s: page does not say the trip was not found", path)
		}
		if !strings.Contains(body, `href="/"`) {
			t.Errorf("GET %s: page has no link back to the front page", path)
		}
	}
}

// S9: A member with nothing planned reads differently from one who is done.
// Both show an empty list of things to pack, so the wording is the only thing
// telling "nothing planned for them" apart from "they are finished".
func TestPackPageTellsNothingPlannedFromFullyPacked(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	packed, empty := familyMembers[2], familyMembers[3]
	packItem(t, database, insertItem(t, database, tripID, packed, "Pyjamas", 1))
	insertItem(t, database, tripID, familyMembers[0], "Underwear", 4)

	body := get(t, app, packURLFor(tripID))

	if section := memberSection(t, body, empty); !strings.Contains(section, "Nothing planned") {
		t.Errorf("a member with no items should say nothing is planned:\n%s", section)
	}
	if section := memberSection(t, body, packed); !strings.Contains(section, "All packed") {
		t.Errorf("a member whose items are all packed should say so:\n%s", section)
	}
}

func packURLFor(id int64) string {
	return fmt.Sprintf("/trips/%d/pack", id)
}

// packItem marks an already-inserted item packed, standing in for the parent
// having ticked it on an earlier visit.
func packItem(t *testing.T, database *sql.DB, itemID int64) {
	t.Helper()
	if _, err := database.Exec("UPDATE item SET status = 'packed' WHERE id = ?", itemID); err != nil {
		t.Fatalf("pack item %d: %v", itemID, err)
	}
}

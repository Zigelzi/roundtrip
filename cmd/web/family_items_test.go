package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/Zigelzi/roundtrip/internal/db"
)

// Milestone 03 (spec/milestones/03-family-items.md), trip page half:
// S1, S5, S6, S7, and S2's section-ordering assertion. The packing page half
// of these scenarios belongs to a later slice.

// S1: Parent adds an item that belongs to the whole family.
func TestAddFamilyItem(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	family := memberID(t, database, "Family")
	if family != familyBucketID {
		t.Fatalf("Family bucket id = %d, want %d", family, familyBucketID)
	}

	rec := addItem(t, app, tripID, family, "Sunscreen", "1")

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want 303 back to the trip", rec.Code)
	}
	body := get(t, app, tripURLFor(tripID))
	section := memberSection(t, body, "Family")
	if !strings.Contains(section, "Sunscreen") || !strings.Contains(section, "1 pc") {
		t.Errorf("Family section does not list Sunscreen, 1 pc:\n%s", section)
	}
	for _, person := range familyMembers {
		if strings.Contains(memberSection(t, body, person), "Sunscreen") {
			t.Errorf("Sunscreen also shows up under %s, want only the Family section", person)
		}
	}
}

// S5: Parent removes a family item (the trip page half; the packing page
// half is a later slice).
func TestRemoveFamilyItem(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	boardGame := insertItem(t, database, tripID, "Family", "Board game", 1)

	rec := send(t, app, http.MethodPost, fmt.Sprintf("/trips/%d/items/%d/delete", tripID, boardGame), nil)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want 303", rec.Code)
	}
	section := memberSection(t, get(t, app, tripURLFor(tripID)), "Family")
	if strings.Contains(section, "Board game") {
		t.Errorf("Family section still lists Board game:\n%s", section)
	}
}

// S6 (trip page half): the Family section is always shown, even on a trip
// where nothing belongs to the family, the same way an empty member section
// always shows on milestone 01's trip page.
func TestFamilySectionAlwaysShownOnTripPage(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	insertItem(t, database, tripID, "Parent 1", "Underwear", 4)

	body := get(t, app, tripURLFor(tripID))

	section := memberSection(t, body, "Family")
	if !strings.Contains(section, "No items yet") {
		t.Errorf("Family section does not say it is empty:\n%s", section)
	}
}

// S2 (trip page half): the Family section comes after all four people's
// sections.
func TestFamilySectionComesLastOnTripPage(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	insertItem(t, database, tripID, "Family", "Sunscreen", 1)

	body := get(t, app, tripURLFor(tripID))

	lastPerson := strings.Index(body, `data-member="Child 2"`)
	family := strings.Index(body, `data-member="Family"`)
	if lastPerson < 0 || family < 0 {
		t.Fatalf("could not find both Child 2's and Family's sections")
	}
	if family < lastPerson {
		t.Errorf("Family's section (offset %d) comes before Child 2's (offset %d)", family, lastPerson)
	}
}

// S7: The same family item cannot be added twice.
func TestAddDuplicateFamilyItemRejected(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	insertItem(t, database, tripID, "Family", "Sunscreen", 1)

	rec := addItem(t, app, tripID, memberID(t, database, "Family"), "sunscreen", "1")

	if rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want 422", rec.Code)
	}
	if n := count(t, database, "item"); n != 1 {
		t.Errorf("%d items saved, want 1", n)
	}
	section := memberSection(t, rec.Body.String(), "Family")
	if !strings.Contains(section, "Family already has") {
		t.Errorf("Family section does not say Family already has it:\n%s", section)
	}
}

// ListPeople must exclude the Family bucket. Milestone 04's activity
// expansion means "each person" by it, and a bucket in that list would hand
// the family a copy of every personal item. sql/queries/item.sql names the id
// as a literal, because sqlc cannot reach the familyBucketID constant, so
// this test is the only thing keeping the query and the constant in step.
func TestListPeopleExcludesTheFamilyBucket(t *testing.T) {
	_, database := newTestApp(t)

	people, err := db.New(database).ListPeople(context.Background())
	if err != nil {
		t.Fatalf("ListPeople: %v", err)
	}

	if len(people) != len(familyMembers) {
		t.Errorf("ListPeople returned %d rows, want the %d people", len(people), len(familyMembers))
	}
	for _, p := range people {
		if p.ID == familyBucketID {
			t.Errorf("ListPeople returned the Family bucket (id %d) as a person", p.ID)
		}
	}
}

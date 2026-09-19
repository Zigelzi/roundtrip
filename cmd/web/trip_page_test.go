package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

// Milestone 01 (spec/milestones/01-packing-list.md), trip page scenarios S7–S8.

var familyMembers = []string{"Parent 1", "Parent 2", "Child 1", "Child 2"}

// S7: Parent views a trip.
func TestViewTrip(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	insertItem(t, database, tripID, "Parent 1", "Underwear", 4)
	insertItem(t, database, tripID, "Parent 1", "Sunscreen", 1)

	body := get(t, app, tripURLFor(tripID))

	for _, want := range []string{"Parainen", "20.–22.9.2026", "3 days"} {
		if !strings.Contains(body, want) {
			t.Errorf("trip page does not show %q", want)
		}
	}
	for _, member := range familyMembers {
		section := memberSection(t, body, member)
		if member == "Parent 1" {
			if !strings.Contains(section, "Underwear") || !strings.Contains(section, "4 pcs") {
				t.Errorf("Parent 1's section does not list Underwear, 4 pcs:\n%s", section)
			}
			// Quantity is written out, not "× 4", so it can't be mistaken
			// for the ✕ remove button; a single item reads "1 pc".
			if !strings.Contains(section, "1 pc") || strings.Contains(section, "1 pcs") {
				t.Errorf("Parent 1's Sunscreen should read \"1 pc\":\n%s", section)
			}
		} else if !strings.Contains(section, "No items yet") {
			t.Errorf("%s's section does not say they have no items yet", member)
		}
		inputTag(t, section, "name") // fails the test if the add form is missing
		if qty := inputTag(t, section, "quantity"); !strings.Contains(qty, `value="1"`) {
			t.Errorf("%s's add form quantity does not default to 1: %s", member, qty)
		}
	}
}

// S8: Parent opens a trip that does not exist.
func TestViewMissingTrip(t *testing.T) {
	app, _ := newTestApp(t)

	for _, path := range []string{"/trips/999", "/trips/not-a-number"} {
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

func tripURLFor(id int64) string {
	return fmt.Sprintf("/trips/%d", id)
}

func memberID(t *testing.T, database *sql.DB, name string) int64 {
	t.Helper()
	var id int64
	if err := database.QueryRow("SELECT id FROM family_member WHERE name = ?", name).Scan(&id); err != nil {
		t.Fatalf("family member %q: %v", name, err)
	}
	return id
}

func insertItem(t *testing.T, database *sql.DB, tripID int64, member, name string, quantity int) int64 {
	t.Helper()
	res, err := database.Exec("INSERT INTO item (trip_id, family_member_id, name, name_key, quantity) VALUES (?, ?, ?, ?, ?)",
		tripID, memberID(t, database, member), name, strings.ToLower(name), quantity)
	if err != nil {
		t.Fatalf("insert item: %v", err)
	}
	id, _ := res.LastInsertId()
	return id
}

// memberSection returns the HTML of one family member's section on the trip
// page. The page contract: each member is a <section data-member="Name">.
func memberSection(t *testing.T, body, member string) string {
	t.Helper()
	i := strings.Index(body, `data-member="`+member+`"`)
	if i < 0 {
		t.Fatalf("no section for family member %q", member)
	}
	start := strings.LastIndex(body[:i], "<section")
	end := strings.Index(body[i:], "</section>")
	if start < 0 || end < 0 {
		t.Fatalf("section for %q is malformed", member)
	}
	return body[start : i+end]
}

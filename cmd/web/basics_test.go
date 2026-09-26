package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"
)

// Milestone 04 (spec/milestones/04-trip-basics.md), trip basics S1 to S7.

// basicRow is one row of the catalogue's Basics section: the owner's
// family_member id (1 to 4 for the people, 100 for the Family bucket), the
// item name, and its quantity as perDay x trip days + fixed.
type basicRow struct {
	member int64
	name   string
	perDay int
	fixed  int
}

// catalogueBasics is a hand-typed copy of spec/activity-catalogue.md's Basics
// section as it stood on 2026-09-25, in catalogue order. It is deliberately
// not read from the database: the test must prove the seed matches the
// catalogue, not that the code copies its own table. Changing the basics means
// changing this list too.
var catalogueBasics = []basicRow{
	// Parent 1
	{1, "Underwear", 1, 1},
	{1, "Socks", 1, 1},
	{1, "T-shirt", 1, 0},
	{1, "Sweatpants", 0, 1},
	{1, "Khakis", 0, 1},
	{1, "Belt", 0, 1},
	{1, "Long-sleeved shirt", 0, 1},
	{1, "Hoodie", 0, 1},
	{1, "Outdoor jacket", 0, 1},
	{1, "Hat", 0, 1},
	{1, "Gloves", 0, 1},
	{1, "Shoes", 0, 2},
	{1, "Toothbrush", 0, 1},
	{1, "Water bottle", 0, 1},
	{1, "Phone charger", 0, 1},
	{1, "Wallet and keys", 0, 1},
	{1, "Deodorant", 0, 1},
	// Parent 2
	{2, "Underwear", 1, 1},
	{2, "Socks", 1, 2},
	{2, "T-shirt", 1, 0},
	{2, "Trousers", 0, 2},
	{2, "Jumper", 0, 1},
	{2, "Long-sleeved shirt", 0, 1},
	{2, "Pyjamas", 0, 1},
	{2, "Outdoor jacket", 0, 1},
	{2, "Hat", 0, 1},
	{2, "Gloves", 0, 1},
	{2, "Shoes", 0, 2},
	{2, "Toothbrush", 0, 1},
	{2, "Hairbrush", 0, 1},
	{2, "Water bottle", 0, 1},
	{2, "Phone charger", 0, 1},
	{2, "Wallet and keys", 0, 1},
	// Child 1
	{3, "Underwear", 2, 1},
	{3, "Socks", 1, 1},
	{3, "Outfit", 1, 1},
	{3, "Long-sleeved shirt", 0, 2},
	{3, "Hat", 0, 1},
	{3, "Gloves", 0, 1},
	{3, "Pyjama dress", 0, 1},
	{3, "Pyjama pants", 0, 1},
	{3, "Pyjama socks", 0, 1},
	{3, "Outdoor jacket", 0, 1},
	{3, "Outdoor overalls", 0, 1},
	{3, "Shoes", 0, 2},
	{3, "Toothbrush", 0, 1},
	{3, "Water bottle", 0, 1},
	{3, "Small towel", 0, 1},
	{3, "Sleep toy", 0, 4},
	{3, "Hairbrush", 0, 1},
	{3, "Hair spray", 0, 1},
	{3, "Hair tie", 1, 0},
	// Child 2
	{4, "Socks", 1, 1},
	{4, "Outfit", 1, 1},
	{4, "Long-sleeved shirt", 0, 2},
	{4, "Hat", 0, 1},
	{4, "Gloves", 0, 1},
	{4, "Pyjamas", 0, 1},
	{4, "Outdoor jacket", 0, 1},
	{4, "Outdoor overalls", 0, 1},
	{4, "Shoes", 0, 2},
	{4, "Toothbrush", 0, 1},
	{4, "Water bottle", 0, 1},
	{4, "Sleep toy", 0, 1},
	{4, "Sleeping bag", 0, 1},
	{4, "Pacifier", 0, 2},
	{4, "Small towel", 0, 1},
	{4, "Nappies", 6, 0},
	{4, "Hairbrush", 0, 1},
	// Family
	{familyBucketID, "Kids toothpaste", 0, 1},
	{familyBucketID, "Adults toothpaste", 0, 1},
	{familyBucketID, "Wet wipes", 0, 1},
	{familyBucketID, "Painkillers", 0, 1},
	{familyBucketID, "Double stroller", 0, 1},
	{familyBucketID, "Stroller rain cover", 0, 2},
}

// tripItem is one stored item, in the order the trip page lists them.
type tripItem struct {
	member   int64
	name     string
	quantity int
	status   string
}

// createTripOf posts the new trip form for a trip of days days and returns
// the new trip's id, failing unless the app redirects to it.
func createTripOf(t *testing.T, app *application, database *sql.DB, days int) int64 {
	t.Helper()
	rec := send(t, app, http.MethodPost, "/trips", url.Values{
		"destination":    {"Parainen"},
		"departure_date": {"2026-09-20"},
		"duration_days":  {strconv.Itoa(days)},
	})
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("create %d-day trip: status = %d, want 303", days, rec.Code)
	}
	var id int64
	if err := database.QueryRow("SELECT max(id) FROM trip").Scan(&id); err != nil {
		t.Fatalf("read new trip id: %v", err)
	}
	return id
}

// tripItems returns a trip's items in page order: owners in family_member id
// order (people first, Family last), each owner's items in insertion order.
func tripItems(t *testing.T, database *sql.DB, tripID int64) []tripItem {
	t.Helper()
	rows, err := database.Query(`SELECT family_member_id, name, quantity, status
		FROM item WHERE trip_id = ? ORDER BY family_member_id, id`, tripID)
	if err != nil {
		t.Fatalf("list items: %v", err)
	}
	defer rows.Close()
	var items []tripItem
	for rows.Next() {
		var it tripItem
		if err := rows.Scan(&it.member, &it.name, &it.quantity, &it.status); err != nil {
			t.Fatalf("scan item: %v", err)
		}
		items = append(items, it)
	}
	return items
}

// quantityOf returns the quantity of member's item called name on the trip.
func quantityOf(t *testing.T, database *sql.DB, tripID, member int64, name string) int {
	t.Helper()
	var q int
	err := database.QueryRow("SELECT quantity FROM item WHERE trip_id = ? AND family_member_id = ? AND name = ?",
		tripID, member, name).Scan(&q)
	if err != nil {
		t.Fatalf("trip %d has no %q for member %d: %v", tripID, name, member, err)
	}
	return q
}

func itemIDOf(t *testing.T, database *sql.DB, tripID, member int64, name string) int64 {
	t.Helper()
	var id int64
	err := database.QueryRow("SELECT id FROM item WHERE trip_id = ? AND family_member_id = ? AND name = ?",
		tripID, member, name).Scan(&id)
	if err != nil {
		t.Fatalf("trip %d has no %q for member %d: %v", tripID, name, member, err)
	}
	return id
}

// S1: A new trip already holds everyone's basics.
func TestNewTripHoldsBasics(t *testing.T) {
	app, database := newTestApp(t)

	tripID := createTripOf(t, app, database, 3)

	var want []tripItem
	for _, b := range catalogueBasics {
		want = append(want, tripItem{b.member, b.name, b.perDay*3 + b.fixed, "planned"})
	}
	got := tripItems(t, database, tripID)
	if len(got) != 75 || len(want) != 75 {
		t.Fatalf("trip holds %d items, want 75 (expected list has %d)", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("item %d = %+v, want %+v", i+1, got[i], want[i])
		}
	}
}

// S2: Per-day items are sized to the trip's length.
// S3: Fixed items do not scale.
func TestBasicsSizedToTripLength(t *testing.T) {
	cases := []struct {
		days                                int
		p1Underwear, c1Underwear, c2Nappies int
	}{
		{1, 2, 3, 6},
		{3, 4, 7, 18},
		{7, 8, 15, 42},
		{14, 15, 29, 84},
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("%d days", c.days), func(t *testing.T) {
			app, database := newTestApp(t)
			tripID := createTripOf(t, app, database, c.days)

			// S2
			if q := quantityOf(t, database, tripID, 1, "Underwear"); q != c.p1Underwear {
				t.Errorf("Parent 1 Underwear = %d, want %d", q, c.p1Underwear)
			}
			if q := quantityOf(t, database, tripID, 3, "Underwear"); q != c.c1Underwear {
				t.Errorf("Child 1 Underwear = %d, want %d", q, c.c1Underwear)
			}
			if q := quantityOf(t, database, tripID, 4, "Nappies"); q != c.c2Nappies {
				t.Errorf("Child 2 Nappies = %d, want %d", q, c.c2Nappies)
			}
			// S3 (checked for 1, 3 and 14 days; 7 costs nothing extra)
			if q := quantityOf(t, database, tripID, 1, "Shoes"); q != 2 {
				t.Errorf("Parent 1 Shoes = %d, want 2", q)
			}
			if q := quantityOf(t, database, tripID, familyBucketID, "Double stroller"); q != 1 {
				t.Errorf("Family Double stroller = %d, want 1", q)
			}
		})
	}
}

// S5: Basics behave like any other item.
func TestBasicsBehaveLikeOtherItems(t *testing.T) {
	app, database := newTestApp(t)
	tripID := createTripOf(t, app, database, 3)

	belt := itemIDOf(t, database, tripID, 1, "Belt")
	toothbrush := itemIDOf(t, database, tripID, 1, "Toothbrush")

	if rec := send(t, app, http.MethodPost, fmt.Sprintf("/trips/%d/items/%d/delete", tripID, belt), nil); rec.Code >= 400 {
		t.Fatalf("remove Belt: status = %d", rec.Code)
	}
	if rec := send(t, app, http.MethodPost, packItemURL(tripID, 1, toothbrush), nil); rec.Code >= 400 {
		t.Fatalf("pack Toothbrush: status = %d", rec.Code)
	}
	if rec := addItem(t, app, tripID, familyBucketID, "Sunscreen", "1"); rec.Code >= 400 {
		t.Fatalf("add Sunscreen: status = %d", rec.Code)
	}

	var n int
	database.QueryRow("SELECT count(*) FROM item WHERE id = ?", belt).Scan(&n)
	if n != 0 {
		t.Errorf("Belt was not removed")
	}
	if s := itemStatus(t, database, toothbrush); s != "packed" {
		t.Errorf("Toothbrush status = %q, want packed", s)
	}
	if q := quantityOf(t, database, tripID, familyBucketID, "Sunscreen"); q != 1 {
		t.Errorf("Family Sunscreen = %d, want 1", q)
	}
	if body := get(t, app, packURLFor(tripID)); !strings.Contains(body, "1 of 75 packed") {
		t.Errorf("packing page does not show 1 of 75 packed")
	}
}

// S6: Existing trips are left alone.
func TestExistingTripsGetNoBasics(t *testing.T) {
	app, database := newTestApp(t)
	// A trip from before this milestone never went through trip creation.
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)

	get(t, app, tripURLFor(tripID))

	if items := tripItems(t, database, tripID); len(items) != 0 {
		t.Errorf("existing trip has %d items, want 0", len(items))
	}
}

// S7: A failed trip creation leaves nothing behind.
func TestFailedTripCreationLeavesNothing(t *testing.T) {
	app, database := newTestApp(t)
	// A second "underwear" for Parent 1 collides with "Underwear" under the
	// item table's per-owner unique name, so saving the basics fails partway.
	if _, err := database.Exec(`INSERT INTO basic_item (family_member_id, position, name, per_day, fixed)
		VALUES (1, 999, 'underwear', 0, 1)`); err != nil {
		t.Fatalf("seed a clashing basics row: %v", err)
	}

	rec := send(t, app, http.MethodPost, "/trips", url.Values{
		"destination":    {"Parainen"},
		"departure_date": {"2026-09-20"},
		"duration_days":  {"3"},
	})

	if rec.Code == http.StatusSeeOther {
		t.Fatalf("status = 303: the trip was reported as created")
	}
	body := rec.Body.String()
	if !strings.Contains(body, "The trip could not be created. Please try again.") {
		t.Errorf("page does not explain that the trip was not created")
	}
	if !strings.Contains(inputTag(t, body, "destination"), `value="Parainen"`) {
		t.Errorf("the form lost what the parent entered")
	}
	if n := count(t, database, "trip"); n != 0 {
		t.Errorf("%d trips saved, want 0", n)
	}
	if n := count(t, database, "item"); n != 0 {
		t.Errorf("%d items saved, want 0", n)
	}
}

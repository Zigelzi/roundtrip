package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

// Milestone 05 (spec/milestones/05-ui-pass.md), the automated halves of S5
// and S6. The rest of the milestone is how the pages look on a phone and is
// checked there.

// S5: Parent jumps to a person's items.
func TestTripHeaderJumpsToEachPerson(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)

	body := get(t, app, tripURLFor(tripID))
	start, end := strings.Index(body, "<header"), strings.Index(body, "</header>")
	if start < 0 || end < start {
		t.Fatalf("trip page has no <header>")
	}
	header := body[start:end]

	// One link per section, in the order the sections appear on the page,
	// each pointing at its section's id.
	last := -1
	for _, member := range append(familyMembers, "Family") {
		link := fmt.Sprintf(`href="#member-%d"`, memberID(t, database, member))
		i := strings.Index(header, link)
		if i < 0 {
			t.Errorf("header has no link to %s's items (%s)", member, link)
			continue
		}
		if a, _, _ := strings.Cut(header[i:], "</a>"); !strings.Contains(a, ">"+member+"<") {
			t.Errorf("link to %s's items is not labelled with their name", member)
		}
		if i < last {
			t.Errorf("link to %s is out of page order", member)
		}
		last = i
		if want := fmt.Sprintf(`id="member-%d"`, memberID(t, database, member)); !strings.Contains(memberSection(t, body, member), want) {
			t.Errorf("%s's section has no %s for the link to land on", member, want)
		}
	}
}

// S6: Parent sees how many items each person has.
func TestMemberHeadingShowsItemCount(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	insertItem(t, database, tripID, "Parent 1", "Underwear", 4)
	insertItem(t, database, tripID, "Parent 1", "Socks", 4)
	insertItem(t, database, tripID, "Child 1", "Hat", 1)

	body := get(t, app, tripURLFor(tripID))

	for member, want := range map[string]string{"Parent 1": "2 items", "Child 1": "1 item", "Child 2": "0 items"} {
		heading := memberHeading(t, memberSection(t, body, member))
		if !strings.Contains(heading, ">"+want+"</span>") {
			t.Errorf("%s's heading does not say %q: %s", member, want, heading)
		}
	}
}

// S6: the count stays right when an item is added or removed in place, which
// replaces only that person's part of the page.
func TestItemCountUpdatesInPlace(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	parent := memberID(t, database, "Parent 1")
	socks := insertItem(t, database, tripID, "Parent 1", "Socks", 4)

	rec := sendHTMX(t, app, fmt.Sprintf("/trips/%d/members/%d/items", tripID, parent), url.Values{"name": {"Underwear"}, "quantity": {"4"}})
	if rec.Code != http.StatusOK {
		t.Fatalf("add: status = %d, want 200", rec.Code)
	}
	if body := rec.Body.String(); !strings.Contains(memberHeading(t, body), ">2 items</span>") {
		t.Errorf("reply to adding does not carry the new count, 2 items:\n%s", body)
	}

	rec = sendHTMX(t, app, fmt.Sprintf("/trips/%d/items/%d/delete", tripID, socks), nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("remove: status = %d, want 200", rec.Code)
	}
	if body := rec.Body.String(); !strings.Contains(memberHeading(t, body), ">1 item</span>") {
		t.Errorf("reply to removing does not carry the new count, 1 item:\n%s", body)
	}
}

// memberHeading is the first <h2> in a piece of the page.
func memberHeading(t *testing.T, s string) string {
	t.Helper()
	start := strings.Index(s, "<h2")
	end := strings.Index(s, "</h2>")
	if start < 0 || end < start {
		t.Fatalf("no <h2> heading in:\n%s", s)
	}
	return s[start:end]
}

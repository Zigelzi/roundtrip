package main

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// Milestone 10 (spec/milestones/10-packing-progress.md): packing progress as
// a bar for the trip, a bar and a count under each person's heading, and a
// count in every filter pill.
//
// The page contract these tests state:
//   - The trip bar is the element with id "pack-trip-bar", inside the
//     packing page's header. Each person's heading block (their <h2> and
//     their bar) has id "pack-head-member-N". The filter row has id
//     "pack-filter". A tap's reply carries all three out of band, the
//     heading only for the person whose item moved.
//   - A bar is two segments, data-bar="packed" and data-bar="prepared", each
//     with its width as a percentage in an inline style. An empty segment
//     may be left out.

const (
	tripBarID    = "pack-trip-bar"
	packFilterID = "pack-filter"
)

func packHeadID(memberID int64) string {
	return fmt.Sprintf("pack-head-member-%d", memberID)
}

var barWidthRe = regexp.MustCompile(`width:\s*([0-9.]+)%`)

// barWidths reads the first bar in html: the width of its packed and its
// prepared segment, in percent. A missing segment counts as 0.
func barWidths(t *testing.T, html string) (packed, prepared float64) {
	t.Helper()
	if !strings.Contains(html, `data-bar=`) {
		t.Fatalf("no bar in:\n%s", html)
	}
	return segmentWidth(t, html, "packed"), segmentWidth(t, html, "prepared")
}

func segmentWidth(t *testing.T, html, segment string) float64 {
	t.Helper()
	if !strings.Contains(html, `data-bar="`+segment+`"`) {
		return 0
	}
	tag := tagWith(t, html, `data-bar="`+segment+`"`, "<")
	m := barWidthRe.FindStringSubmatch(tag)
	if m == nil {
		t.Fatalf("bar segment %q has no width: %s", segment, tag)
	}
	w, err := strconv.ParseFloat(m[1], 64)
	if err != nil {
		t.Fatalf("bar segment %q width %q: %v", segment, m[1], err)
	}
	return w
}

func wantBar(t *testing.T, where, html string, packed, prepared float64) {
	t.Helper()
	gotPacked, gotPrepared := barWidths(t, html)
	if gotPacked != packed || gotPrepared != prepared {
		t.Errorf("%s: bar is %v%% packed and %v%% prepared, want %v%% and %v%%:\n%s", where, gotPacked, gotPrepared, packed, prepared, html)
	}
}

// filterPill is the filter row's link to href, from its opening tag to its
// closing one. The Everyone link's href is the packing page itself.
func filterPill(t *testing.T, body, href string) string {
	t.Helper()
	i := strings.Index(body, `<a href="`+href+`"`)
	if i < 0 {
		t.Fatalf("no filter link to %s", href)
	}
	end := strings.Index(body[i:], "</a>")
	if end < 0 {
		t.Fatalf("filter link to %s is not closed", href)
	}
	return body[i : i+end]
}

// wantPillCount checks that a pill shows short, "1/3", to the eye only, and
// speaks long, "1 of 3 packed", to a screen reader (S6).
func wantPillCount(t *testing.T, name, pill, short, long string) {
	t.Helper()
	if !strings.Contains(pill, short+"</span>") {
		t.Errorf("%s pill does not show %q:\n%s", name, short, pill)
		return
	}
	if span := enclosingTag(t, pill, short+"</span>", "<span"); !strings.Contains(span, `aria-hidden="true"`) {
		t.Errorf("%s pill's %q is read aloud as it is; want it hidden from screen readers: %s", name, short, span)
	}
	if !strings.Contains(pill, long) {
		t.Errorf("%s pill does not say %q to a screen reader:\n%s", name, long, pill)
	}
}

func wantOOB(t *testing.T, name, fragment string) {
	t.Helper()
	if open := fragment[:strings.Index(fragment, ">")+1]; !strings.Contains(open, "hx-swap-oob") {
		t.Errorf("%s is in the reply but not swapped out of band: %s", name, open)
	}
}

// S1: The header shows how far the whole trip has come.
func TestPackingHeaderShowsTripBar(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	packItem(t, database, insertItem(t, database, tripID, "Parent 1", "Socks", 4))
	prepareItem(t, database, insertItem(t, database, tripID, "Parent 1", "Hat", 1))
	insertItem(t, database, tripID, "Parent 2", "Raincoat", 1)
	insertItem(t, database, tripID, "Family", "Sunscreen", 1)

	header := pageHeader(t, get(t, app, packURLFor(tripID)))

	if !strings.Contains(header, "Packing · 1 of 4 packed") {
		t.Errorf("header does not read \"Packing · 1 of 4 packed\":\n%s", header)
	}
	if !strings.Contains(header, `id="`+tripBarID+`"`) {
		t.Fatalf("the trip bar is not inside the header:\n%s", header)
	}
	bar := packFragment(t, header, tripBarID)
	wantBar(t, "trip bar", bar, 25, 25)
	if open := bar[:strings.Index(bar, ">")+1]; !strings.Contains(open, `aria-hidden="true"`) {
		t.Errorf("the trip bar is read aloud; the text beside it carries the numbers: %s", open)
	}
}

// S2: The trip bar follows every tap.
func TestTripBarFollowsEveryTap(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	parent := memberID(t, database, "Parent 1")
	packItem(t, database, insertItem(t, database, tripID, "Parent 1", "Socks", 4))
	hat := insertItem(t, database, tripID, "Parent 1", "Hat", 1)
	prepareItem(t, database, hat)
	towel := insertItem(t, database, tripID, "Parent 1", "Towel", 1)
	insertItem(t, database, tripID, "Parent 2", "Raincoat", 1)

	body := sendHTMX(t, app, prepareItemURL(tripID, parent, towel), url.Values{"packed_shown": {"1"}}).Body.String()
	bar := packFragment(t, body, tripBarID)
	wantOOB(t, "trip bar", bar)
	wantBar(t, "after preparing Towel", bar, 25, 50)

	body = sendHTMX(t, app, packItemURL(tripID, parent, hat), url.Values{"packed_shown": {"1"}}).Body.String()
	wantBar(t, "after packing Hat", packFragment(t, body, tripBarID), 50, 25)
	if progress := packFragment(t, body, "pack-progress"); !strings.Contains(progress, "Packing · 2 of 4 packed") {
		t.Errorf("header line does not count the newly packed item:\n%s", progress)
	}
}

// S3: Narrowed to one person, the header still shows the whole trip.
func TestNarrowedHeaderShowsTheWholeTrip(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	parent := memberID(t, database, "Parent 1")
	insertItem(t, database, tripID, "Parent 1", "Socks", 4)
	insertItem(t, database, tripID, "Parent 1", "Hat", 1)
	packItem(t, database, insertItem(t, database, tripID, "Parent 2", "Raincoat", 1))
	insertItem(t, database, tripID, "Parent 2", "Towel", 1)
	insertItem(t, database, tripID, "Family", "Sunscreen", 1)

	body := get(t, app, narrowedPackURL(tripID, parent))

	if header := pageHeader(t, body); !strings.Contains(header, "Packing · 1 of 5 packed") {
		t.Errorf("narrowed header does not count the whole trip, \"Packing · 1 of 5 packed\":\n%s", header)
	}
	wantBar(t, "narrowed trip bar", packFragment(t, body, tripBarID), 20, 0)
	if heading := memberHeading(t, memberSection(t, body, "Parent 1")); !strings.Contains(heading, "0 of 2 packed") {
		t.Errorf("Parent 1's heading does not read \"0 of 2 packed\":\n%s", heading)
	}
}

// S4: Each person's heading shows their own progress.
func TestPersonHeadingShowsProgress(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	parent := memberID(t, database, "Parent 1")
	parent2 := memberID(t, database, "Parent 2")
	packItem(t, database, insertItem(t, database, tripID, "Parent 1", "Socks", 4))
	prepareItem(t, database, insertItem(t, database, tripID, "Parent 1", "Hat", 1))
	insertItem(t, database, tripID, "Parent 1", "Towel", 1)
	insertItem(t, database, tripID, "Parent 1", "Book", 1)

	body := get(t, app, packURLFor(tripID))

	section := memberSection(t, body, "Parent 1")
	heading := memberHeading(t, section)
	for _, want := range []string{"Parent 1", "1 of 4 packed"} {
		if !strings.Contains(heading, want) {
			t.Errorf("Parent 1's heading does not read %q:\n%s", want, heading)
		}
	}
	if strings.Contains(heading, "4 items") {
		t.Errorf("Parent 1's heading still counts items instead of progress:\n%s", heading)
	}
	head := packFragment(t, section, packHeadID(parent))
	wantBar(t, "Parent 1's bar", head, 25, 25)
	if open := head[strings.Index(head, "data-bar"):]; strings.Contains(open, "</h2>") {
		t.Errorf("Parent 1's bar is inside the heading; it belongs under it:\n%s", head)
	}

	empty := memberSection(t, body, "Parent 2")
	if !strings.Contains(empty, "Nothing planned.") {
		t.Errorf("Parent 2, with nothing planned, no longer says so:\n%s", empty)
	}
	// Look for the count's wording, not bare words: "items" is also in a
	// Tailwind class name (items-baseline).
	if h := memberHeading(t, empty); strings.Contains(h, "of 0 packed") || strings.Contains(h, "0 items") {
		t.Errorf("Parent 2's heading has a count though nothing is planned:\n%s", h)
	}
	if strings.Contains(packFragment(t, empty, packHeadID(parent2)), "data-bar") {
		t.Errorf("Parent 2 has a bar though nothing is planned")
	}
}

// S5: A person's heading follows every tap.
func TestPersonHeadingFollowsEveryTap(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	parent := memberID(t, database, "Parent 1")
	parent2 := memberID(t, database, "Parent 2")
	socks := insertItem(t, database, tripID, "Parent 1", "Socks", 4)
	insertItem(t, database, tripID, "Parent 1", "Hat", 1)
	insertItem(t, database, tripID, "Parent 2", "Raincoat", 1)

	body := sendHTMX(t, app, prepareItemURL(tripID, parent, socks), url.Values{"packed_shown": {"0"}}).Body.String()
	head := packFragment(t, body, packHeadID(parent))
	wantOOB(t, "Parent 1's heading", head)
	if !strings.Contains(head, "0 of 2 packed") {
		t.Errorf("after preparing, Parent 1's heading does not read \"0 of 2 packed\":\n%s", head)
	}
	wantBar(t, "Parent 1's bar after preparing", head, 0, 50)

	body = sendHTMX(t, app, packItemURL(tripID, parent, socks), url.Values{"packed_shown": {"0"}}).Body.String()
	head = packFragment(t, body, packHeadID(parent))
	if !strings.Contains(head, "1 of 2 packed") {
		t.Errorf("after packing, Parent 1's heading does not read \"1 of 2 packed\":\n%s", head)
	}
	wantBar(t, "Parent 1's bar after packing", head, 50, 0)
	if strings.Contains(body, `id="`+packHeadID(parent2)+`"`) {
		t.Errorf("the reply also sends Parent 2's heading, which did not change")
	}
}

// S6: Every filter pill shows packed out of total.
func TestFilterPillsShowPackedOutOfTotal(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	packItem(t, database, insertItem(t, database, tripID, "Parent 1", "Socks", 4))
	insertItem(t, database, tripID, "Parent 1", "Hat", 1)
	insertItem(t, database, tripID, "Parent 1", "Towel", 1)
	insertItem(t, database, tripID, "Family", "Sunscreen", 1)

	body := get(t, app, packURLFor(tripID))

	if !strings.Contains(body, `id="`+packFilterID+`"`) {
		t.Fatalf("the filter row has no id %q to be swapped by", packFilterID)
	}
	nav := packFragment(t, body, packFilterID)
	for _, c := range []struct {
		name, href, short, long string
	}{
		{"Everyone", packURLFor(tripID), "1/4", "1 of 4 packed"},
		{"Family", narrowedPackURL(tripID, familyBucketID), "0/1", "0 of 1 packed"},
		{"Parent 1", narrowedPackURL(tripID, memberID(t, database, "Parent 1")), "1/3", "1 of 3 packed"},
		{"Parent 2", narrowedPackURL(tripID, memberID(t, database, "Parent 2")), "0/0", "0 of 0 packed"},
	} {
		pill := filterPill(t, nav, c.href)
		if !strings.Contains(pill, c.name) {
			t.Errorf("pill to %s does not name %q:\n%s", c.href, c.name, pill)
		}
		wantPillCount(t, c.name, pill, c.short, c.long)
	}

	// A person with nothing on the trip is dimmed, so the pills show who is
	// coming. Which style dims it is the template's call; the test checks
	// only that the empty pill is styled apart from a pill with items, and
	// that it is still a link.
	pillTag := func(href string) string { return tagWith(t, nav, `href="`+href+`"`, "<a") }
	classOf := func(tag string) string {
		return strings.TrimPrefix(between(t, tag, `class="`, `">`), `class="`)
	}
	parent2 := pillTag(narrowedPackURL(tripID, memberID(t, database, "Parent 2")))
	if classOf(parent2) == classOf(pillTag(narrowedPackURL(tripID, familyBucketID))) {
		t.Errorf("Parent 2's 0/0 pill looks like Family's 0/1 pill; want it dimmed: %s", parent2)
	}

	// The pill being viewed stays solid blue, even with nothing in it.
	narrowed := packFragment(t, get(t, app, narrowedPackURL(tripID, memberID(t, database, "Parent 2"))), packFilterID)
	active := tagWith(t, narrowed, `aria-current="page"`, "<a")
	if !strings.Contains(active, "?member="+fmt.Sprint(memberID(t, database, "Parent 2"))) {
		t.Fatalf("narrowed to Parent 2, but another pill is current: %s", active)
	}
	familyActive := tagWith(t, packFragment(t, get(t, app, narrowedPackURL(tripID, familyBucketID)), packFilterID), `aria-current="page"`, "<a")
	if classOf(active) != classOf(familyActive) {
		t.Errorf("the current pill is styled differently when it shows 0/0; want it solid blue like any current pill: %s", active)
	}
}

// S7: A finished pill turns green, and every pill follows each tap.
func TestFinishedPillAndPillsFollowEveryTap(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	insertItem(t, database, tripID, "Parent 1", "Socks", 4)
	sunscreen := insertItem(t, database, tripID, "Family", "Sunscreen", 1)
	prepareItem(t, database, sunscreen)

	body := sendHTMX(t, app, packItemURL(tripID, familyBucketID, sunscreen), url.Values{"packed_shown": {"0"}}).Body.String()

	nav := packFragment(t, body, packFilterID)
	wantOOB(t, "filter row", nav)
	family := filterPill(t, nav, narrowedPackURL(tripID, familyBucketID))
	wantPillCount(t, "Family", family, "1/1", "1 of 1 packed")
	if !strings.Contains(family, "<svg") {
		t.Errorf("finished Family pill has no tick:\n%s", family)
	}
	wantPillCount(t, "Everyone", filterPill(t, nav, packURLFor(tripID)), "1/2", "1 of 2 packed")
	if everyone := filterPill(t, nav, packURLFor(tripID)); strings.Contains(everyone, "<svg") {
		t.Errorf("unfinished Everyone pill has a tick:\n%s", everyone)
	}
	empty := filterPill(t, nav, narrowedPackURL(tripID, memberID(t, database, "Parent 2")))
	if strings.Contains(empty, "<svg") {
		t.Errorf("Parent 2's 0/0 pill is marked finished:\n%s", empty)
	}
	// The reply's pills keep the page's narrowing: Everyone is still the
	// one being viewed.
	if !strings.Contains(filterPill(t, nav, packURLFor(tripID)), `aria-current="page"`) {
		t.Errorf("after a tap on the page showing everyone, Everyone is no longer marked current:\n%s", nav)
	}
}

// S8: A past trip's packing page shows its progress.
func TestPastTripPackingPageShowsProgress(t *testing.T) {
	app, database := newTestApp(t)
	// The trip came home on 2026-10-07, a week before "today".
	app.now = fixedNow(t, "2026-10-15T09:00:00Z")
	tripID := insertTrip(t, database, "Parainen", "2026-10-05", 3)
	parent := memberID(t, database, "Parent 1")
	packItem(t, database, insertItem(t, database, tripID, "Parent 1", "Socks", 4))
	packItem(t, database, insertItem(t, database, tripID, "Parent 1", "Hat", 1))
	insertItem(t, database, tripID, "Parent 1", "Towel", 1)

	body := get(t, app, packURLFor(tripID))

	if header := pageHeader(t, body); !strings.Contains(header, "Packing · 2 of 3 packed") {
		t.Errorf("past trip's header does not read \"Packing · 2 of 3 packed\":\n%s", header)
	}
	if !strings.Contains(packFragment(t, body, tripBarID), `data-bar="packed"`) {
		t.Errorf("past trip has no trip bar")
	}
	if !strings.Contains(packFragment(t, body, packHeadID(parent)), `data-bar="packed"`) {
		t.Errorf("past trip has no bar under Parent 1's heading")
	}
	wantPillCount(t, "Parent 1", filterPill(t, body, narrowedPackURL(tripID, parent)), "2/3", "2 of 3 packed")
}

// S2, from the code review: a tap whose reply finds the trip empty (the other
// parent removed its last item on the trip page) clears the bar, instead of
// leaving the old green and amber on the page.
func TestTripBarClearsWhenTheTripEmpties(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	parent := memberID(t, database, "Parent 1")
	socks := insertItem(t, database, tripID, "Parent 1", "Socks", 4)
	packItem(t, database, socks)
	if _, err := database.Exec("DELETE FROM item WHERE id = ?", socks); err != nil {
		t.Fatalf("remove Socks: %v", err)
	}

	body := sendHTMX(t, app, unpackItemURL(tripID, parent, socks), url.Values{"packed_shown": {"1"}}).Body.String()

	bar := packFragment(t, body, tripBarID)
	wantOOB(t, "trip bar", bar)
	if strings.Contains(bar, "data-bar") {
		t.Errorf("the trip is empty but the reply's bar still has segments:\n%s", bar)
	}
}

// S7, from the code review: a tap on a narrowed page keeps the narrowed pill
// as the current one in the swapped filter row.
func TestNarrowedTapKeepsItsPillCurrent(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	insertItem(t, database, tripID, "Parent 1", "Socks", 4)
	sunscreen := insertItem(t, database, tripID, "Family", "Sunscreen", 1)
	prepareItem(t, database, sunscreen)

	tap := fmt.Sprintf("%s?member=%d", packItemURL(tripID, familyBucketID, sunscreen), familyBucketID)
	nav := packFragment(t, sendHTMX(t, app, tap, url.Values{"packed_shown": {"0"}}).Body.String(), packFilterID)

	if family := filterPill(t, nav, narrowedPackURL(tripID, familyBucketID)); !strings.Contains(family, `aria-current="page"`) {
		t.Errorf("after a tap narrowed to Family, its pill is no longer current:\n%s", nav)
	}
	if n := strings.Count(nav, `aria-current="page"`); n != 1 {
		t.Errorf("%d pills are marked current after the tap, want 1", n)
	}
}

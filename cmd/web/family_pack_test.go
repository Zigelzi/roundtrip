package main

import (
	"fmt"
	"strings"
	"testing"
)

// Milestone 03 (spec/milestones/03-family-items.md), packing page half:
// S2, S3, S4 and S6 there, plus S8's wording and S9's automated half.
//
// These were written before the implementation, from the scenarios. Some pass
// already: seeding the bucket as a family_member row gave the packing page its
// Family section, its progress and its filter for free, and those tests are
// here to keep it that way. The ones that fail are the milestone's remaining
// work: the filter's order (S2), and S8's wording throughout.

// filterLinkFor is the filter row's link to one member, or to everyone when
// id is 0. The closing quote matters: without it, member=1 also matches
// member=100.
func filterLinkFor(tripID, id int64) string {
	if id == 0 {
		return fmt.Sprintf(`href="%s"`, packURLFor(tripID))
	}
	return fmt.Sprintf(`href="%s?member=%d"`, packURLFor(tripID), id)
}

// S2: Family items show up on the packing page as their own section.
func TestPackPageShowsFamilySection(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	insertItem(t, database, tripID, familyMembers[0], "Toothbrush", 1)
	insertItem(t, database, tripID, "Family", "Sunscreen", 1)
	insertItem(t, database, tripID, "Family", "First aid kit", 1)

	body := get(t, app, packURLFor(tripID))

	section := memberSection(t, body, "Family")
	for _, name := range []string{"Sunscreen", "First aid kit"} {
		if !strings.Contains(section, name) {
			t.Errorf("Family section does not list %s:\n%s", name, section)
		}
	}
	if !strings.Contains(body, "0 of 3 packed") {
		t.Errorf("progress does not count family items alongside personal ones:\n%s", body)
	}
}

// S2: the Family section comes after all four people's sections. Specific
// things before general ones, and nothing else pins this ordering.
func TestFamilySectionComesLastOnPackPage(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	insertItem(t, database, tripID, "Family", "Sunscreen", 1)

	body := get(t, app, packURLFor(tripID))

	family := strings.Index(body, `data-member="Family"`)
	lastPerson := strings.Index(body, `data-member="`+familyMembers[3]+`"`)
	if family < 0 || lastPerson < 0 {
		t.Fatalf("could not find both %s's and Family's sections", familyMembers[3])
	}
	if family < lastPerson {
		t.Errorf("Family's section (offset %d) comes before %s's (offset %d)", family, familyMembers[3], lastPerson)
	}
}

// S2: in the filter row Family comes second, directly after "Everyone", even
// though its section is last. The filter is navigation, not the work order,
// and the bucket nobody owns should not be the one to hunt for.
func TestFamilyFilterComesSecond(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	insertItem(t, database, tripID, "Family", "Sunscreen", 1)

	body := get(t, app, packURLFor(tripID))

	everyone := strings.Index(body, filterLinkFor(tripID, 0))
	family := strings.Index(body, filterLinkFor(tripID, familyBucketID))
	if everyone < 0 || family < 0 {
		t.Fatalf("filter row is missing the Everyone or Family link")
	}
	if family < everyone {
		t.Errorf("Family (offset %d) comes before Everyone (offset %d)", family, everyone)
	}
	for _, person := range familyMembers {
		id := memberID(t, database, person)
		at := strings.Index(body, filterLinkFor(tripID, id))
		if at < 0 {
			t.Fatalf("filter row is missing %s", person)
		}
		if at < family {
			t.Errorf("%s (offset %d) comes before Family (offset %d); Family should be second", person, at, family)
		}
	}
}

// S3: Parent packs a family item. It behaves like any other item: out of the
// list to pack, into the packed section, progress up by one.
func TestPackFamilyItem(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	insertItem(t, database, tripID, familyMembers[0], "Toothbrush", 1)
	sunscreen := insertItem(t, database, tripID, "Family", "Sunscreen", 1)
	insertItem(t, database, tripID, "Family", "First aid kit", 1)

	body := sendHTMX(t, app, packItemURL(tripID, familyBucketID, sunscreen), nil).Body.String()

	toPack := packFragment(t, body, toPackListID(familyBucketID))
	if strings.Contains(toPack, "Sunscreen") {
		t.Errorf("packed family item is still in the list of things to pack:\n%s", toPack)
	}
	if !strings.Contains(toPack, "First aid kit") {
		t.Errorf("the family's other item vanished from the list:\n%s", toPack)
	}
	if packed := packFragment(t, body, "packed-list"); !strings.Contains(packed, "Sunscreen") {
		t.Errorf("packed family item is not in the packed section:\n%s", packed)
	}
	if progress := packFragment(t, body, "pack-progress"); !strings.Contains(progress, "1 of 3 packed") {
		t.Errorf("progress does not say 1 of 3 packed:\n%s", progress)
	}
	if got := itemStatus(t, database, sunscreen); got != "packed" {
		t.Errorf("stored status = %q, want %q", got, "packed")
	}
}

// S4: Parent packs only the family things.
func TestPackPageNarrowedToFamily(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	insertItem(t, database, tripID, familyMembers[0], "Toothbrush", 1)
	insertItem(t, database, tripID, "Family", "Sunscreen", 1)
	insertItem(t, database, tripID, "Family", "First aid kit", 1)

	body := get(t, app, narrowedPackURL(tripID, familyBucketID))

	if section := memberSection(t, body, "Family"); !strings.Contains(section, "Sunscreen") {
		t.Errorf("the Family section is missing its items:\n%s", section)
	}
	if strings.Contains(body, `data-member="`+familyMembers[0]+`"`) {
		t.Errorf("a person's section is still on the page")
	}
	if strings.Contains(body, "Toothbrush") {
		t.Errorf("a person's items are still on the page")
	}
	if !strings.Contains(body, "0 of 2 packed") {
		t.Errorf("progress counts more than the family's items:\n%s", body)
	}
}

// S6: the Family section is shown even when nothing in it is planned, the way
// an empty person's section is. Until milestone 04 that is most trips.
func TestFamilySectionShownWhenEmptyOnPackPage(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	insertItem(t, database, tripID, familyMembers[0], "Toothbrush", 1)

	body := get(t, app, packURLFor(tripID))

	if section := memberSection(t, body, "Family"); !strings.Contains(section, "Nothing planned") {
		t.Errorf("an empty Family section should say nothing is planned:\n%s", section)
	}
}

// S8: a parent packing one person is not told the trip is finished. The copy
// is pinned in the spec because this test asserts it literally.
func TestNarrowedPersonDoneButTripIsNot(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	first := memberID(t, database, familyMembers[0])
	packItem(t, database, insertItem(t, database, tripID, familyMembers[0], "Toothbrush", 1))
	packItem(t, database, insertItem(t, database, tripID, familyMembers[0], "Underwear", 4))
	insertItem(t, database, tripID, "Family", "Sunscreen", 1)

	body := get(t, app, narrowedPackURL(tripID, first))

	if want := familyMembers[0] + " is packed."; !strings.Contains(body, want) {
		t.Errorf("page does not say %q:\n%s", want, body)
	}
	if want := "2 of 3 packed on the trip."; !strings.Contains(body, want) {
		t.Errorf("page does not say %q, so the unpacked family items stay hidden:\n%s", want, body)
	}
	if strings.Contains(body, "Everything is packed") {
		t.Errorf("page claims everything is packed while the family's items are not")
	}
}

// S8: the same, narrowed to Family. "Family is packed" reads oddly for a
// bucket, so the wording differs.
func TestNarrowedFamilyDoneButTripIsNot(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	packItem(t, database, insertItem(t, database, tripID, "Family", "Sunscreen", 1))
	packItem(t, database, insertItem(t, database, tripID, "Family", "First aid kit", 1))
	insertItem(t, database, tripID, familyMembers[0], "Toothbrush", 1)

	body := get(t, app, narrowedPackURL(tripID, familyBucketID))

	if want := "The family's things are packed."; !strings.Contains(body, want) {
		t.Errorf("page does not say %q:\n%s", want, body)
	}
	if want := "2 of 3 packed on the trip."; !strings.Contains(body, want) {
		t.Errorf("page does not say %q:\n%s", want, body)
	}
	if strings.Contains(body, "Everything is packed") {
		t.Errorf("page claims everything is packed while a person's items are not")
	}
}

// S8: the whole-family view is unchanged. With items left it says nothing,
// and when the trip really is done it still says what milestone 02 said.
func TestWholeFamilyViewKeepsItsWording(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	toothbrush := insertItem(t, database, tripID, familyMembers[0], "Toothbrush", 1)
	sunscreen := insertItem(t, database, tripID, "Family", "Sunscreen", 1)

	body := get(t, app, packURLFor(tripID))
	for _, unwanted := range []string{" is packed.", "The family's things are packed.", "Everything is packed", "on the trip."} {
		if strings.Contains(body, unwanted) {
			t.Errorf("whole-family view with items left says %q:\n%s", unwanted, body)
		}
	}

	packItem(t, database, toothbrush)
	packItem(t, database, sunscreen)

	if done := get(t, app, packURLFor(tripID)); !strings.Contains(done, "Everything is packed") {
		t.Errorf("a fully packed trip no longer says everything is packed:\n%s", done)
	}
}

// S8: once the whole trip is packed, a narrowed view says so too. The person
// is done and so is the trip, so the honest message is the global one.
// The spec pinned only the "person done, trip not" case; this is the gap.
func TestNarrowedViewOnAFullyPackedTrip(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	first := memberID(t, database, familyMembers[0])
	packItem(t, database, insertItem(t, database, tripID, familyMembers[0], "Toothbrush", 1))
	packItem(t, database, insertItem(t, database, tripID, "Family", "Sunscreen", 1))

	body := get(t, app, narrowedPackURL(tripID, first))

	if !strings.Contains(body, "Everything is packed") {
		t.Errorf("the trip is fully packed but the narrowed view does not say so:\n%s", body)
	}
}

// S9: the filter row offers every bucket and marks exactly one as current.
// The layout itself (that six links stay tappable at phone width) is a manual
// check; this is the part a Go test can see. Milestone 02's S1 cannot cover
// it: that test loops over the four people and never sees six links.
func TestPackFilterOffersEveryBucketWithOneCurrent(t *testing.T) {
	app, database := newTestApp(t)
	tripID := insertTrip(t, database, "Parainen", "2026-09-20", 3)
	insertItem(t, database, tripID, familyMembers[0], "Toothbrush", 1)
	insertItem(t, database, tripID, "Family", "Sunscreen", 1)

	for _, narrowed := range []int64{0, familyBucketID, memberID(t, database, familyMembers[0])} {
		url := packURLFor(tripID)
		if narrowed != 0 {
			url = narrowedPackURL(tripID, narrowed)
		}
		body := get(t, app, url)

		links := []int64{0, familyBucketID}
		for _, person := range familyMembers {
			links = append(links, memberID(t, database, person))
		}
		for _, id := range links {
			if !strings.Contains(body, filterLinkFor(tripID, id)) {
				t.Errorf("%s: filter row is missing the link for %d", url, id)
			}
		}
		if n := strings.Count(body, `aria-current="page"`); n != 1 {
			t.Errorf("%s: %d links marked current, want exactly 1", url, n)
		}
	}
}

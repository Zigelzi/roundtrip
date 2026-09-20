package view

import (
	"fmt"

	"github.com/a-h/templ"
)

// PackPageData is everything the packing page shows. It is a separate view
// from the trip page on purpose: planning and packing are different
// activities (see spec/milestones/02-pack-items.md).
type PackPageData struct {
	Trip    Trip
	Members []PackMember
	// Packed and Total drive the progress line. They count what is shown:
	// narrowed to one member, they are that member's numbers.
	Packed int
	Total  int
	// TripTotal counts the whole trip regardless of narrowing, so that
	// narrowing to a member with nothing planned does not look like an
	// empty trip.
	TripTotal int
	// Filter is the family member the page is narrowed to, 0 for everyone.
	Filter int64
}

// Shown is the members whose items the page is displaying: all of them, or
// just the one the page is narrowed to.
func (p PackPageData) Shown() []PackMember {
	if p.Filter == 0 {
		return p.Members
	}
	for _, m := range p.Members {
		if m.ID == p.Filter {
			return []PackMember{m}
		}
	}
	return nil
}

// PackMember is one family member's section on the packing page: what is
// still to pack, and what is already in the bag.
type PackMember struct {
	ID     int64
	Name   string
	ToPack []Item
	Packed []Item
}

// HasItems says whether the member is on the trip at all. A member with
// nothing planned and a member who is finished both show an empty to-pack
// list, and they must not read the same (S9).
func (m PackMember) HasItems() bool {
	return len(m.ToPack)+len(m.Packed) > 0
}

// PackURL is a trip's packing page.
func PackURL(tripID int64) string {
	return fmt.Sprintf("/trips/%d/pack", tripID)
}

// packMemberURL narrows the packing page to one family member.
func packMemberURL(tripID, memberID int64) string {
	return fmt.Sprintf("%s?member=%d", PackURL(tripID), memberID)
}

// packItemURL and unpackItemURL carry the narrowing, so a tap comes back
// with the same slice of the page the parent is looking at.
func packItemURL(tripID, memberID, itemID, filter int64) string {
	return WithFilter(fmt.Sprintf("/trips/%d/members/%d/items/%d/pack", tripID, memberID, itemID), filter)
}

func unpackItemURL(tripID, memberID, itemID, filter int64) string {
	return WithFilter(fmt.Sprintf("/trips/%d/members/%d/items/%d/unpack", tripID, memberID, itemID), filter)
}

// WithFilter keeps the page narrowed to one member across a request.
func WithFilter(url string, filter int64) string {
	if filter == 0 {
		return url
	}
	return fmt.Sprintf("%s?member=%d", url, filter)
}

// toPackListID is the id of one member's swappable list of things to pack.
func toPackListID(memberID int64) string {
	return fmt.Sprintf("topack-member-%d", memberID)
}

// oobAttrs marks a fragment as an out-of-band swap. One tap changes three
// places (the member's list, the packed section and the progress), and
// htmx swaps the extra two by id wherever they sit on the page.
func oobAttrs(oob bool) templ.Attributes {
	if oob {
		return templ.Attributes{"hx-swap-oob": "true"}
	}
	return templ.Attributes{}
}

// FormatCount renders a plain number for templates, which interpolate
// strings rather than ints.
func FormatCount(n int) string {
	return fmt.Sprintf("%d", n)
}

// FormatProgress shows how much of the list is in the bag: "7 of 12 packed".
func FormatProgress(packed, total int) string {
	return fmt.Sprintf("%d of %d packed", packed, total)
}

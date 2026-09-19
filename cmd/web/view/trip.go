package view

import (
	"fmt"
	"time"
)

// Trip is a trip as pages show it.
type Trip struct {
	ID           int64
	Destination  string
	Departure    time.Time
	DurationDays int64
}

// TripForm holds the create-trip form's values as typed, so they can be
// shown again next to validation errors.
type TripForm struct {
	Destination   string
	DepartureDate string
	EndDate       string
	DurationDays  string
	Errors        []string
}

// FormatDate shows a date the Finnish way: 20.9.2026.
func FormatDate(t time.Time) string {
	return t.Format("2.1.2006")
}

// FormatDateRange shows a trip from departure to return day, dropping the
// month and year from the start date when they repeat: 20.–22.9.2026,
// 30.11.–3.12.2026, 30.12.2026–3.1.2027. A one-day trip is just its date.
func FormatDateRange(departure time.Time, days int64) string {
	end := departure.AddDate(0, 0, int(days)-1)
	switch {
	case days <= 1:
		return FormatDate(departure)
	case departure.Year() != end.Year():
		return FormatDate(departure) + "–" + FormatDate(end)
	case departure.Month() != end.Month():
		return departure.Format("2.1.") + "–" + FormatDate(end)
	default:
		return departure.Format("2.") + "–" + FormatDate(end)
	}
}

// FormatQuantity shows how many to pack: "1 pc", "4 pcs". Written out
// rather than "× 4" so it can't be mistaken for the ✕ remove button.
func FormatQuantity(n int64) string {
	if n == 1 {
		return "1 pc"
	}
	return fmt.Sprintf("%d pcs", n)
}

// FormatDays shows a duration: "1 day", "3 days".
func FormatDays(n int64) string {
	if n == 1 {
		return "1 day"
	}
	return fmt.Sprintf("%d days", n)
}

// TripURL is a trip's page.
func TripURL(id int64) string {
	return fmt.Sprintf("/trips/%d", id)
}

// TripPageData is everything the trip page shows.
type TripPageData struct {
	Trip    Trip
	Members []MemberItems
	// Suggestions are item names used on any trip, offered while typing.
	Suggestions []string
}

// MemberItems is one family member's section: their items and add form.
type MemberItems struct {
	ID    int64
	Name  string
	Items []Item
	Form  ItemForm
}

type Item struct {
	ID       int64
	Name     string
	Quantity int64
}

// ItemForm holds a member's add-item form values as typed.
type ItemForm struct {
	Name     string
	Quantity string
	Errors   []string
}

// memberAnchor is the id of a member's section on the trip page.
func memberAnchor(id int64) string {
	return fmt.Sprintf("member-%d", id)
}

// itemListID is the id of a member's swappable item list.
func itemListID(memberID int64) string {
	return fmt.Sprintf("items-member-%d", memberID)
}

// MemberURL points at a member's section on the trip page, so the browser
// scrolls back to where the parent was working.
func MemberURL(tripID, memberID int64) string {
	return TripURL(tripID) + "#" + memberAnchor(memberID)
}

// addItemURL keeps the member's anchor so that if the input is invalid, the
// re-rendered page scrolls to the form with the error.
func addItemURL(tripID, memberID int64) string {
	return fmt.Sprintf("/trips/%d/members/%d/items#%s", tripID, memberID, memberAnchor(memberID))
}

func removeItemURL(tripID, itemID int64) string {
	return fmt.Sprintf("/trips/%d/items/%d/delete", tripID, itemID)
}

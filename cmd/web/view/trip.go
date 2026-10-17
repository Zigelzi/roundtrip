package view

import (
	"fmt"
	"strconv"
	"time"
)

// Trip is a trip as pages show it.
type Trip struct {
	ID           int64
	Destination  string
	Departure    time.Time
	DurationDays int64
	// Packed, Prepared and Total count the trip's items for the trips list's
	// progress (14/S7). Only the list fills them in, and only for upcoming
	// trips, so they are zero on every other page.
	Packed   int
	Prepared int
	Total    int
}

// ReturnDay is the trip's last day: the departure day counts, so a 3-day
// trip from 20.9 returns on 22.9.
func (t Trip) ReturnDay() time.Time {
	return t.Departure.AddDate(0, 0, int(t.DurationDays)-1)
}

// MaxTripNights caps a trip's length at 14 days, so a typo can't create a
// trip with hundreds of basics that can't be deleted (04/S11). It is the end
// of the new trip form's slider, which starts at 1: a trip is at least 2
// days (13/R2).
const MaxTripNights = 13

// defaultNights is where the slider starts: 3 days, as the form started
// before nights replaced days (13/S1).
const defaultNights = 2

// DateLayout is how dates travel in forms (<input type="date">) and the
// database.
const DateLayout = "2006-01-02"

// returnHint holds the return date's place on the new trip form until a
// departure is chosen, so the line does not jump in later (13/Q2).
const returnHint = "Pick a departure date to see the return date."

// TripForm holds the create-trip form's values as typed, so they can be
// shown again next to validation errors.
type TripForm struct {
	Destination   string
	DepartureDate string
	Nights        string
	People        []TripPerson
	Errors        []string
}

// SliderNights is where the slider sits: the nights sent, kept within 1 to
// MaxTripNights, or defaultNights when there are none (a new form) or they
// are not a number. Only a stale or forged form sends a value outside the
// slider; drawn as sent, the browser would move the slider into range and
// leave the label and return date disagreeing with it.
func (f TripForm) SliderNights() int {
	n, err := strconv.Atoi(f.Nights)
	if err != nil {
		return defaultNights
	}
	return min(max(n, 1), MaxTripNights)
}

// NightsLabel is the slider's value in words, "7 nights", shown beside it and
// read out by screen readers (13/S6).
func (f TripForm) NightsLabel() string {
	if n := f.SliderNights(); n != 1 {
		return fmt.Sprintf("%d nights", n)
	}
	return "1 night"
}

// ReturnDate is the day the trip comes back for the slider's nights, with its
// weekday so it can be checked against the booking (13/Q1): "Sun 27.12.2026".
// It is "" until a departure is known. A trip of n nights is n + 1 days, so
// this is the last day of FormatDateRange.
func (f TripForm) ReturnDate() string {
	departure, err := time.Parse(DateLayout, f.DepartureDate)
	if err != nil {
		return ""
	}
	end := departure.AddDate(0, 0, f.SliderNights())
	return end.Format("Mon ") + FormatDate(end)
}

// TripPerson is one tick box on the create-trip form: a person, and whether
// they are going.
type TripPerson struct {
	ID     int64
	Name   string
	Ticked bool
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

// FormatItemCount shows how many items a person has: "1 item", "18 items".
func FormatItemCount(n int) string {
	if n == 1 {
		return "1 item"
	}
	return fmt.Sprintf("%d items", n)
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
	Edit        ItemEdit
	// Packed and Total count the whole trip's items, for the packing bar
	// (12/S1). Prepared items are not packed (12/Q2).
	Packed int
	Total  int
}

// ItemEdit is the trip page's edit mode (?edit=ITEM_ID): the one item whose
// name and quantity are being changed. While ItemID is set, every other row's
// Edit and Remove buttons are disabled (04/S10). Zero means no row is being
// edited.
type ItemEdit struct {
	ItemID int64
	// Name and Quantity are the values in the edit fields: the item's own, or
	// what was submitted if Submitted is set (a refused save), even when that
	// is empty (15/S5).
	Name      string
	Quantity  string
	Submitted bool
	Errors    []string
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
	// Prepared is set on the packing page for an item gathered but not yet
	// in the bag. The trip page leaves it false: planning does not show it.
	Prepared bool
}

// ItemForm holds a member's add-item form values as typed.
type ItemForm struct {
	Name     string
	Quantity string
	Errors   []string
}

// MemberLink is what a jump link needs: the member's section and its label.
type MemberLink struct {
	ID   int64
	Name string
}

func tripLinks(members []MemberItems) []MemberLink {
	links := make([]MemberLink, len(members))
	for i, m := range members {
		links[i] = MemberLink{ID: m.ID, Name: m.Name}
	}
	return links
}

// memberAnchor is the id of a member's section on the trip and basics pages.
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
// re-rendered page scrolls to the form with the error. While a row is being
// edited it also carries ?edit=, so adding an item keeps that row open and
// the others locked.
func addItemURL(tripID, memberID, editID int64) string {
	edit := ""
	if editID != 0 {
		edit = fmt.Sprintf("?edit=%d", editID)
	}
	return fmt.Sprintf("/trips/%d/members/%d/items%s#%s", tripID, memberID, edit, memberAnchor(memberID))
}

// itemAnchor is the id of an item's row on the trip page.
func itemAnchor(id int64) string {
	return fmt.Sprintf("item-%d", id)
}

// ItemURL points at an item's row on the trip page, so on a long list the
// browser lands where the parent was working.
func ItemURL(tripID, itemID int64) string {
	return TripURL(tripID) + "#" + itemAnchor(itemID)
}

func saveItemURL(tripID, itemID int64) string {
	return fmt.Sprintf("/trips/%d/items/%d", tripID, itemID)
}

func removeItemURL(tripID, itemID int64) string {
	return fmt.Sprintf("/trips/%d/items/%d/delete", tripID, itemID)
}

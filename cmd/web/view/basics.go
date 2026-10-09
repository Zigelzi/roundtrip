package view

import (
	"fmt"
	"strconv"
)

// BasicsURL is the page listing every family member's basics.
const BasicsURL = "/basics"

// MemberBasics is one family member's section on the basics page.
type MemberBasics struct {
	ID     int64
	Name   string
	Basics []Basic
	// Form is the add form's state: empty normally, or what was typed and
	// the messages when an add was refused (only Name, PerDay, Fixed and
	// Errors are used).
	Form BasicEdit
}

// Basic is one of the items every new trip starts with. Its quantity on a
// trip is PerDay x trip days + Fixed; a basic with nothing a day is a
// fixed-quantity basic (spec/milestones/06-edit-basics.md, Q8).
type Basic struct {
	ID     int64
	Name   string
	PerDay int64
	Fixed  int64
}

// FormatBasicQuantity shows a basic's quantity as two kinds: a fixed basic
// is just its number ("2"), a daily one "1 / day", plus "+ 1 extra" when it
// has extra.
func FormatBasicQuantity(perDay, fixed int64) string {
	switch {
	case perDay == 0:
		return fmt.Sprint(fixed)
	case fixed == 0:
		return fmt.Sprintf("%d / day", perDay)
	default:
		return fmt.Sprintf("%d / day + %d extra", perDay, fixed)
	}
}

// MaxBasicQuantity is the most either quantity field takes (R7). No family
// packs more than 20 of one thing a day or as a fixed number, and a cap keeps
// a typo from breaking trip creation (per day x days + fixed must not
// overflow). Loosen it if a real case turns up. Only the handler checks it:
// the fields have no browser min or max, so the page shows the message.
const MaxBasicQuantity = 20

// BasicEdit is the basics page's edit mode (?edit=BASIC_ID): the one basic
// being changed. While ID is set, every other row's
// Change button is disabled. Zero means no row is being edited.
type BasicEdit struct {
	ID int64
	// Name, PerDay and Fixed are the values in the fields: the basic's own, or
	// what was submitted if it was refused. A zero shows as an empty field,
	// so an empty "/ day" is what marks a fixed basic (Q8).
	Name   string
	PerDay string
	Fixed  string
	// Submitted means the fields hold what the parent typed, even if every
	// one is empty, so they must not be refilled from the stored basic.
	Submitted bool
	Errors    []string
}

// FieldValue is a quantity as shown in an edit field: zero is empty.
func FieldValue(n int64) string {
	if n == 0 {
		return ""
	}
	return strconv.FormatInt(n, 10)
}

// basicAnchor is the id of a basic's row on the basics page.
func basicAnchor(id int64) string {
	return fmt.Sprintf("basic-%d", id)
}

// BasicURL points at a basic's row, so the browser lands where the parent
// was working.
func BasicURL(id int64) string {
	return BasicsURL + "#" + basicAnchor(id)
}

// BasicEditURL opens a basic's row for editing.
func BasicEditURL(id int64) string {
	return fmt.Sprintf("%s?edit=%d#%s", BasicsURL, id, basicAnchor(id))
}

func basicSaveURL(id int64) string {
	return fmt.Sprintf("%s/%d", BasicsURL, id)
}

// MemberBasicsURL points at a member's section, so the browser lands where
// the parent was working.
func MemberBasicsURL(memberID int64) string {
	return BasicsURL + "#" + memberAnchor(memberID)
}

func addBasicURL(memberID int64) string {
	return fmt.Sprintf("/members/%d/basics", memberID)
}

func removeBasicURL(id int64) string {
	return fmt.Sprintf("%s/%d/delete", BasicsURL, id)
}

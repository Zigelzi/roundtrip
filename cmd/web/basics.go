package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Zigelzi/roundtrip/cmd/web/view"
	"github.com/Zigelzi/roundtrip/internal/db"
)

// handleBasics renders every family member's basics (milestone 06), in the
// order a new trip gets them: members by id, so the Family bucket (100)
// comes last, and each member's basics by position. ?edit=ID opens one row
// for editing.
func (app *application) handleBasics(w http.ResponseWriter, r *http.Request) {
	editID, _ := strconv.ParseInt(r.URL.Query().Get("edit"), 10, 64)
	sections, edit, err := app.basicsPage(r.Context(), view.BasicEdit{ID: editID})
	if err != nil {
		app.serverError(w, "load basics", err)
		return
	}
	app.render(w, r, http.StatusOK, view.BasicsPage(sections, edit))
}

// basicsPage loads the sections. edit is the row to show in edit mode; it is
// dropped if the basic no longer exists, and its fields start at the basic's
// own values unless edit already carries what was submitted.
func (app *application) basicsPage(ctx context.Context, edit view.BasicEdit) ([]view.MemberBasics, view.BasicEdit, error) {
	members, err := app.queries.ListFamilyMembers(ctx)
	if err != nil {
		return nil, edit, err
	}
	basics, err := app.queries.ListBasicItems(ctx)
	if err != nil {
		return nil, edit, err
	}
	found := view.BasicEdit{}
	sections := make([]view.MemberBasics, len(members))
	for i, m := range members {
		sections[i] = view.MemberBasics{ID: m.ID, Name: m.Name}
		for _, b := range basics {
			if b.FamilyMemberID != m.ID {
				continue
			}
			sections[i].Basics = append(sections[i].Basics, view.Basic{ID: b.ID, Name: b.Name, PerDay: b.PerDay, Fixed: b.Fixed})
			if b.ID == edit.ID {
				found = edit
				if !found.Submitted {
					found.Name, found.PerDay, found.Fixed = b.Name, view.FieldValue(b.PerDay), view.FieldValue(b.Fixed)
				}
			}
		}
	}
	return sections, found, nil
}

const (
	badBasicNumber = "Up to 20 per field."
	emptyBasic     = "Enter how many a day, how many extra, or both."
	noBasicName    = "Enter a name for the basic."
)

// handleSaveBasic saves a basic's name and its daily and extra quantity (S2,
// S3, S8, S11). An empty quantity field counts as 0. A refused value
// re-renders the page with the row still open and the message (S7, S9, S12).
// A basic that is already gone is not an error (R5): the page is shown as it
// is.
func (app *application) handleSaveBasic(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Redirect(w, r, view.BasicsURL, http.StatusSeeOther)
		return
	}
	form := view.BasicEdit{
		ID:        id,
		Submitted: true,
		Name:      strings.TrimSpace(r.PostFormValue("name")),
		PerDay:    strings.TrimSpace(r.PostFormValue("per_day")),
		Fixed:     strings.TrimSpace(r.PostFormValue("fixed")),
	}
	sections, form, err := app.basicsPage(r.Context(), form)
	if err != nil {
		app.serverError(w, "load basics", err)
		return
	}
	// The page drops an edit for a basic that is gone (R5).
	if form.ID == 0 {
		http.Redirect(w, r, view.BasicsURL, http.StatusSeeOther)
		return
	}
	perDay, fixed, errs := validateBasic(form, siblingsOf(sections, id))
	if len(errs) > 0 {
		app.refuseBasic(w, r, sections, form, errs)
		return
	}

	rows, err := app.queries.UpdateBasicItem(r.Context(), db.UpdateBasicItemParams{
		Name:    form.Name,
		NameKey: itemKey(form.Name),
		PerDay:  perDay,
		Fixed:   fixed,
		ID:      id,
	})
	// A unique violation is the other parent saving the same name at the same
	// moment: refuse it like any duplicate.
	if db.IsUniqueViolation(err) {
		app.refuseBasic(w, r, sections, form, []string{alreadyHasBasic(siblingsOf(sections, id), form.Name)})
		return
	}
	if err != nil {
		app.serverError(w, "update basic", err)
		return
	}
	if rows == 0 {
		http.Redirect(w, r, view.BasicsURL, http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, view.BasicURL(id), http.StatusSeeOther)
}

func (app *application) refuseBasic(w http.ResponseWriter, r *http.Request, sections []view.MemberBasics, form view.BasicEdit, errs []string) {
	form.Errors = errs
	app.render(w, r, http.StatusUnprocessableEntity, view.BasicsPage(sections, form))
}

// siblings are the basics of the family member who owns basic id, other than
// that basic itself, plus the member's name.
type siblings struct {
	member string
	others []view.Basic
}

func siblingsOf(sections []view.MemberBasics, id int64) siblings {
	for _, m := range sections {
		for _, b := range m.Basics {
			if b.ID == id {
				s := siblings{member: m.Name}
				for _, o := range m.Basics {
					if o.ID != id {
						s.others = append(s.others, o)
					}
				}
				return s
			}
		}
	}
	return siblings{}
}

// alreadyHasBasic is the S10 and S12 message, naming the existing basic as
// it is spelled on the page.
func alreadyHasBasic(s siblings, name string) string {
	for _, o := range s.others {
		if itemKey(o.Name) == itemKey(name) {
			return fmt.Sprintf("%s already has %s.", s.member, o.Name)
		}
	}
	return fmt.Sprintf("%s already has %s.", s.member, name)
}

// validateBasic checks the edit form. It returns the quantities to store, or
// one user-facing message per problem.
func validateBasic(form view.BasicEdit, s siblings) (perDay, fixed int64, errs []string) {
	if form.Name == "" {
		errs = append(errs, noBasicName)
	} else {
		for _, o := range s.others {
			if itemKey(o.Name) == itemKey(form.Name) {
				errs = append(errs, alreadyHasBasic(s, form.Name))
				break
			}
		}
	}
	perDay, okDay := parseBasicNumber(form.PerDay)
	fixed, okFixed := parseBasicNumber(form.Fixed)
	switch {
	case !okDay || !okFixed:
		errs = append(errs, badBasicNumber)
	case perDay+fixed < 1:
		errs = append(errs, emptyBasic)
	}
	return perDay, fixed, errs
}

// handleAddBasic adds a basic last in a family member's list (S4). It takes
// the same three fields as the edit form (R2) and follows the same rules (S7,
// S9, S10). A refused add re-renders the page with the message under that
// member's form.
func (app *application) handleAddBasic(w http.ResponseWriter, r *http.Request) {
	memberID, err := strconv.ParseInt(r.PathValue("memberID"), 10, 64)
	if err != nil {
		app.render(w, r, http.StatusNotFound, view.NotFound("Family member not found"))
		return
	}
	sections, _, err := app.basicsPage(r.Context(), view.BasicEdit{})
	if err != nil {
		app.serverError(w, "load basics", err)
		return
	}
	i := -1
	for n, m := range sections {
		if m.ID == memberID {
			i = n
		}
	}
	if i < 0 {
		app.render(w, r, http.StatusNotFound, view.NotFound("Family member not found"))
		return
	}

	form := view.BasicEdit{
		Submitted: true,
		Name:      strings.TrimSpace(r.PostFormValue("name")),
		PerDay:    strings.TrimSpace(r.PostFormValue("per_day")),
		Fixed:     strings.TrimSpace(r.PostFormValue("fixed")),
	}
	perDay, fixed, errs := validateBasic(form, siblings{member: sections[i].Name, others: sections[i].Basics})
	if len(errs) > 0 {
		form.Errors = errs
		sections[i].Form = form
		app.render(w, r, http.StatusUnprocessableEntity, view.BasicsPage(sections, view.BasicEdit{}))
		return
	}
	err = app.queries.CreateBasicItem(r.Context(), db.CreateBasicItemParams{
		FamilyMemberID: memberID,
		Name:           form.Name,
		NameKey:        itemKey(form.Name),
		PerDay:         perDay,
		Fixed:          fixed,
	})
	// A unique violation is a double-tapped Add: the first request already
	// saved the basic, so treat it as success.
	if err != nil && !db.IsUniqueViolation(err) {
		app.serverError(w, "create basic", err)
		return
	}
	http.Redirect(w, r, view.MemberBasicsURL(memberID), http.StatusSeeOther)
}

// handleRemoveBasic removes a basic, with no confirmation (Q3). A basic that
// is already gone (a double tap, or the other parent) leaves the page as it
// is (R5).
func (app *application) handleRemoveBasic(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Redirect(w, r, view.BasicsURL, http.StatusSeeOther)
		return
	}
	memberID, err := app.queries.DeleteBasicItem(r.Context(), id)
	if errors.Is(err, sql.ErrNoRows) {
		http.Redirect(w, r, view.BasicsURL, http.StatusSeeOther)
		return
	}
	if err != nil {
		app.serverError(w, "delete basic", err)
		return
	}
	http.Redirect(w, r, view.MemberBasicsURL(memberID), http.StatusSeeOther)
}

// parseBasicNumber reads a typed quantity: empty is 0, otherwise a whole
// number from 0 to view.MaxBasicQuantity.
func parseBasicNumber(s string) (n int64, ok bool) {
	if s == "" {
		return 0, true
	}
	n, err := strconv.ParseInt(s, 10, 64)
	return n, err == nil && n >= 0 && n <= view.MaxBasicQuantity
}

package main

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
)

// Milestone 06 (spec/milestones/06-edit-basics.md), edit basics.

const basicsURL = "/basics"

// basicQuantityText is how the basics page shows a basic's quantity, written
// out here from the spec's Test notes rather than by calling the view's
// formatter, so the test states the contract instead of copying the code: a
// fixed basic is just its number, a daily one "N / day", plus "+ M extra"
// when it has extra.
func basicQuantityText(b basicRow) string {
	switch {
	case b.perDay == 0:
		return strconv.Itoa(b.fixed)
	case b.fixed == 0:
		return strconv.Itoa(b.perDay) + " / day"
	default:
		return strconv.Itoa(b.perDay) + " / day + " + strconv.Itoa(b.fixed) + " extra"
	}
}

// memberNameByID is the display name of a family member row, which
// FAMILY_NAMES may have renamed, so tests look it up rather than assume it.
func memberNameByID(t *testing.T, database *sql.DB, id int64) string {
	t.Helper()
	var name string
	if err := database.QueryRow("SELECT name FROM family_member WHERE id = ?", id).Scan(&name); err != nil {
		t.Fatalf("family member %d: %v", id, err)
	}
	return name
}

// S1: Parent sees everyone's basics
func TestBasicsPageListsEveryonesBasics(t *testing.T) {
	app, database := newTestApp(t)

	if !strings.Contains(get(t, app, "/"), `href="`+basicsURL+`"`) {
		t.Fatalf("start page has no link to %s", basicsURL)
	}
	body := get(t, app, basicsURL)

	// Sections appear in the order a new trip gets them: Parent 1, Parent 2,
	// Child 1, Child 2, then the Family bucket.
	order := []int64{1, 2, 3, 4, familyBucketID}
	last := -1
	for _, id := range order {
		name := memberNameByID(t, database, id)
		at := strings.Index(body, `data-member="`+name+`"`)
		if at < 0 {
			t.Fatalf("no section for %s", name)
		}
		if at < last {
			t.Errorf("section for %s is out of order", name)
		}
		last = at
	}

	for _, id := range order {
		name := memberNameByID(t, database, id)
		section := memberSection(t, body, name)
		pos := -1
		for _, b := range catalogueBasics {
			if b.member != id {
				continue
			}
			row := basicRowHTML(t, section, b.name)
			if want := ">" + basicQuantityText(b) + "<"; !strings.Contains(row, want) {
				t.Errorf("%s's %q: quantity shown is not %q in:\n%s", name, b.name, basicQuantityText(b), row)
			}
			at := strings.Index(section, ">"+b.name+"<")
			if at < pos {
				t.Errorf("%s's %q is out of catalogue order", name, b.name)
			}
			pos = at
		}
	}
}

// basicRowHTML returns the <li> on the basics page whose name is name.
func basicRowHTML(t *testing.T, section, name string) string {
	t.Helper()
	i := strings.Index(section, ">"+name+"<")
	if i < 0 {
		t.Fatalf("no basic %q in section", name)
	}
	start := strings.LastIndex(section[:i], "<li")
	end := strings.Index(section[i:], "</li>")
	if start < 0 || end < 0 {
		t.Fatalf("basic %q is not in a list row", name)
	}
	return section[start : i+end]
}

func basicIDOf(t *testing.T, database *sql.DB, member int64, name string) int64 {
	t.Helper()
	var id int64
	if err := database.QueryRow("SELECT id FROM basic_item WHERE family_member_id = ? AND name = ?", member, name).Scan(&id); err != nil {
		t.Fatalf("member %d has no basic %q: %v", member, name, err)
	}
	return id
}

func basicStored(t *testing.T, database *sql.DB, id int64) (perDay, fixed int) {
	t.Helper()
	if err := database.QueryRow("SELECT per_day, fixed FROM basic_item WHERE id = ?", id).Scan(&perDay, &fixed); err != nil {
		t.Fatalf("read basic %d: %v", id, err)
	}
	return perDay, fixed
}

func basicSaveURL(id int64) string { return basicsURL + "/" + strconv.FormatInt(id, 10) }

// saveBasic posts the edit form: the name and both quantities, as the form
// always sends all three.
func saveBasic(t *testing.T, app *application, id int64, name, perDay, fixed string) *httptest.ResponseRecorder {
	t.Helper()
	return send(t, app, http.MethodPost, basicSaveURL(id), url.Values{"name": {name}, "per_day": {perDay}, "fixed": {fixed}})
}

func basicNameStored(t *testing.T, database *sql.DB, id int64) string {
	t.Helper()
	var name string
	if err := database.QueryRow("SELECT name FROM basic_item WHERE id = ?", id).Scan(&name); err != nil {
		t.Fatalf("read basic %d: %v", id, err)
	}
	return name
}

// assertBasicShown checks the quantity the basics page shows for a basic,
// read from the page rather than the database.
func assertBasicShown(t *testing.T, app *application, database *sql.DB, member int64, name string, want basicRow) {
	t.Helper()
	section := memberSection(t, get(t, app, basicsURL), memberNameByID(t, database, member))
	row := basicRowHTML(t, section, name)
	if text := ">" + basicQuantityText(want) + "<"; !strings.Contains(row, text) {
		t.Errorf("%q should show %q, row is:\n%s", name, basicQuantityText(want), row)
	}
}

// S3: Parent changes the daily quantity
func TestChangeBasicDailyQuantity(t *testing.T) {
	app, database := newTestApp(t)
	id := basicIDOf(t, database, 1, "Underwear")

	if rec := saveBasic(t, app, id, "Underwear", "2", "1"); rec.Code != http.StatusSeeOther {
		t.Fatalf("save: status = %d, want 303", rec.Code)
	}

	assertBasicShown(t, app, database, 1, "Underwear", basicRow{perDay: 2, fixed: 1})
	if got := quantityOf(t, database, createTripOf(t, app, database, 3), 1, "Underwear"); got != 7 {
		t.Errorf("new 3-day trip has %d underwear, want 7", got)
	}
}

// S8: Parent changes the extra quantity
func TestChangeBasicExtraQuantity(t *testing.T) {
	app, database := newTestApp(t)
	id := basicIDOf(t, database, 1, "Underwear")

	if rec := saveBasic(t, app, id, "Underwear", "1", "2"); rec.Code != http.StatusSeeOther {
		t.Fatalf("save: status = %d, want 303", rec.Code)
	}

	assertBasicShown(t, app, database, 1, "Underwear", basicRow{perDay: 1, fixed: 2})
	if got := quantityOf(t, database, createTripOf(t, app, database, 3), 1, "Underwear"); got != 5 {
		t.Errorf("new 3-day trip has %d underwear, want 5", got)
	}
}

// S11: Parent changes a fixed quantity (an empty "a day" means fixed)
func TestChangeFixedBasicQuantity(t *testing.T) {
	app, database := newTestApp(t)
	id := basicIDOf(t, database, 1, "Shoes")

	if rec := saveBasic(t, app, id, "Shoes", "", "3"); rec.Code != http.StatusSeeOther {
		t.Fatalf("save: status = %d, want 303", rec.Code)
	}

	assertBasicShown(t, app, database, 1, "Shoes", basicRow{fixed: 3})
	for _, days := range []int{1, 14} {
		if got := quantityOf(t, database, createTripOf(t, app, database, days), 1, "Shoes"); got != 3 {
			t.Errorf("new %d-day trip has %d shoes, want 3", days, got)
		}
	}
}

// S7, S9, S12 (and bad numbers): a basic that cannot be saved stays as it
// was, and the parent sees why. Adding has its own table (TestBasicAddRejected).
func TestBasicSaveRejected(t *testing.T) {
	const nothing = "Enter how many a day, how many extra, or both."
	const needName = "Enter a name for the basic."
	const badNumber = "Up to 20 per field."
	cases := []struct {
		name          string
		member        int64
		basic, rename string
		perDay, fixed string
		wantMsg       string
	}{
		{"S7 empty name", 1, "Khakis", "", "", "1", needName},
		{"S7 name of only spaces", 1, "Khakis", "   ", "", "1", needName},
		{"S9 daily with nothing", 1, "Underwear", "Underwear", "0", "0", nothing},
		{"S9 daily left empty", 1, "Underwear", "Underwear", "", "", nothing},
		{"S9 fixed set to 0", 1, "Shoes", "Shoes", "", "0", nothing},
		{"S12 renamed to a name the member has", 3, "Cap", "hat", "", "1", "Child 1 already has Hat."},
		{"not a number", 1, "Underwear", "Underwear", "two", "1", badNumber},
		{"negative", 1, "Underwear", "Underwear", "1", "-1", badNumber},
		{"R7 daily over the cap", 1, "Underwear", "Underwear", "21", "1", badNumber},
		{"R7 extra over the cap", 1, "Underwear", "Underwear", "1", "21", badNumber},
		{"R7 a number too big to count", 1, "Underwear", "Underwear", "9223372036854775807", "1", badNumber},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			app, database := newTestApp(t)
			if c.basic == "Cap" {
				// S12 needs Child 1 to have both "Hat" and "Cap".
				insertBasic(t, database, c.member, "Cap", 0, 1)
			}
			id := basicIDOf(t, database, c.member, c.basic)
			wantPD, wantFx := basicStored(t, database, id)
			if c.member == 3 {
				c.wantMsg = memberNameByID(t, database, 3) + strings.TrimPrefix(c.wantMsg, "Child 1")
			}

			rec := saveBasic(t, app, id, c.rename, c.perDay, c.fixed)

			if rec.Code != http.StatusUnprocessableEntity {
				t.Fatalf("status = %d, want 422", rec.Code)
			}
			if !strings.Contains(rec.Body.String(), c.wantMsg) {
				t.Errorf("page does not say %q", c.wantMsg)
			}
			// The row stays open with what was typed, not the stored values.
			row := basicRowHTML(t, memberSection(t, rec.Body.String(), memberNameByID(t, database, c.member)), c.basic)
			if tag := inputTag(t, row, "name"); !strings.Contains(tag, `value="`+strings.TrimSpace(c.rename)+`"`) {
				t.Errorf("name field lost what was typed (%q): %s", c.rename, tag)
			}
			if got := basicNameStored(t, database, id); got != c.basic {
				t.Errorf("name stored %q, want unchanged %q", got, c.basic)
			}
			if pd, fx := basicStored(t, database, id); pd != wantPD || fx != wantFx {
				t.Errorf("stored %d/%d, want unchanged %d/%d", pd, fx, wantPD, wantFx)
			}
		})
	}
}

// insertBasic adds a basic last in a member's list, as the add form will.
func insertBasic(t *testing.T, database *sql.DB, member int64, name string, perDay, fixed int) {
	t.Helper()
	_, err := database.Exec(`INSERT INTO basic_item (family_member_id, position, name, name_key, per_day, fixed)
		VALUES (?, (SELECT max(position) + 1 FROM basic_item WHERE family_member_id = ?), ?, lower(?), ?, ?)`,
		member, member, name, name, perDay, fixed)
	if err != nil {
		t.Fatalf("insert basic %q: %v", name, err)
	}
}

// S2: Parent changes a basic item's name
func TestRenameBasic(t *testing.T) {
	app, database := newTestApp(t)
	id := basicIDOf(t, database, 1, "Khakis")

	if rec := saveBasic(t, app, id, "Chinos", "", "1"); rec.Code != http.StatusSeeOther {
		t.Fatalf("save: status = %d, want 303", rec.Code)
	}

	section := memberSection(t, get(t, app, basicsURL), memberNameByID(t, database, 1))
	if !strings.Contains(section, ">Chinos<") || strings.Contains(section, ">Khakis<") {
		t.Errorf("basics page should show Chinos and not Khakis for Parent 1")
	}
	tripID := createTripOf(t, app, database, 3)
	if got := quantityOf(t, database, tripID, 1, "Chinos"); got != 1 {
		t.Errorf("new trip has %d Chinos, want 1", got)
	}
	var khakis int
	database.QueryRow("SELECT count(*) FROM item WHERE trip_id = ? AND name = 'Khakis' AND family_member_id = 1", tripID).Scan(&khakis)
	if khakis != 0 {
		t.Errorf("new trip still has Khakis")
	}
}

// Changing only the capitalisation of a basic's own name is not a duplicate
// (the check ignores capitalisation but not against the basic itself), and
// saving an unchanged name with new quantities works.
func TestRenameBasicToOwnNameInOtherCase(t *testing.T) {
	app, database := newTestApp(t)
	id := basicIDOf(t, database, 3, "Hat")

	if rec := saveBasic(t, app, id, "hat", "", "1"); rec.Code != http.StatusSeeOther {
		t.Fatalf("save: status = %d, want 303", rec.Code)
	}
	if got := basicNameStored(t, database, id); got != "hat" {
		t.Errorf("name stored %q, want %q", got, "hat")
	}
}

// The same name for different family members is allowed.
func TestRenameBasicToAnotherMembersName(t *testing.T) {
	app, database := newTestApp(t)
	id := basicIDOf(t, database, 1, "Khakis")

	// "Pyjama dress" is Child 1's basic; Parent 1 has none, so it is free.
	if rec := saveBasic(t, app, id, "Pyjama dress", "", "1"); rec.Code != http.StatusSeeOther {
		t.Fatalf("save: status = %d, want 303", rec.Code)
	}
}

// R1: the database refuses two basics of one name (ignoring capitalisation)
// for one family member, so a double tap or two parents at once cannot save a
// pair that would make every new trip fail.
func TestBasicNamesAreUniquePerMemberInTheDatabase(t *testing.T) {
	_, database := newTestApp(t)

	_, err := database.Exec(`INSERT INTO basic_item (family_member_id, position, name, name_key, per_day, fixed)
		VALUES (3, 99, 'HAT', 'hat', 0, 1)`)
	if err == nil {
		t.Fatal("a second Hat for Child 1 was saved")
	}
	if _, err := database.Exec(`INSERT INTO basic_item (family_member_id, position, name, name_key, per_day, fixed)
		VALUES (3, 99, 'Sunglasses', 'sunglasses', 0, 1)`); err != nil {
		t.Fatalf("a different name was refused: %v", err)
	}
	var blank int
	database.QueryRow("SELECT count(*) FROM basic_item WHERE name_key != lower(name)").Scan(&blank)
	if blank != 0 {
		t.Errorf("%d seeded basics have a name_key that is not their lowercased name", blank)
	}
}

// S6: Existing trips are not changed (slice 2: changing a quantity)
func TestChangingBasicQuantityLeavesExistingTrips(t *testing.T) {
	app, database := newTestApp(t)
	tripID := createTripOf(t, app, database, 3)
	before := tripItems(t, database, tripID)

	saveBasic(t, app, basicIDOf(t, database, 1, "Underwear"), "Underwear", "5", "4")

	after := tripItems(t, database, tripID)
	if len(after) != len(before) {
		t.Fatalf("trip has %d items, was %d", len(after), len(before))
	}
	for i := range before {
		if after[i] != before[i] {
			t.Errorf("item %d = %+v, was %+v", i+1, after[i], before[i])
		}
	}
}

// S6 (slice 3: renaming): a trip keeps the old name.
func TestRenamingBasicLeavesExistingTrips(t *testing.T) {
	app, database := newTestApp(t)
	tripID := createTripOf(t, app, database, 3)
	before := tripItems(t, database, tripID)

	saveBasic(t, app, basicIDOf(t, database, 1, "Khakis"), "Chinos", "", "1")

	after := tripItems(t, database, tripID)
	for i := range before {
		if after[i] != before[i] {
			t.Errorf("item %d = %+v, was %+v", i+1, after[i], before[i])
		}
	}
}

// R5: saving a basic the other parent has removed leaves the page as it is.
func TestSavingMissingBasicIsNotAnError(t *testing.T) {
	app, _ := newTestApp(t)

	rec := saveBasic(t, app, 99999, "Whatever", "1", "1")

	if rec.Code != http.StatusSeeOther || rec.Header().Get("Location") != basicsURL {
		t.Errorf("status = %d, location = %q, want 303 to %s", rec.Code, rec.Header().Get("Location"), basicsURL)
	}
}

// Q8: the edit form shows zeroes as empty fields, so an empty "a day" is what
// marks a fixed basic; only the row being edited has a form.
func TestBasicEditFormShowsZeroesEmpty(t *testing.T) {
	app, database := newTestApp(t)
	shoes := basicIDOf(t, database, 1, "Shoes")
	underwear := basicIDOf(t, database, 1, "Underwear")

	page := get(t, app, basicsURL+"?edit="+strconv.FormatInt(shoes, 10))
	body := basicRowHTML(t, memberSection(t, page, memberNameByID(t, database, 1)), "Shoes")
	if tag := inputTag(t, body, "per_day"); !strings.Contains(tag, `value=""`) {
		t.Errorf("fixed basic's daily field is not empty: %s", tag)
	}
	if tag := inputTag(t, body, "fixed"); !strings.Contains(tag, `value="2"`) {
		t.Errorf("fixed basic's extra field is not 2: %s", tag)
	}
	if tag := inputTag(t, body, "name"); !strings.Contains(tag, `value="Shoes"`) {
		t.Errorf("edit form's name field is not Shoes: %s", tag)
	}
	if n := strings.Count(page, "· edit basic"); n != 1 {
		t.Errorf("%d edit forms on the page, want 1", n)
	}

	page = get(t, app, basicsURL+"?edit="+strconv.FormatInt(underwear, 10))
	body = basicRowHTML(t, memberSection(t, page, memberNameByID(t, database, 1)), "Underwear")
	if tag := inputTag(t, body, "per_day"); !strings.Contains(tag, `value="1"`) {
		t.Errorf("daily basic's daily field is not 1: %s", tag)
	}
}

func basicsOf(t *testing.T, database *sql.DB, member int64) []string {
	t.Helper()
	rows, err := database.Query("SELECT name FROM basic_item WHERE family_member_id = ? ORDER BY position", member)
	if err != nil {
		t.Fatalf("list basics: %v", err)
	}
	defer rows.Close()
	var names []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			t.Fatalf("scan basic: %v", err)
		}
		names = append(names, n)
	}
	return names
}

func addBasicURL(member int64) string { return "/members/" + strconv.FormatInt(member, 10) + "/basics" }

// addBasic posts a member's add form: the same three fields as the edit form (R2).
func addBasic(t *testing.T, app *application, member int64, name, perDay, fixed string) *httptest.ResponseRecorder {
	t.Helper()
	return send(t, app, http.MethodPost, addBasicURL(member), url.Values{"name": {name}, "per_day": {perDay}, "fixed": {fixed}})
}

func removeBasic(t *testing.T, app *application, id int64) *httptest.ResponseRecorder {
	t.Helper()
	return send(t, app, http.MethodPost, basicSaveURL(id)+"/delete", nil)
}

// Every section has an add form, the Family bucket included.
func TestEachSectionHasAnAddForm(t *testing.T) {
	app, database := newTestApp(t)
	body := get(t, app, basicsURL)
	for _, id := range []int64{1, 2, 3, 4, familyBucketID} {
		section := memberSection(t, body, memberNameByID(t, database, id))
		if !strings.Contains(section, `action="`+addBasicURL(id)+`"`) {
			t.Errorf("section of member %d has no add form posting to %s", id, addBasicURL(id))
		}
	}
}

// S4: Parent adds a basic item
func TestAddBasic(t *testing.T) {
	app, database := newTestApp(t)

	if rec := addBasic(t, app, 3, "Sunglasses", "", "1"); rec.Code != http.StatusSeeOther {
		t.Fatalf("add: status = %d, want 303", rec.Code)
	}

	section := memberSection(t, get(t, app, basicsURL), memberNameByID(t, database, 3))
	last := strings.LastIndex(section, "<li")
	if !strings.Contains(section[last:], ">Sunglasses<") {
		t.Errorf("Sunglasses is not last in Child 1's basics")
	}
	assertBasicShown(t, app, database, 3, "Sunglasses", basicRow{fixed: 1})
	if got := quantityOf(t, database, createTripOf(t, app, database, 3), 3, "Sunglasses"); got != 1 {
		t.Errorf("new 3-day trip has %d sunglasses for Child 1, want 1", got)
	}

	// A second add goes after the first, even though positions have gaps.
	addBasic(t, app, 3, "Scarf", "", "1")
	names := basicsOf(t, database, 3)
	if n := len(names); names[n-2] != "Sunglasses" || names[n-1] != "Scarf" {
		t.Errorf("last two basics = %v, want Sunglasses then Scarf", names[n-2:])
	}
}

// S7, S9, S10 for adding: nothing is saved, and the parent sees why.
func TestBasicAddRejected(t *testing.T) {
	cases := []struct {
		name, basic   string
		perDay, fixed string
		wantMsg       string
	}{
		{"S7 empty name", "", "", "1", "Enter a name for the basic."},
		{"S7 name of only spaces", "   ", "", "1", "Enter a name for the basic."},
		{"S9 nothing to pack", "Sunglasses", "", "", "Enter how many a day, how many extra, or both."},
		{"S9 both set to 0", "Sunglasses", "0", "0", "Enter how many a day, how many extra, or both."},
		{"S10 same name as an existing basic", "hat", "", "1", " already has Hat."},
		{"not a number", "Sunglasses", "x", "1", "Up to 20 per field."},
		{"R7 over the cap", "Sunglasses", "", "21", "Up to 20 per field."},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			app, database := newTestApp(t)
			before := len(basicsOf(t, database, 3))

			rec := addBasic(t, app, 3, c.basic, c.perDay, c.fixed)

			if rec.Code != http.StatusUnprocessableEntity {
				t.Fatalf("status = %d, want 422", rec.Code)
			}
			if !strings.Contains(rec.Body.String(), c.wantMsg) {
				t.Errorf("page does not say %q", c.wantMsg)
			}
			if after := len(basicsOf(t, database, 3)); after != before {
				t.Errorf("Child 1 has %d basics, want %d", after, before)
			}
		})
	}
}

// The same name for different family members is allowed.
func TestAddBasicWithAnotherMembersName(t *testing.T) {
	app, database := newTestApp(t)

	// "Pyjama dress" is Child 1's basic; Parent 1 has none, so it is free.
	if rec := addBasic(t, app, 1, "Pyjama dress", "", "1"); rec.Code != http.StatusSeeOther {
		t.Fatalf("add: status = %d, want 303", rec.Code)
	}
	if n := len(basicsOf(t, database, 1)); n != 18 {
		t.Errorf("Parent 1 has %d basics, want 18", n)
	}
}

// S10: a double-tapped Add saves one basic and is not shown as an error (R1).
func TestDoubleTappedAddSavesOne(t *testing.T) {
	app, database := newTestApp(t)
	before := len(basicsOf(t, database, 3))

	first := addBasic(t, app, 3, "Sunglasses", "", "1")
	second := addBasic(t, app, 3, "Sunglasses", "", "1")

	// The second request finds the name taken. A request that checked just
	// before the first one saved would meet the database rule instead; both
	// must end without an error page, so only the first is asserted exactly.
	if first.Code != http.StatusSeeOther {
		t.Errorf("first add: status = %d, want 303", first.Code)
	}
	if second.Code != http.StatusSeeOther && second.Code != http.StatusUnprocessableEntity {
		t.Errorf("second add: status = %d, want a redirect or the duplicate message", second.Code)
	}
	if after := len(basicsOf(t, database, 3)); after != before+1 {
		t.Errorf("Child 1 has %d basics, want %d", after, before+1)
	}
}

// S5: Parent removes a basic item
func TestRemoveBasic(t *testing.T) {
	app, database := newTestApp(t)
	id := basicIDOf(t, database, 4, "Nappies")

	if rec := removeBasic(t, app, id); rec.Code != http.StatusSeeOther {
		t.Fatalf("remove: status = %d, want 303", rec.Code)
	}

	section := memberSection(t, get(t, app, basicsURL), memberNameByID(t, database, 4))
	if strings.Contains(section, ">Nappies<") {
		t.Errorf("basics page still lists Nappies for Child 2")
	}
	var nappies int
	database.QueryRow("SELECT count(*) FROM item WHERE trip_id = ? AND family_member_id = 4 AND name = 'Nappies'",
		createTripOf(t, app, database, 3)).Scan(&nappies)
	if nappies != 0 {
		t.Errorf("new trip gives Child 2 nappies")
	}
}

// R5: removing a basic the other parent already removed leaves the page as it is.
func TestRemovingMissingBasicIsNotAnError(t *testing.T) {
	app, database := newTestApp(t)
	id := basicIDOf(t, database, 4, "Nappies")
	removeBasic(t, app, id)

	rec := removeBasic(t, app, id)

	if rec.Code != http.StatusSeeOther || rec.Header().Get("Location") != basicsURL {
		t.Errorf("status = %d, location = %q, want 303 to %s", rec.Code, rec.Header().Get("Location"), basicsURL)
	}
}

// S6 (slice 4: adding and removing): an existing trip keeps its items.
func TestAddingAndRemovingBasicsLeaveExistingTrips(t *testing.T) {
	app, database := newTestApp(t)
	tripID := createTripOf(t, app, database, 3)
	before := tripItems(t, database, tripID)

	addBasic(t, app, 3, "Sunglasses", "", "1")
	removeBasic(t, app, basicIDOf(t, database, 4, "Nappies"))

	after := tripItems(t, database, tripID)
	if len(after) != len(before) {
		t.Fatalf("trip has %d items, was %d", len(after), len(before))
	}
	for i := range before {
		if after[i] != before[i] {
			t.Errorf("item %d = %+v, was %+v", i+1, after[i], before[i])
		}
	}
}

// Removing is not offered while a row is open for editing, like the trip page.
func TestRemoveIsLockedWhileEditing(t *testing.T) {
	app, database := newTestApp(t)
	editing := basicIDOf(t, database, 1, "Underwear")

	body := get(t, app, basicsURL+"?edit="+strconv.FormatInt(editing, 10))

	tag := buttonTag(t, body, "Remove Socks")
	if !strings.Contains(tag, "disabled") {
		t.Errorf("Remove is not disabled while another row is open: %s", tag)
	}
	if tag := buttonTag(t, get(t, app, basicsURL), "Remove Socks"); strings.Contains(tag, "disabled") {
		t.Errorf("Remove is disabled with nothing open: %s", tag)
	}
}

// R7: 20 is the most either field takes, and 20 itself is saved.
func TestBasicQuantityCapIsInclusive(t *testing.T) {
	app, database := newTestApp(t)
	id := basicIDOf(t, database, 1, "Underwear")

	if rec := saveBasic(t, app, id, "Underwear", "20", "20"); rec.Code != http.StatusSeeOther {
		t.Fatalf("save 20/20: status = %d, want 303", rec.Code)
	}
	if rec := addBasic(t, app, 3, "Sunglasses", "20", ""); rec.Code != http.StatusSeeOther {
		t.Fatalf("add 20: status = %d, want 303", rec.Code)
	}
	// A 14-day trip then holds 20 x 14 + 20 underwear, so the cap cannot
	// overflow the trip's arithmetic.
	if got := quantityOf(t, database, createTripOf(t, app, database, 14), 1, "Underwear"); got != 300 {
		t.Errorf("14-day trip has %d underwear, want 300", got)
	}
}

// R7: the quantity fields carry no min or max, so the browser's own check
// cannot stop the form before it is sent. Its small bubble was easy to miss on
// a phone (nothing seemed to happen); the page's message is the one feedback.
func TestBasicQuantityFieldsLeaveValidationToThePage(t *testing.T) {
	app, database := newTestApp(t)
	id := basicIDOf(t, database, 1, "Underwear")

	edit := get(t, app, basicsURL+"?edit="+strconv.FormatInt(id, 10))
	for what, body := range map[string]string{"edit row": basicRowHTML(t, memberSection(t, edit, memberNameByID(t, database, 1)), "Underwear"), "add form": get(t, app, basicsURL)} {
		for _, field := range []string{"per_day", "fixed"} {
			tag := inputTag(t, body, field)
			if strings.Contains(tag, " max=") || strings.Contains(tag, " min=") {
				t.Errorf("%s: %s field has a browser min or max: %s", what, field, tag)
			}
		}
	}
}

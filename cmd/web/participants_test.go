package main

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"slices"
	"strings"
	"testing"
)

// Milestone 07 (spec/milestones/07-trip-participants.md), trip participants
// S1 to S6.

// everyone ticks all four people on the new trip form: the form field is
// "members", one box per person, valued with the person's id.
var everyone = []string{"1", "2", "3", "4"}

const choosePersonMessage = "Choose at least one person who is going."

// createTripFor posts a valid 3-day trip with the given boxes ticked (nil
// means the form carries no member field at all).
func createTripFor(t *testing.T, app *application, members []string) *httptest.ResponseRecorder {
	t.Helper()
	form := url.Values{
		"destination":    {"Parainen"},
		"departure_date": {"2026-09-20"},
		"duration_days":  {"3"},
	}
	if members != nil {
		form["members"] = members
	}
	return send(t, app, http.MethodPost, "/trips", form)
}

// memberBoxes returns the tick boxes on the page by the person id they carry,
// as the <input> tag text.
func memberBoxes(t *testing.T, body string) map[string]string {
	t.Helper()
	boxes := map[string]string{}
	value := regexp.MustCompile(`value="([^"]*)"`)
	for _, tag := range regexp.MustCompile(`<input[^>]*>`).FindAllString(body, -1) {
		if !strings.Contains(tag, `name="members"`) {
			continue
		}
		m := value.FindStringSubmatch(tag)
		if m == nil {
			t.Fatalf("tick box without a value: %s", tag)
		}
		boxes[m[1]] = tag
	}
	return boxes
}

// ticked lists the ids of the boxes that are ticked, sorted.
func ticked(t *testing.T, body string) []string {
	t.Helper()
	var ids []string
	for id, tag := range memberBoxes(t, body) {
		if strings.Contains(tag, "checked") {
			ids = append(ids, id)
		}
	}
	slices.Sort(ids)
	return ids
}

// ownersOf returns the distinct owners of a trip's items, sorted.
func ownersOf(t *testing.T, database *sql.DB, tripID int64) []int64 {
	t.Helper()
	var owners []int64
	for _, it := range tripItems(t, database, tripID) {
		if !slices.Contains(owners, it.member) {
			owners = append(owners, it.member)
		}
	}
	slices.Sort(owners)
	return owners
}

// tripIDOf returns the id of the only trip, failing if there is not one.
func tripIDOf(t *testing.T, database *sql.DB) int64 {
	t.Helper()
	var id int64
	if err := database.QueryRow("SELECT id FROM trip").Scan(&id); err != nil {
		t.Fatalf("no trip was saved: %v", err)
	}
	return id
}

// S1: The new trip form starts with everyone going.
func TestNewTripFormTicksEveryone(t *testing.T) {
	app, _ := newTestApp(t)

	body := get(t, app, "/trips/new")

	if got, want := ticked(t, body), []string{"1", "2", "3", "4"}; !slices.Equal(got, want) {
		t.Errorf("ticked = %v, want %v", got, want)
	}
	if _, ok := memberBoxes(t, body)["100"]; ok {
		t.Errorf("the Family has a tick box")
	}
	// Labels come from the stored names, not hardcoded words.
	people, err := app.queries.ListPeople(t.Context())
	if err != nil {
		t.Fatalf("ListPeople: %v", err)
	}
	for _, p := range people {
		if !strings.Contains(body, p.Name) {
			t.Errorf("form does not show %q", p.Name)
		}
	}
}

// S2: A trip with everyone gets the same list as before. The expected list
// is the basics as they are in the database now (they are editable), with
// each quantity worked out here as per day x days + fixed.
func TestTripWithEveryoneHoldsAllBasics(t *testing.T) {
	app, database := newTestApp(t)

	rec := createTripFor(t, app, everyone)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want 303", rec.Code)
	}
	basics, err := app.queries.ListBasicItems(t.Context())
	if err != nil {
		t.Fatalf("ListBasicItems: %v", err)
	}
	var want []tripItem
	for _, b := range basics {
		want = append(want, tripItem{b.FamilyMemberID, b.Name, int(b.PerDay*3 + b.Fixed), "planned"})
	}
	got := tripItems(t, database, tripIDOf(t, database))
	if !slices.Equal(got, want) {
		t.Errorf("trip items differ from the basics:\n got %v\nwant %v", got, want)
	}
	if want := []int64{1, 2, 3, 4, familyBucketID}; !slices.Equal(ownersOf(t, database, tripIDOf(t, database)), want) {
		t.Errorf("owners = %v, want %v", ownersOf(t, database, tripIDOf(t, database)), want)
	}
}

// basicItemsOf returns the quantities the basics give the owners listed, for a
// 3-day trip, in trip order.
func basicItemsOf(t *testing.T, app *application, owners ...int64) []tripItem {
	t.Helper()
	basics, err := app.queries.ListBasicItems(t.Context())
	if err != nil {
		t.Fatalf("ListBasicItems: %v", err)
	}
	var want []tripItem
	for _, b := range basics {
		if slices.Contains(owners, b.FamilyMemberID) {
			want = append(want, tripItem{b.FamilyMemberID, b.Name, int(b.PerDay*3 + b.Fixed), "planned"})
		}
	}
	return want
}

// S3: Only the adults go.
// S4: Family basics need both children.
func TestTripGetsBasicsOfTickedPeopleOnly(t *testing.T) {
	tests := []struct {
		name    string
		members []string
		owners  []int64 // whose basics the trip holds
	}{
		{"only the adults", []string{"1", "2"}, []int64{1, 2}},
		{"one child is not enough for the Family", []string{"1", "2", "3"}, []int64{1, 2, 3}},
		{"the other child is not enough either", []string{"1", "2", "4"}, []int64{1, 2, 4}},
		{"both children bring the Family", []string{"3", "4"}, []int64{3, 4, familyBucketID}},
		{"one person", []string{"2"}, []int64{2}},
		{"a repeated id counts once", []string{"3", "3"}, []int64{3}},
		{"odd values are ignored, valid ones kept", []string{"1", "99", "abc", "100", ""}, []int64{1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app, database := newTestApp(t)

			rec := createTripFor(t, app, tt.members)

			if rec.Code != http.StatusSeeOther {
				t.Fatalf("status = %d, want 303", rec.Code)
			}
			got := tripItems(t, database, tripIDOf(t, database))
			want := basicItemsOf(t, app, tt.owners...)
			if len(want) == 0 {
				t.Fatalf("test setup: no basics for owners %v", tt.owners)
			}
			if !slices.Equal(got, want) {
				t.Errorf("trip items:\n got %v\nwant %v", got, want)
			}
		})
	}
}

// S5: A trip needs at least one person. Also covers odd form data (R2) and
// an old cached form with no tick boxes at all (R3).
func TestTripNeedsAPerson(t *testing.T) {
	tests := []struct {
		name    string
		members []string
	}{
		{"nobody ticked, or an old form with no boxes", nil},
		{"an empty value", []string{""}},
		{"an unknown id", []string{"99"}},
		{"not a number", []string{"abc"}},
		{"the Family bucket sent directly", []string{"100"}},
		{"a number too big to read", []string{"99999999999999999999"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app, database := newTestApp(t)

			rec := createTripFor(t, app, tt.members)

			if rec.Code != http.StatusUnprocessableEntity {
				t.Errorf("status = %d, want 422", rec.Code)
			}
			if n := count(t, database, "trip"); n != 0 {
				t.Errorf("%d trips saved, want 0", n)
			}
			body := rec.Body.String()
			if !strings.Contains(body, choosePersonMessage) {
				t.Errorf("page does not say %q", choosePersonMessage)
			}
			if got := ticked(t, body); len(got) != 0 {
				t.Errorf("ticked = %v, want nobody", got)
			}
			if !strings.Contains(inputTag(t, body, "destination"), `value="Parainen"`) {
				t.Errorf("destination was not kept")
			}
			if !strings.Contains(inputTag(t, body, "departure_date"), `value="2026-09-20"`) {
				t.Errorf("departure date was not kept")
			}
			if !strings.Contains(inputTag(t, body, "duration_days"), `value="3"`) {
				t.Errorf("days were not kept")
			}
		})
	}
}

// The person message comes after the other form messages (form order).
func TestPersonMessageComesLast(t *testing.T) {
	app, _ := newTestApp(t)

	rec := send(t, app, http.MethodPost, "/trips", url.Values{
		"destination":    {""},
		"departure_date": {"2026-09-20"},
		"duration_days":  {"3"},
	})

	body := rec.Body.String()
	first, second := strings.Index(body, "Enter a destination."), strings.Index(body, choosePersonMessage)
	if first < 0 || second < 0 || first > second {
		t.Errorf("destination message at %d, person message at %d; want both, destination first", first, second)
	}
}

// S6: The ticks survive a refused form, and a failed save.
func TestTicksSurviveARefusedForm(t *testing.T) {
	app, _ := newTestApp(t)

	rec := send(t, app, http.MethodPost, "/trips", url.Values{
		"destination":    {""},
		"departure_date": {"2026-09-20"},
		"duration_days":  {"3"},
		"members":        {"2", "3"},
	})

	body := rec.Body.String()
	if !strings.Contains(body, "Enter a destination.") {
		t.Errorf("page does not ask for a destination")
	}
	if got, want := ticked(t, body), []string{"2", "3"}; !slices.Equal(got, want) {
		t.Errorf("ticked = %v, want %v", got, want)
	}
}

func TestTicksSurviveAFailedSave(t *testing.T) {
	app, database := newTestApp(t)
	if _, err := database.Exec(`CREATE TRIGGER refuse_item BEFORE INSERT ON item
		WHEN NEW.name = 'Hairbrush' BEGIN SELECT RAISE(ABORT, 'forced failure'); END`); err != nil {
		t.Fatalf("create trigger: %v", err)
	}
	// Parent 2 has a Hairbrush basic, so the failure is reached for a ticked person.

	rec := createTripFor(t, app, []string{"2", "3"})

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "The trip could not be created. Please try again.") {
		t.Fatalf("page does not explain that the trip was not created")
	}
	if n := count(t, database, "trip"); n != 0 {
		t.Errorf("%d trips saved, want 0", n)
	}
	if got, want := ticked(t, body), []string{"2", "3"}; !slices.Equal(got, want) {
		t.Errorf("ticked = %v, want %v", got, want)
	}
	if n := count(t, database, "item"); n != 0 {
		t.Errorf("%d items saved, want 0", n)
	}
}

// The browser does not block the form with its own popups, which phones show
// small and easy to miss: every problem comes back in the page's red box.
func TestFormShowsEveryMessageTogether(t *testing.T) {
	app, _ := newTestApp(t)

	if form := get(t, app, "/trips/new"); !strings.Contains(form[strings.Index(form, "<form"):], "novalidate") {
		t.Errorf("the new trip form lets the browser block the submit")
	}

	rec := send(t, app, http.MethodPost, "/trips", url.Values{})

	body := rec.Body.String()
	for _, msg := range []string{"Enter a destination.", "Choose a departure date.", choosePersonMessage} {
		if !strings.Contains(body, msg) {
			t.Errorf("page does not say %q", msg)
		}
	}
}

// childIDs (app.go) rests on the seed order: people are ids 1 to 4 and the
// children are the last two. Reseeding or reordering people must fail here,
// not quietly send the Family basics to the wrong people.
func TestChildIDsAreTheLastTwoSeededPeople(t *testing.T) {
	app, _ := newTestApp(t)

	people, err := app.queries.ListPeople(t.Context())
	if err != nil {
		t.Fatalf("ListPeople: %v", err)
	}
	var ids []int64
	for _, p := range people {
		ids = append(ids, p.ID)
	}
	if want := []int64{1, 2, 3, 4}; !slices.Equal(ids, want) {
		t.Fatalf("people ids = %v, want %v", ids, want)
	}
	if !slices.Equal(childIDs, ids[2:]) {
		t.Errorf("childIDs = %v, want the last two people %v", childIDs, ids[2:])
	}
}

package main

import (
	"context"
	"strings"
	"testing"

	"github.com/Zigelzi/roundtrip/internal/db"
)

// memberNames reads the four people in seed order, which is the order
// FAMILY_NAMES maps onto. It uses ListPeople, not ListFamilyMembers, because
// the Family bucket is not one of the names FAMILY_NAMES supplies.
func memberNames(t *testing.T, queries *db.Queries) []string {
	t.Helper()
	members, err := queries.ListPeople(context.Background())
	if err != nil {
		t.Fatalf("ListPeople: %v", err)
	}
	names := make([]string, 0, len(members))
	for _, member := range members {
		names = append(names, member.Name)
	}
	return names
}

// The real names are not in the repository, so a deployment supplies them via
// FAMILY_NAMES and the migration only seeds placeholders.
func TestFamilyNamesComeFromConfiguration(t *testing.T) {
	tests := []struct {
		name string
		list string
		want []string
	}{
		{
			name: "names replace the placeholders in seed order",
			list: "Ada,Bo,Cy,Di",
			want: []string{"Ada", "Bo", "Cy", "Di"},
		},
		{
			name: "unset list leaves every placeholder alone",
			list: "",
			want: []string{"Parent 1", "Parent 2", "Child 1", "Child 2"},
		},
		{
			name: "a blank entry leaves that one member alone",
			list: "Ada,,Cy,",
			want: []string{"Ada", "Parent 2", "Cy", "Child 2"},
		},
		{
			name: "surrounding spaces are trimmed",
			list: " Ada , Bo ,Cy,Di",
			want: []string{"Ada", "Bo", "Cy", "Di"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, database := newTestApp(t)
			queries := db.New(database)

			if err := db.ApplyFamilyNames(context.Background(), queries, tt.list); err != nil {
				t.Fatalf("ApplyFamilyNames(%q): %v", tt.list, err)
			}

			got := memberNames(t, queries)
			if strings.Join(got, ",") != strings.Join(tt.want, ",") {
				t.Errorf("family is %v, want %v", got, tt.want)
			}
		})
	}
}

// A typo that adds a fifth name should be loud, not silently ignored.
func TestFamilyNamesRejectsMoreNamesThanMembers(t *testing.T) {
	_, database := newTestApp(t)
	queries := db.New(database)

	err := db.ApplyFamilyNames(context.Background(), queries, "Ada,Bo,Cy,Di,Ed")
	if err == nil {
		t.Fatal("a name with no matching family member was accepted")
	}
	if !strings.Contains(err.Error(), "5") {
		t.Errorf("error %q does not point at the fifth entry", err)
	}
}

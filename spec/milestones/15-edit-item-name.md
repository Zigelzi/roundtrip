# Milestone 15: Edit an item's name

**Status:** done · **Branch:** feature/edit-item-name
<!-- Status: draft → red-teamed → in-progress → done -->

## Why
Flow: step 2, Plan the list (`../user-flows.md`). A typo in an item name today means removing the item and adding it again, with its quantity retyped. That is enough effort that typos get left in, and the names are what the family reads while packing and what later trips reuse. Fixing a name should take about as long as changing a quantity.

## Scope
- **In:** the row that opens when a parent taps an item on the trip page gets a Name field next to the quantity stepper, laid out like the basics page's edit row (Name field on top, stepper and Cancel/Save below). One Save changes the name, the quantity, or both.
- **Out:**
  1. Moving an item to another family member.
  2. Editing a name from the packing page (the packing page stays a place to pack, not to plan).
  3. Changing the name of the basics an item was copied from (an item has no link back to its basic).
  4. Name suggestions inside the edit field (they help adding, and a typo fix is the opposite of picking from a list).

## Plan
One vertical slice, in this order:
1. Tests for S1 to S13 (written before the code).
2. `UpdateItem` query (name, name_key, quantity) replacing `UpdateItemQuantity`, scoped to the trip like the other item queries; `make generate`.
3. The save route becomes `POST /trips/{id}/items/{itemID}` (it saves more than a quantity now), like the basics page's `POST /basics/{id}`.
4. The handler loads the trip page first and answers "Item not found" if the item is not on this trip, before any checking of the input. It then validates the name with the same rules as adding (not empty, no other item of that member with the same name ignoring capitalisation; the item being edited does not count against itself) and the quantity, and reports both problems at once.
5. A refused save keeps what was typed, including an empty name (`ItemEdit.Submitted`), so the field and the message agree.
6. A unique-constraint failure on update (another parent added the same name a moment earlier) is refused with the same "already has" message, not treated as success as adding does.
7. A save with no `name` field at all keeps the item's name (a stale tab from before the deploy).
8. The edit row is laid out like the basics page's: Name field, Quantity stepper with its label, Cancel and Save at the right. No autofocus on Name (most opens are quantity changes and the keyboard would cover the stepper), no suggestions list.
9. The row's tap label changes from "Change quantity of X" to "Edit X"; the existing tests that use the old label are updated.
10. Phone check.

## Acceptance conditions (BDD)

### S1: A parent opens an item and sees its name to edit
- Given Parent 1 has "Sunscren" with quantity 2 on a trip
- When a parent taps the item
- Then the open row shows a Name field containing "Sunscren" and the quantity 2

### S2: A parent fixes a typo
- Given Parent 1 has "Sunscren" with quantity 2
- When a parent opens it, changes the name to "Sunscreen" and saves
- Then Parent 1's list shows "Sunscreen" with quantity 2 and no "Sunscren"
- And the row is closed again at the same place on the page

### S3: A parent fixes only the capitalisation
- Given Parent 1 has "sunscreen"
- When a parent opens it, changes the name to "Sunscreen" and saves
- Then the item is saved as "Sunscreen" (it does not count as a duplicate of itself)

### S4: A name the member already has is refused
- Given Parent 1 has "Sunscreen" and "Sunscren"
- When a parent opens "Sunscren", changes the name to "sunscreen" and saves
- Then nothing is changed, the row stays open with the typed name, and the message says Parent 1 already has Sunscreen

### S5: An empty name is refused
- Given Parent 1 has "Sunscren"
- When a parent opens it, clears the name and saves
- Then nothing is changed and the row stays open with an empty name field and the message "Enter an item name."

### S6: The same name on another person is fine
- Given Parent 1 has "Sunscren" and Parent 2 has "Sunscreen"
- When a parent opens Parent 1's "Sunscren", changes it to "Sunscreen" and saves
- Then both lists show "Sunscreen"

### S7: Changing only the quantity still works
- Given Parent 1 has "Sunscreen" with quantity 2
- When a parent opens it, changes the quantity to 3 and saves, leaving the name alone
- Then the list shows "Sunscreen" with quantity 3

### S8: A rename keeps where the item is in packing
- Given Parent 1's "Sunscren" is packed, and Parent 1's "Hat" is prepared but not packed
- When a parent renames "Sunscren" to "Sunscreen" and "Hat" to "Cap"
- Then "Sunscreen" is still packed and "Cap" is still prepared, and the packing bar's counts are the same as before

### S9: The item was removed in the meantime
- Given a parent has an item's row open on one phone, and the other parent removes that item on theirs
- When the first parent saves
- Then they see "Item not found" and neither phone's list changes

### S10: A bad quantity keeps the typed name
- Given Parent 1 has "Sunscren" with quantity 2
- When a parent opens it, changes the name to "Sunscreen", sets the quantity to 0 and saves
- Then nothing is changed, and the row stays open with "Sunscreen" in the name field

### S11: Both problems are shown together
- Given Parent 1 has "Sunscreen" and "Sunscren"
- When a parent opens "Sunscren", changes the name to "sunscreen", sets the quantity to 0 and saves
- Then the row stays open and shows both the "already has" message and the quantity message

### S12: Another parent added the same name a moment earlier
- Given a parent has "Sunscren" open to rename it to "Sunscreen", and meanwhile the other parent adds "Sunscreen" for the same person
- When the first parent saves
- Then the rename is refused with the "already has Sunscreen" message, not saved and not an error page

### S13: An old page without a name field keeps the name
- Given Parent 1 has "Sunscreen" with quantity 2, and a page opened before this change is still on a phone
- When the old page saves a new quantity of 3, sending no name at all
- Then the list shows "Sunscreen" with quantity 3

## Test notes
- All scenarios are server-rendered, so they run as ordinary `go test` scenarios through the existing helpers. A refused save is a 422 with the row still open, like the quantity refusal today. S12 is tested by inserting the duplicate straight into the database after the page is loaded, which is what the other parent's add does.
- The existing quantity tests post only a `quantity` field; with S13 they keep working as they are, apart from the new row label.
- That focus, the keyboard, the layout at phone width and the packing bar behave on a phone is checked by hand.

## Decisions log
- Q1: the typo comes back on the next trip when the item was copied from a basic that has the same typo (an item has no link back to its basic). Accepted for now; the basic is fixed on the basics page.
- Q2: two phones editing the same item: last save wins, so a stale row can bring an old name back. Accepted, rare with two users.
- Q3: a save without a name field keeps the item's name (S13).
- Q4: names are editable on the trip page only, not on the packing page.
- Not handled, low risk for two users: a double space inside a name or a differently encoded "ä" passes the duplicate check; a capitalisation-only fix may still suggest the old spelling if another trip uses it.

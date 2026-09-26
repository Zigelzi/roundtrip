# Milestone 05: UI pass for long lists

**Status:** done · **Branch:** feature/ui-pass
<!-- Status: draft → red-teamed → in-progress → done -->

## Why
Flow: Plan what to pack and Pack (`../user-flows.md` steps 2 and 5).

Since milestone 04, a new trip starts with about 75 items, and the screens were built for lists of five. On a phone the items sit inside a card inside a padded page, so about a sixth of the screen width goes to empty margins, and once a parent scrolls down they lose the trip's name, the way to packing, and any sense of whose items they are looking at. After this milestone the items use the whole screen width, rows are a little easier to hit, the top of the page stays in view, and a parent can jump straight to a person's items instead of scrolling past everyone else's.

This is the "dedicated visual pass" that `../constitution.md` defers until the functional parts are far enough along, limited to what long lists need.

## Scope
**In:**
1. Full width on a phone: each person's items span the screen edge to edge, without the page margin and the card margin stacking up. On a wider screen the content stops at a comfortable maximum width (Q1).
2. Slightly taller item rows on both the trip page and the packing page, so a row is easier to hit (Q2).
3. A header that stays at the top while scrolling. Trip page: the trip's name, dates and the way to start packing. Packing page: the trip's name and the packing progress (Q3).
4. Jump to a person: the trip page header has one link per person (and Family) that scrolls to that person's items.
5. Each person's heading on the trip page says how many items they have.
6. Everything that already scrolls to a spot (opening a quantity edit, saving it, cancelling) still lands where the row can be seen, not hidden under the new header.

**Out:**
1. Searching or filtering items by name.
2. Grouping items into categories such as "Hygiene" (still `04` Out 8).
3. Sorting or reordering items.
4. Collapsing a person's section on the trip page (Q4).
5. A layout designed for desktop or tablet. Wide screens only get the width cap (Q1).
6. Changing what any control does. This milestone moves and resizes things; it adds no new actions beyond the jump links.
7. Adding Alpine.js or any other client-side library.

## Acceptance conditions (BDD)
<!-- Most of this milestone is how things look and feel on a phone, which a Go test cannot see. Those scenarios are marked manual; the part a test can observe is still tested. -->

### S1: Items use the full width of the phone
- Given a trip with items for everyone
- When the parent opens the trip page or the packing page on a phone
- Then each person's items reach from one edge of the screen to the other, with only a small gap so text does not touch the edge
- Verification: manual

### S2: Item rows are easier to hit
- Given a trip with items
- When the parent taps near the top or bottom edge of a row on the packing page
- Then the tap packs that item, not the one above or below
- Verification: manual

### S3: The trip page header stays in view
- Given a trip with about 75 items
- When the parent scrolls down to the Family items on the trip page
- Then the trip's name, dates and Start packing are still visible at the top of the screen
- Verification: manual

### S4: The packing page header stays in view
- Given a trip being packed
- When the parent scrolls down and packs an item near the bottom
- Then the trip's name and the updated progress are visible at the top without scrolling back up
- Verification: manual (that the progress updates is already covered by `02` tests)

### S5: Parent jumps to a person's items
- Given a trip with items for everyone
- When the parent taps Child 2 in the trip page header
- Then the page scrolls to Child 2's items, with the heading "Child 2" visible just below the header
- Verification: the links, their order and where they point are tested; the landing position is manual

### S6: Parent sees how many items each person has
- Given Parent 1 has 18 items on the trip
- When the parent opens the trip page
- Then Parent 1's heading shows "18 items"
- And after the parent adds or removes one of Parent 1's items, the heading shows the new count without reloading the page

### S7: Editing a quantity still lands on the row
- Given a trip with about 75 items
- When the parent opens the quantity edit for an item halfway down, then saves it
- Then after each step the row is visible below the header, not hidden under it
- Verification: manual

### S8: Adding an item keeps the field in view
- Given the parent is adding items for Child 1 with the phone keyboard open
- When they add an item
- Then the name field is still visible between the header and the keyboard, ready for the next item
- Verification: manual

## Open questions
None open.

## Decisions log
Open questions resolved (Human-PM, 2026-09-26):
- **Q1 Wide screens:** full width means dropping the stacked padding on a phone. On a desktop the content keeps a maximum width.
- **Q2 Row heights:** 56px on the trip page and 64px on the packing page, accepted at the slice 1 phone check.
- **Q3 Packing filter:** not sticky. Only the trip's name and the progress stick on the packing page.
- **Q4 Collapsing sections:** out. The jump links cover getting to a person's items.
- **Q5 Numbering:** this is milestone 05. Other spec files refer to the activity milestones generally ("later milestones") rather than by number, since more milestones may land in between based on testing and use.

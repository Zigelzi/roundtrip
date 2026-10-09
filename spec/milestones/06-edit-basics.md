# Milestone 06: Edit basics

**Status:** done · **Branch:** feature/edit-basics
<!-- Status: draft → red-teamed → in-progress → done -->

## Why
Flow: Define the basics (`../user-flows.md` step 1), rated the most problematic step on 2026-09-27.

Each person's basics are fixed in the app since milestone 04, so changing one ("Child 1 has outgrown nappies") needs a code change. Until then the parents correct the same rows on every new trip instead. After this milestone a parent can view and edit each person's basics in the app, so that:

1. A new trip's list needs fewer edits, because the basics already fit each person.
2. The basics stay accurate as the family's needs change.
3. Both parents can review and update them together, without code changes.

This also resolves [A10](../activity-catalogue.md#open-items) ("parents adjusting their basics is a later milestone") and is the prerequisite for the post-trip review (idea 2 in `../user-flows.md`).

## Scope
**In:**
1. From the app's start page, a parent can open a page that shows the basics of each family member, including the Family bucket.
2. Changing a basic item's name and quantity.
3. Adding a new basic item to a family member including family bucket.
4. Removing a basic item from a family member including family bucket.
5. Changes apply only to trips created after the change. Existing trips keep their items exactly as they were.
6. Updating [`../domain-model.md`](../domain-model.md): parents change the basics in the app (the catalogue is a historical record, Out 8), names are unique per family member, and a basic with nothing a day is a fixed-quantity basic (R4).

**Out:**
1. Changing trips that already exist when the basics change (by design, see In 5).
2. Reordering basics, or moving a basic from one family member to another.
3. Categories within a person's basics (idea 6 in `../user-flows.md`).
4. Season-aware basics ([A12](../activity-catalogue.md#open-items)).
5. Choosing or skipping basics per trip.
6. The post-trip review (idea 2 in `../user-flows.md`).
7. Adding or removing family members.
8. Keeping `../activity-catalogue.md` in sync with the app. The app is the source of truth for the basics; the catalogue's Basics section is a historical record from 2026-09-25.
9. Undo or a history of changes to the basics.
10. Switching a basic between a fixed quantity and a daily one as a supported action (Q9). Remove it and add it again instead.
11. An upper limit on the trip page's own quantity field (R7 caps only the basics forms; the trip page is unchanged).

## Slice plan
Ordered by risk, then reuse: the Q8 form is the biggest unknown, so it is on the phone first; the name and quantity rules exist before adding reuses them.

| Slice | Scenarios | Phone check |
|---|---|---|
| 1. View basics | S1 | The page opens from the start page, fixed and daily quantities read right, and Child 2's basics are quick to find (R8) |
| 2. Edit quantities | S3, S8, S11, S9, S6 | Does editing a fixed basic feel wrong (Q8)? A new trip gets the new numbers |
| 3. Rename | S2, S7, S12 | Rename, then try an empty name and a duplicate |
| 4. Add and remove | S4, S10, S5 | Add, try a duplicate, remove, then create a trip |

- S6 gains a case in each slice from 2 on (changing, renaming, adding, removing).
- The database rule against duplicate names (R1) lands in slice 3, the first slice that can save a name, so slice 2 stays the size it is.
- Re-check after slice 1: if slice 2 looks too large, S9 moves to slice 3.

## Acceptance conditions (BDD)
<!-- Written so anyone in the family can read and challenge them: no code terms. How each one is tested lives under Test notes. -->

### S1: Parent sees everyone's basics
- Given the family has basics
- When the parent goes to the basics from the start page
- Then they see each family member's basics
- And each basic shows its quantity: either a fixed number, or so many a day plus any extra

### S2: Parent changes a basic item's name
- Given Parent 1 has the basic "Khakis"
- When the parent renames it to "Chinos"
- Then the basics page shows "Chinos" for Parent 1
- And the next new trip gives Parent 1 "Chinos"

### S3: Parent changes the daily quantity
- Given Parent 1 has underwear as a basic, set to 1 a day and 1 extra
- When the parent changes the daily quantity
- Then the basics page shows the new daily quantity
- And the extra is unchanged
- And the next new trip's underwear uses the new daily amount

### S4: Parent adds a basic item
- Given Child 1 has no "Sunglasses" basic
- When the parent adds "Sunglasses" to Child 1, with how many a day and how many extra
- Then the basics page lists it last in Child 1's basics
- And the next new trip gives Child 1 sunglasses

### S5: Parent removes a basic item
- Given Child 2 has the basic "Nappies"
- When the parent removes it
- Then the basics page no longer lists it
- And the next new trip gives Child 2 no nappies

### S6: Existing trips are not changed
- Given a trip created before the basics changed
- When the parent renames, changes, adds or removes a basic
- Then that trip's items are exactly as they were

### S7: A basic item cannot be saved without a name
- Given Parent 1 has the basic "Khakis"
- When the parent clears its name and saves
- Then it is not saved
- And the basic stays "Khakis"
- And the parent sees that a name is needed

### S8: Parent changes the extra quantity
- Given Parent 1 has underwear as a basic, set to 1 a day and 1 extra
- When the parent changes how many extra
- Then the basics page shows the new extra amount
- And the daily amount is unchanged
- And the next new trip's underwear uses the new extra amount

### S9: A basic item cannot be saved with nothing to pack
- Given Parent 1 has underwear as a basic, set to 1 a day and 1 extra
- When the parent sets both the daily quantity and the extra to none and saves
- Then it is not saved
- And the underwear quantities stay as they were
- And the parent sees that a basic needs something to pack

### S10: The same basic cannot be added twice to a family member
- Given Child 1 has the basic "Hat"
- When the parent adds "hat" to Child 1
- Then nothing is saved
- And the parent sees that Child 1 already has it

### S11: Parent changes a fixed quantity
- Given Parent 1 has shoes as a basic, set to 2
- When the parent changes the quantity
- Then the basics page shows the new quantity
- And the next new trip gives Parent 1 that many shoes, whatever the trip's length

### S12: A basic cannot be renamed to a name the family member already has
- Given Child 1 has the basics "Hat" and "Cap"
- When the parent renames "Cap" to "hat"
- Then it is not saved
- And the basic stays "Cap"
- And the parent sees that Child 1 already has it

### Test notes
- The trip-side arithmetic (so many a day x trip days + extra) is milestone 04's contract, tested by 04/S2. This milestone's tests only prove that a new trip uses the saved values, so the numbers live here and not in the scenarios.
- S3: underwear 1 a day + 1 extra, change to 2 a day; a new 3-day trip gets 7.
- S8: underwear 1 a day + 1 extra, change to 2 extra; a new 3-day trip gets 5.
- S1: each family member's basics are listed in the order a new trip gets them, and the Family bucket is included. A fixed basic shows only its number (shoes: "2"); a daily one shows "1 / day", plus "+ 1 extra" when it has extra (Human-PM changed "a day" to "/ day" after seeing it).
- S11: shoes 2, change to 3; a new 1-day trip and a new 14-day trip both get 3.
- S4: the add form has the same "a day" and "extra" fields as the edit form (R2). Sunglasses: none a day, 1 extra; a new 3-day trip gets 1. The rules of S7 and S9 apply to adding too, and are tested for it.
- S7, S9, S10, S12: one table-driven test, one case per scenario, each asserting its own message.
- S7: a name of only spaces counts as empty (Human-PM).
- S9: covers both kinds: a daily basic with none a day and no extra, and a fixed basic set to 0.
- S10, S12: the check ignores capitalisation ("hat" vs "Hat"), the same rule as items on a trip (01/S11). The same name for different family members is allowed. Changing the capitalisation of a basic's own name ("hat" to "Hat") is not a duplicate and is saved. The S12 test adds "Cap" to Child 1 first if the basics do not have it.
- S10: a double-tapped Add saves one basic, not two, and is not shown as an error (R1, same as the trip page).
- R7: 21 in either field is refused with "Up to 20 per field.", and 20 is saved. A 14-day trip with 20 / day + 20 extra holds 300, so the cap cannot overflow the trip's arithmetic.
- Saving or removing a basic the other parent has already removed leaves the page as it is, with no error page (R5).


## Design notes
Settled from the red-team pass; technical, so the implementer's call, recorded so review can check them.

- **Unique names (R1).** A migration adds `name_key` to `basic_item` with `UNIQUE (family_member_id, name_key)`, the same rule as `item`. Without it, a double-tapped Add or two parents adding at once could save two rows of one name, and then every new trip fails, because trip creation copies the basics in one transaction and the item table refuses the duplicate. The app still checks first, to show the S10 and S12 message; the database rule catches the race. The migration fills `name_key` for the seeded rows with SQL `lower()`, which is safe only because the seed is ASCII; from then on keys come from Go's `itemKey`, as for items.
- **Position of an added basic (Q2, R6).** Computed inside the insert statement (the highest position for that family member plus 1), not read first and written second, so two adds at once cannot clash on `UNIQUE (family_member_id, position)`. Removing leaves a gap in the numbers, which is harmless: only the order matters.
- **Missing basic (R5).** Editing or removing a basic id that no longer exists is not an error, the same as removing a trip item.

## Open questions
<!-- For the red-team pass (workflow step 2) and Human-PM to resolve (step 3). -->
- None open.

## Decisions log
- **Q1 Quantity editing:** both numbers are editable, as separate scenarios (S3 daily, S8 extra). Scenarios describe the change in plain words; the numbers live in Test notes (Human-PM).
- **Q2 Position of a new basic:** last in that family member's list (Human-PM).
- **Q3 Removing:** no confirmation, the same as removing an item on a trip (Human-PM).
- **Q4 Duplicate names:** not allowed within one family member, the same rule as items on a trip (Human-PM). S10 (adding) and S12 (renaming).
- **Q5 Both parents editing at once:** the last save wins (Human-PM).
- **Q6 Page patterns:** reuse the trip page's patterns, at least initially (Human-PM).
- **Q7 Label for the fixed number:** "extra", to start with. For underwear it reads as spares, for shoes as the whole quantity (Human-PM).
- **Q8 Fixed or daily on the page:** the page shows two kinds of quantity: a fixed number (shoes: 2) or so many a day plus any extra (underwear: 1 a day + 1 extra). Storage stays one shape, a daily number plus a fixed one, with a fixed basic having nothing a day. Human-PM's instinct is that the edit form should also treat them as two kinds (pick fixed or daily first), but the simplest form ships first: one form with "a day" and "extra" fields, where an empty "a day" means a fixed basic. The phone check of the editing slice judges whether that feels wrong; if it does, the two-kind form is a follow-up. This also narrows Q7: on the list, "extra" appears only on daily basics.
- **Q8 phone check verdict:** the simple form works as a starting point, but editing a basic feels slightly confusing, and defining a fixed value (leave "/ day" empty) is not natural. Follow-up for a later milestone: the two-kind form from Q8 (pick fixed or daily first), which would also make switching kind (Q9) a supported action (Human-PM, 2026-10-09).
- **Q9 Switching kind:** deferred. No scenario; the supported way is to remove the basic and add it again (Out 10). The simple form of Q8 does not block a switch by editing either; blocking it would be extra work for no user value (Human-PM).
- **Q9** confirmed: deferring means not blocked, just not promised (Human-PM).
- **R1 Duplicate basics break trip creation:** accepted. A database rule makes names unique per family member (Design notes).
- **R2 Quantity when adding:** accepted. The add form has the same "a day" and "extra" fields as the edit form (S4).
- **R3 "Spare" in the scenarios:** accepted. Scenarios use "extra", the label on the page.
- **R4 Domain model:** accepted. In 6 widened.
- **R5 Saving a basic the other parent removed:** accepted, as a Test notes line.
- **R6 Two adds at once:** accepted, as a Design note.
- **R7 Upper limit on quantities:** first declined (a typo shows up on the trip page), then reversed after the code review: a number of about 19 digits would overflow when a trip is created and break trip creation for everyone. Each of "/ day" and "extra" takes 0 to 20; no family packs more of one thing, and the cap can be loosened if a real case turns up (Human-PM). Applies to the edit row and the add form only.
- **R8 Long page on a phone:** accepted, in the slice 1 phone check.
- **R9 S6 lists four actions in one When:** kept as is.
- **Size:** 12 scenarios, above the rough ceiling of 8, kept as one milestone: one user goal, four of the twelve are rejection rules tested together, and a split would leave the basics half editable (red-team, Human-PM).
- **Quantity message:** "Up to 20 per field." for an over-the-cap, negative or non-number quantity (Human-PM). Short and friendly; it is slightly off for a negative or a non-number, which are rare on a number field.
- **Daily label:** the list shows "1 / day + 1 extra", and the edit and add forms label the daily field "/ day" to match, e.g. `[ 1 ] / day + [ 1 ] extra` (Human-PM, after the slice 1 phone check). Scenarios keep the plain words "a day".

# Milestone 04: Trip basics

**Status:** done · **Branch:** feature/trip-basics
<!-- Status: draft → red-teamed → in-progress → done -->

## Why
Flow: Plan what to pack (`../user-flows.md` step 2).

Every trip starts with the same boring list: underwear, socks, toothbrushes, nappies, the stroller. Today a parent types all of it by hand on every trip and re-decides how many of each for the trip's length. After this milestone a new trip already holds each person's basics and the Family bucket's, with per-day items sized to the trip's duration, so the parent starts from a nearly complete list and only adds what is special about this trip.

The content and the reasoning behind it live in [`../activity-catalogue.md`](../activity-catalogue.md#basics). On the worked Parainen weekend, basics are 75 of 96 rows, which is why they ship before activities.

## Scope
**In:**
1. Creating a trip adds the basics for Parent 1, Parent 2, Child 1, Child 2 and the Family bucket, as listed in [the catalogue's Basics section](../activity-catalogue.md#basics) on 2026-09-25. From then on the app holds the basics (Q2).
2. Per-day quantities ([A2](../activity-catalogue.md#open-items)): a quantity is either fixed (2 shoes) or so many per day plus spares (1 underwear a day + 1 spare), counting days, not nights.
3. Basics are ordinary items: they start unpacked, and can be removed, packed and counted in the trip's progress exactly like items the parent adds by hand.
4. Changing an item's quantity on the trip page, for any item, generated or hand-added (Q3). About 75 rows arrive at once, and without this the only fix for "Underwear, 4, I want 5" is remove and re-add. One row is edited at a time (R1).
5. A trip is at most 14 days (R3). This keeps a typo from creating a trip with hundreds of items that cannot be deleted, and keeps every generated quantity to two digits (the largest is Nappies, 84).
6. The Create trip button is disabled after the first tap (R8), because a double tap now costs a 75-item trip that cannot be deleted.
7. Updating [`../domain-model.md`](../domain-model.md) with where the basics are stored.

**Out:**
1. Activities of any kind (later milestones).
2. Choosing, skipping or editing basics, per trip or per person ([A10](../activity-catalogue.md#open-items)). Basics are always on.
3. Season-aware basics ([A12](../activity-catalogue.md#open-items)).
4. Basics for trips created before this milestone ([A12](../activity-catalogue.md#open-items)).
5. Recomputing quantities if the trip's duration changes. There is no trip editing today, so this cannot happen yet.
6. Showing where an item came from (basics or hand-added).
7. Editing anything about an item other than its quantity (name, owner).
8. Grouping each person's items into categories such as "Hygiene" or "Outdoors" (Q4). Wanted later, not here.
9. Trips longer than 14 days (R3). A later improvement: decide how to allow them safely, for example an extra confirmation, or trip deletion or archiving ([A7](../activity-catalogue.md#open-items)).
10. Changing a packed item's status when its quantity changes (R2). It stays packed.

## Acceptance conditions (BDD)
<!-- Written so anyone in the family can read and challenge them: no code terms. How each one is tested lives under "Test notes" in the Design notes below. -->

### S1: A new trip already holds everyone's basics
- Given no trips
- When the parent creates a 3-day trip
- Then the trip already lists exactly the rows marked Basics in the catalogue's [worked example](../activity-catalogue.md#the-list) as it stood on 2026-09-25: 75 items (Parent 1 17, Parent 2 16, Child 1 19, Child 2 17, Family 6), with the same quantities, none packed
- And the people appear in the order Parent 1, Parent 2, Child 1, Child 2, Family, and each person's items in the order of the catalogue's Basics section

### S2: Per-day items are sized to the trip's length
- Given no trips
- When the parent creates a trip of the length below
- Then Parent 1's underwear, Child 1's underwear and Child 2's nappies have these quantities

| Trip length | Parent 1 underwear (1 a day + 1 spare) | Child 1 underwear (2 a day + 1 spare) | Child 2 nappies (6 a day) |
|---|---|---|---|
| 1 day | 2 | 3 | 6 |
| 3 days | 4 | 7 | 18 |
| 7 days | 8 | 15 | 42 |
| 14 days | 15 | 29 | 84 |

### S3: Fixed items do not scale
- Given no trips
- When the parent creates a 1-day, a 3-day or a 14-day trip
- Then Parent 1 always has 2 shoes and the Family always has 1 double stroller

### S5: Basics behave like any other item
- Given a new 3-day trip with its 75 basics, none packed
- When the parent removes Parent 1's belt, packs Parent 1's toothbrush, and adds "Sunscreen, 1" to Family
- Then the belt is gone, the toothbrush is packed, sunscreen is listed under Family, and the trip shows 1 of 75 packed

### S6: Existing trips are left alone
- Given a trip that was created before this update
- When the parent opens it after the update
- Then its list is exactly as it was: no basics have been added to it

### S7: A failed trip creation leaves nothing behind
- Given something goes wrong while the basics are being added to a new trip
- When the parent creates the trip
- Then the parent is back on the new trip form, with what they entered still filled in, and sees "The trip could not be created. Please try again."
- And the trip list is unchanged: no half-filled trip appears in it

### S8: Parent changes an item's quantity while planning
- Given a 3-day trip where Parent 1 has 4 underwear (not packed), a hand-added "Book, 1", and 1 toothbrush that is already packed
- When the parent taps each item's quantity, changes it and saves: underwear to 5, book to 2, toothbrush to 2
- Then the trip shows 5 underwear, 2 books and 2 toothbrushes, and still does when opened again
- And the toothbrush is still marked packed

### S9: An invalid quantity is refused
- Given Parent 1 has 4 underwear
- When the parent tries to save 0, -1, 2.5 or "abc"
- Then the parent sees "Quantity must be a whole number, at least 1.", the same message as when adding an item, and the underwear stays at 4

### S10: Only one item is edited at a time
- Given the parent is changing the quantity of Parent 1's underwear
- When the parent tries to change or remove any other item
- Then they can't: those buttons are greyed out
- And adding a new item still works
- And once the parent saves or cancels, every item can be changed and removed again, and cancelling leaves the underwear at 4
- And opening the item, saving and cancelling all keep the screen where it was: the item stays at the same height on the screen, and the page does not reload or jump
- Verification: manual for the screen position, on a phone.

### S11: A trip is at most 14 days
- Given the new trip form
- When the parent enters a trip as below
- Then the result is as below

| Entered as | Trip | Result |
|---|---|---|
| Number of days | 14 days | Created |
| Return date | leaving 1 June, returning 14 June (14 days) | Created |
| Number of days | 15 days | Not created; the parent sees "A trip can be at most 14 days." |
| Return date | leaving 1 June, returning 15 June (15 days) | Not created; the parent sees "A trip can be at most 14 days." |

### S12: A double-tapped Create makes one trip
- Given the new trip form, filled in
- When the parent taps Create twice quickly
- Then only one trip is created
- Verification: manual, on a phone.

## Design notes
Settled from the red-team pass (R4 to R7); technical, so the implementer's call, recorded so review can check them.

- **Storage (Q2).** A new table `basic_item (family_member_id, position, name, per_day, fixed)` seeded by the migration (the S7 test inserts into it by these names), keyed by `family_member_id` (1 to 4 and the Family bucket, 100), never by name, because `FAMILY_NAMES` renames members at startup. Each row has an explicit position so trip items are inserted in catalogue order (Q4); `ListTripItems` orders by id. Generation reads this table directly, so it includes the Family bucket without going through `ListPeople`. Fix the `ListPeople` comment, which still names milestone 04 as its user; that is now 05.
- **Quantity.** quantity = per_day x duration_days + fixed. "Shoes, 2" is 0/day + 2; "Underwear, 1/day + 1" is 1/day + 1.
- **Name keys** are computed in Go with `itemKey`, not SQL `lower()`, which only folds ASCII.
- **One transaction (S7).** The trip and all its basics are saved in one database transaction: either everything is saved or nothing is. sqlc already generates `WithTx`; the application needs the `*sql.DB` alongside its queries.
- **Migration Down** drops the basics table. Items already generated stay, as ordinary items.

### Test notes
How the scenarios above are tested. Kept here so the scenarios stay readable without code knowledge.

- **S1:** the test holds its own typed-in copy of the 75 rows. It must not read the expected list back from the basics table, or it would only prove the code copies the table. The copy pins the seed: a later change to the basics updates this test too, which confirms the change was deliberate.
- **S1 to S5:** items are checked as (owner, name, quantity, status); the page shows a quantity as "4 pcs".
- **S2, S3:** one table-driven test, one case per trip length.
- **S6:** a trip from before this milestone is simulated by inserting it with the `CreateTrip` query directly, bypassing the handler; the test asserts it has no items.
- **S7:** the handler re-renders the trip form with the message instead of the bare "internal error" page `serverError` gives today. The test seeds a basics row with a duplicate name for one member, which the item table's uniqueness rule rejects, then asserts no trip and no items exist.
- **S8 to S10:** edit mode is the trip page with `?edit=ITEM_ID`: the edited row becomes a small form, every other row's Change and Remove buttons are rendered `disabled`, and Cancel is a link back to the plain trip page. Chosen because the lock has to span every member's section, which a per-member htmx swap cannot do, and because the page stays usable without JavaScript. The page contract the tests rely on is at the top of `cmd/web/quantity_test.go`. With JavaScript, opening, saving and cancelling swap the container holding every member's list in place (htmx `hx-select` on the full page reply) instead of navigating, so the view does not move; the automated test checks that those controls are wired that way, and the plain page loads remain the fallback.
- **S12:** manual only. The new trip form is a plain form, not htmx, so the guard is a few lines of script that disable the button on submit (and re-enable it if the browser restores the page from its back-forward cache). A Go test could only check that the script text is present, which proves nothing about the behaviour, so there is no automated test.

## Open questions
<!-- For the red-team pass (workflow step 2) and Human-PM to resolve (step 3). -->
- None open.

## Decisions log
Open questions resolved (Human-PM, 2026-09-25):
- **Q1 Unreviewed lists:** ship Parent 2's and the Family bucket's basics as they are. A later correction is a migration; trips created in between keep the old list.
- **Q2 Storage:** the basics are seeded into the database by a migration, because letting parents adjust their basics is a named later milestone and that needs them in the database. From this milestone on, the app is the source of truth for the basics. The catalogue is where they are discussed and may drift from what the app holds (Human-PM, after the red-team pass).
- **Q3 Quantity editing:** in scope (Scope In 4, S8, S9). It is small next to the generation work and makes the generated list correctable.
- **Q4 Order:** no preference, so catalogue order, the simplest option. Categories per person are wanted later and are out of scope here.

Red-team findings resolved (Human-PM, 2026-09-25):
- **R1 Editing pattern:** tap the quantity to edit that row only; while one row is being edited, the others cannot be edited or removed (S10).
- **R2 Packed items:** a quantity change keeps the item's status. Simplest for now; the known risk is an extra pair added after packing never reaching the bag.
- **R3 Trip length:** capped at 14 days (S11). Handling longer trips safely is a recorded later improvement (Out 9).
- **R8 Double tap:** the Create button is disabled after the first tap (S12).
- **R9 (Q5):** checked on the phone after the slice; Human-PM's experience of long single-person packing lists is that they feel fine, but four people at once is untested.
- **R10 Correcting basics:** no action. Schema changes are cheap at this stage and a correction does not need a clean Down. Once parents can edit their basics in the app, corrections stop being migrations at all.
- **R4 to R7** were technical and are folded into the Design notes and S1, S6, S7.

At 12 scenarios this milestone is over the rough ceiling of 8. Accepted: S8 to S10 are one small feature (quantity editing) and S11 and S12 are guards this milestone makes necessary; splitting them out would cost a second spec and review cycle.

Second red-team pass (scenario wording only), applied by the main session, 2026-09-25:
- **T1:** S11 is a table of four cases, 14 and 15 days, each entered as days and as a return date, with the dates spelled out so days are counted, not nights.
- **T3:** S1 names which rows of the worked example it means and the order of the people.
- **T4:** Scope In 1 points to the catalogue as it stood on 2026-09-25, matching Q2.
- **T5:** S7 names what the parent sees: the form again, with their entries and a message, instead of today's bare error page.
- **T6:** S9 includes a decimal (2.5).
- **T7:** S4 removed; its 1-day case is already a row in S2. The number stays unused.
- **T8:** Test notes wording fixed.
- **T2 (Human-PM):** adding a new item still works while one item's quantity is being changed (S10).
- **Q5 (Human-PM, phone check after slice 1, 2026-09-26):** the 75-row trip page is usable, so the collapse fallback is not built. The amount of information is a real problem to solve in a later milestone.
- **S10 screen position (Human-PM, slice 3 phone check, 2026-09-26):** jumping the edited row to the top of the screen felt erratic, since the row is usually 20 to 50% down. Opening, saving and cancelling now keep the view where it was, by swapping the lists in place with htmx (the approach that felt good in go-tiimit). Part of S10's Then, not a new scenario. The automatic focus on the quantity field raised the phone keyboard and shifted the view, so it was dropped: the parent taps the field or uses − / + (Human-PM, phone check).

Code review, round 1 (Opus sub-agent, 2026-09-26): no blocking findings.
- **C1:** fixed. A plain (no JavaScript) successful add closes the open row; the code comment now says so instead of claiming it always stays open.
- **C2:** accepted. With htmx, adding an item to the same person whose row is open resets the open field to the saved quantity, losing an unsaved typed value. Rare, and one tap to redo.
- **C3:** fixed. Tests added for adding while a row is open, for an edit id that is unknown or on another trip, and for the 404 when saving through another trip's URL.
- **C4:** declined. No `max="14"` on the days field, as decided for slice 2: the browser's own popup would replace the app's message, and the server check is the contract.
- **C5:** accepted. Saving an item that was removed in another tab does nothing visible. Only possible with two tabs open; not worth handling for two users.

# Milestone 01: Packing list

**Status:** done · **Branch:** feature/packing-list
<!-- Status: draft → red-teamed → in-progress → done -->

## Why
Flow: Plan what to pack (`../user-flows.md` step 2).
As a parent I want to start a new trip and write down the items each family member needs on it, so that we can keep track of everything we need to bring.

This comes before step 1 (planning activities) on purpose: a packing list is the most valuable part on its own, and activity-based planning later needs a packing list to feed into. It is intentionally a single-player milestone — learning the basics with one user before adding the second parent.

## Scope
**In:**
1. Creating a trip: destination (free text), departure date, duration in days. E.g. "Parainen, 20.9.2026, 3 days".
   1. A trip is always for the whole family; every family member participates.
   2. Duration counts calendar days including the departure day (20.9 + 3 days = 20.–22.9).
   3. Instead of the duration, the parent can pick a return date (handy for longer trips planned ahead, e.g. Christmas). Duration and return date update each other; changing the departure keeps the duration and moves the return date. When a return date is submitted it decides the duration. Only departure + duration are stored; the return date is derived.
   4. Departure date is today or later. "Today" is the current date in Finland (Europe/Helsinki). The departure date is a calendar date with no time zone; only "today" is derived from the current moment.
2. Listing trips: upcoming trips first (nearest at the top), then past trips.
3. Adding an item to a family member on a trip, with a quantity. E.g. "Underwear, Parent 1, 4 pcs".
   1. Each member has their own add form, so the member is implied.
   2. The name field suggests item names used on any earlier trip, so the same name is reused rather than retyped differently.
   3. A member can have the same item (name, ignoring upper/lower case) only once per trip.
4. Viewing a trip's items grouped by family member.
5. Removing an item from a member on a trip.
6. Adding and removing items in place (HTMX), so the name field keeps focus for quick entry.

**Out:**
1. Calculating quantity based on days.
2. Removing or editing trips; editing items (remove and re-add instead).
3. Cloning trips.
4. Packing templates.
5. Auth — single player, no login. Local by default; can be opened to the home network (`ADDR`) for phone testing and demos.
6. Trips with only specific family members, and managing family members (they are fixed).
7. Item states beyond "planned" (Prepared, Needs to be bought, Bought, Packed come in later milestones).
8. Relative time to/from departure ("in 5 days").

## Domain
Entities introduced here (Family member, Trip, Item) are defined in [`../domain-model.md`](../domain-model.md).

## Acceptance conditions (BDD)

### S1: Parent with no trips is prompted to create one
- Given there are no trips
- When the parent opens the front page
- Then they see a message that there are no trips yet
- And a link to create a trip

### S2: Parent sees their trips, upcoming first
- Given a past trip (departed before today) and two upcoming trips
- When the parent opens the front page
- Then they see every trip with its destination, dates from departure to return (e.g. 20.–22.9.2026) and duration in days
- And the upcoming trips are listed first, the nearest departure at the top
- And the past trip is listed after them

### S3: Create-trip form starts with sensible defaults
- Given the parent is on the front page
- When they open the create-trip form
- Then the departure date and the return date are empty
- And the duration is 3

### S4: Parent creates a trip
- Given the parent is on the create-trip form
- When they submit destination "Parainen", departure date 20.9.2026 and duration 3
- Then the trip is saved
- And they land on that trip's page

### S5: Parent creates a trip departing today
- Given today is 20.9.2026 in Finland
- And the parent is on the create-trip form
- When they submit a trip departing 20.9.2026
- Then the trip is saved

### S6: Invalid trip input is rejected
- Given the parent is on the create-trip form
- When they submit with any of: an empty destination, no departure date, a departure date before today, a return date before the departure date, a duration below 1 or not a whole number (with no return date)
- Then no trip is saved
- And the form is shown again with a message saying what is wrong
- And the values they typed are kept

### S7: Parent views a trip
- Given a trip "Parainen, 20.9.2026, 3 days"
- And Parent 1 has "Underwear, 4 pcs" on it and the other members have no items
- When the parent opens the trip's page
- Then they see the destination, the dates from departure to return (20.–22.9.2026) and the duration
- And a section for each of the 4 family members
- And Parent 1's section lists "Underwear" with quantity 4
- And every other member's section says they have no items yet
- And every member's section has a form to add an item, with the quantity defaulting to 1

### S8: Parent opens a trip that does not exist
- Given there is no trip with the requested id
- When the parent opens that trip's page
- Then they see a "trip not found" message with a link back to the front page
- And the response status is 404

### S9: Parent adds an item to a member
- Given a trip where Parent 1 has no items
- When the parent adds "Underwear" with quantity 4 to Parent 1
- Then Parent 1's section lists "Underwear" with quantity 4

### S10: Earlier item names are suggested
- Given an earlier trip has an item named "Sunscreen"
- When the parent opens a new trip's page
- Then the item name field suggests "Sunscreen"

### S11: The same item cannot be added twice to a member
- Given Parent 1 has "Underwear" on the trip
- When the parent adds "underwear" to Parent 1 on the same trip
- Then no new item is saved
- And they see a message that Parent 1 already has it

### S12: Different members can have the same item
- Given Parent 1 has "Underwear" on the trip
- When the parent adds "Underwear" to Parent 2 on the same trip
- Then Parent 2's section lists "Underwear"
- And Parent 1's section still lists "Underwear"

### S13: Invalid item input is rejected
- Given a trip
- When the parent adds an item with an empty name, or a quantity below 1 or not a whole number
- Then no item is saved
- And they see a message saying what is wrong

### S14: Parent removes an item from a member
- Given Parent 1 has "Underwear" on the trip
- When the parent removes it
- Then Parent 1's section no longer lists "Underwear"
- And other members' items are unchanged

### S15: Parent adds several items in a row without reselecting the field
- Given the parent is on a trip page
- When they type an item name in Parent 1's form and press Enter
- Then the item appears in Parent 1's list without the page reloading
- And the name field is empty and still selected, ready for the next item

Verification: manual on a phone (keyboard focus can't be observed by `go test`). Automated: adding or removing an item in place replies with only that member's list.

### S16: Parent adjusts the quantity with − and +
- Given the parent is on a trip page
- When they tap + or − next to a member's quantity field
- Then the quantity goes up or down by one, never below 1
- And the item is not added until they tap Add or press Enter

Verification: manual (tapping and the layout on a phone). Automated: the − and + buttons are present and cannot submit the form.

### S17: Parent sets the return date instead of the duration
- Given the parent is on the create-trip form
- When they submit departure 20.12.2026 and return date 27.12.2026
- Then the trip is saved with a duration of 8 days

### S18: Duration and return date update each other
- Given the parent is on the create-trip form with a departure date
- When they change the duration, the return date updates to match
- And when they change the return date, the duration updates to match
- And when they change the departure date, the duration stays and the return date moves

Verification: manual (the syncing runs in the browser). Automated: the return-date field is on the form.

## Open questions
<!-- For the red-team pass (workflow step 2) and Human-PM to resolve (step 3). -->
- None open.

## Decisions log
Red-team pass resolved (Human-PM, 2026-09-19): cut the "item already exists, add it instead?" prompt in favour of name suggestions; Item is the single entity (per trip + member) with states added later; one item per name per member per trip; fixed seeded members; removing items moved In; invalid-input scenarios added; relative time cut; single-player/local is intentional; add form per member (answers the original open question). Family member names settled (see domain model).

Implementation choices kept (Human-PM, 2026-09-19): plain HTML forms with post/redirect and a `#member-N` anchor (trip creation still works this way; item add/remove moved to HTMX in S15 once per-tap reloads hurt quick entry, with the plain forms kept as fallback); ✕ removes an item immediately, no confirm/undo; a trip already under way lists under past trips; item names compare case-insensitively including Ä/ä (enforced by a lowercased `name_key` + UNIQUE constraint); the Milestone 0 greeting stays on the front page (later removed at acceptance; the `app_user` table stays for future login).

Home-network access (Human-PM, 2026-09-19): reverses "local only" — the listen address is configurable via `ADDR` (default stays 127.0.0.1) so the app can be demoed and tested on a phone. Still no auth, so only on the trusted home network.

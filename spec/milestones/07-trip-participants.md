# Milestone 07: Trip participants

**Status:** done · **Branch:** feature/trip-participants
<!-- Status: draft → red-teamed → in-progress → done -->

## Why
Flow: Plan what to pack (`../user-flows.md` step 3).

A new trip always gets the basics of all four people and the Family bucket, even when only some of the family travels. For a week away with just the two adults, the parents must remove Child 1's and Child 2's basics and the Family's (stroller, nappies and so on) by hand. After this milestone the parent picks who is going when creating the trip, and the trip only gets the basics of those people, so the list is nearly right from the start.

## Scope
**In:**
1. The new trip form shows every person (Parent 1, Parent 2, Child 1, Child 2) with a checkbox, all ticked by default.
2. Creating a trip adds the basics only for the ticked people.
3. The Family bucket's basics are added only when both children are ticked.
4. A trip needs at least one person ticked (Q1).
5. Updating [`../domain-model.md`](../domain-model.md): a new trip gets the basics of the ticked people only, and the Family bucket's basics only when both children are ticked.

**Out:**
1. Changing who is going after the trip is created, and adding the basics of someone left out (Q2). Parents add such items by hand.
2. Choosing the participants for the Family bucket directly: it follows the children (In 3).
3. Remembering who went on earlier trips, or presets such as "adults only".
4. Per-person rules other than the Family one, for example "nappies only when a child under 2 is going".
5. Changing existing trips. They keep their items exactly as they were.
6. Remembering who is going. Nothing about the choice is stored, so the trip page, the packing page and the add forms still show every person, with empty sections for those not going (Q2).
7. Telling an old cached form from one where nobody is ticked (R3). Both get the "choose at least one person" message.

## Acceptance conditions (BDD)
<!-- Written so anyone in the family can read and challenge them: no code terms. How each one is tested lives under Test notes. -->

### S1: The new trip form starts with everyone going
- Given no trips
- When the parent opens the new trip form
- Then Parent 1, Parent 2, Child 1 and Child 2 are all ticked
- And there is no tick for the Family

### S2: A trip with everyone gets the same list as before
- Given no trips
- When the parent creates a 3-day trip with everyone ticked
- Then the trip lists the basics defined for each member including the family.

### S3: Only the adults go
- Given no trips
- When the parent creates a 3-day trip with only Parent 1 and Parent 2 ticked
- Then the trip lists Parent 1's and Parent 2's basics
- And the trip lists no items for Child 1, Child 2 or the Family

### S4: Family basics need both children
- Given no trips
- When the parent creates a trip with Parent 1, Parent 2 and Child 1 ticked, but not Child 2
- Then the trip lists the basics of the three ticked people
- And the trip lists no Family basics
- And Child 2 has no items

### S5: A trip needs at least one person
- Given the new trip form with everyone unticked
- When the parent taps Create
- Then no trip is created
- And the parent sees "Choose at least one person who is going."
- And the destination, dates and days they entered are still filled in
- And nobody is ticked

### S6: The ticks survive a refused form
- Given the parent ticked only Parent 2 and Child 1
- And the destination is empty
- When the parent taps Create
- Then the form is shown again with "Enter a destination."
- And only Parent 2 and Child 1 are still ticked

## Design notes
Settled from the red-team pass (R1, R2, R7, R13); technical, so the implementer's call, recorded so review can check them.

- **Who the children are (R1).** Child 1 and Child 2 are the member ids that the seed gives them (3 and 4), never found by name, because `FAMILY_NAMES` renames people at startup. The rule lives in one place next to `familyBucketID` in `cmd/web/app.go`.
- **Odd form data (R2).** Ids that are not a known person are ignored, including the Family bucket's id: Family is never ticked directly, it only follows the children. A repeated id counts once. If no valid id is left, the S5 message shows.
- **Filtering** happens inside the existing single transaction of `createTripWithBasics`, so a failed save still leaves nothing behind (04/S7) and adds no new rollback case.
- **Message order (R5).** The S5 message comes after the other form messages, in form order (destination, dates, days, people).
- **Validation stays on the server (R13).** HTML cannot require "at least one box", and a refused form reloads the page, so the double-tap guard and the 14-day cap need no change.

### Test notes
- **S1:** ticks are checked through the `checked` attribute on each person's box. The labels come from the stored names, not hardcoded words, so a renamed person still shows.
- **S2:** the expected list is built from the basics as they are in the database when the test runs, for all five owners, with no fixed row count, because the basics are editable since milestone 06. This avoids the trap of 04/S1 only because the generation under test (quantities, filtering) is computed independently in the test, not copied from the code.
- **S5, R2:** table-driven cases for nobody ticked, only unknown ids, only the Family bucket's id, a repeated id (one person), and no tick fields at all (an old cached form, R3).
- **S6:** also checked for the failed-save re-render (04/S7), not only the refused form (R7).

## Decisions log
Open questions resolved (Human-PM, 2026-10-09):
- **Q1 At least one person:** error (S5).
- **Q2 Remember who is going:** no. Only the basics are filtered at creation. Hiding people who are not going and adding them to an existing trip may come later.
- **Q3 Family rule:** both children, nothing else. If only one child comes, the Family basics are added by hand.
- **Q4 Selection widget:** plain checkboxes and CSS. JavaScript only if the phone check says it is needed.

Red-team findings resolved (Human-PM, 2026-10-09):
- **R3 Old cached form:** refuse it with the S5 message. The cost is one reload, once.
- **R4 S2 wording:** no row count; compares against the basics as they are.
- **R5 S5 message:** "Choose at least one person who is going."
- **R9 Empty sections:** S4 gained "Child 2 has no items". A separate children-only scenario is not added; S3 already shows that people not going get nothing.
- **R1, R2, R6 to R8, R10 to R14** were technical or wording and are folded into the scenarios, Design notes and Test notes.

Phone check (Human-PM, 2026-10-09):
- **Pill style:** the people boxes look like the packing page's member filter, for consistency. Ticked is a light fill with a dark outline and a check mark, not the filter's solid dark fill, so the solid dark Create trip button stays the one primary action.
- **Browser popups off:** with the name and dates empty, the browser's own "required" popup blocked the submit, so no message from the app showed and the page felt dead. The form is now `novalidate` and the server's messages show together in one red box. This changes the whole form, not only the people check, and is checked by one test, not a new scenario.
- **Message placement:** the "choose at least one person" message stays in the red box at the top; it is not moved next to the pills.

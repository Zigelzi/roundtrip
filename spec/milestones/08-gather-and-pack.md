# Milestone 08: Gather and pack

**Status:** done · **Branch:** feature/gather-and-pack (merged 2026-10-11)
<!-- Status: draft → red-teamed → in-progress → done -->

## Why
Flow: Gather items (`../user-flows.md` step 4) and Pack (step 6).

After testing, packing became the worst step (recorded in `../user-flows.md`, 2026-10-10). Two problems sit on the same page. An item is either packed or not, so while gathering we keep in our heads which things are already prepared on the bed. And a ticked item moves to one Packed section at the end of the page, so checking whether one person's bag is right means looking in two places. After this milestone each item can be marked as prepared (gathered on the bed, where we review it) before it goes into the bag, and each person's packed things stay in that person's section.

The milestone 02 design that moved packed items to the end was deliberate: it shrank what is left to pack as packing went on. That still matters, so packed items stay out of the way (collapsed), just inside their owner's section instead of a shared one.

## Scope
**In:**
1. Slice 1, per-person packed items: each person's (and the Family's) packed items stay in their own section, collapsed with a count, instead of one Packed section at the end of the page. A section with nothing packed has no packed part. Phone-checked on its own before slice 2 starts.
2. Slice 2, prepared: an item goes planned, then prepared, then packed.
3. Every item is prepared before it is packed; a planned item cannot be packed straight away.
4. Unpacking: tapping an item in the packed part takes it back to prepared.
5. Planned and prepared items share one list per person, in their usual order, told apart by a mark that differs in shape, not only colour. One tap on the row moves an item one step forward; there is no Gather / Pack switch (Q1).
6. Each tap names its step (prepare, pack, unpack, take back) and only moves an item that is in the state the step starts from: prepare only a planned item, pack only a prepared item, unpack only a packed item, take back only a prepared item. Otherwise nothing changes and the list comes back up to date. This stops a parent whose page is out of date from moving an item backwards or two steps at once (S8, S9).
7. Updating [`../domain-model.md`](../domain-model.md): prepared is built, and its note "until `prepared` exists, an unpacked item returns to `planned`" goes. The same note in the comment of `sql/schema/20260920120000_add_item_status.sql` is updated. Slice 3 adds prepared back to planned to both.
8. Milestone 02 scenarios this replaces: 02/S1 (one shared packed section), 02/S2 (a tap packs) and 02/S3 (unpacking returns to planned). Tests follow the business rules, so theirs are rewritten or removed; 02's other scenarios still hold. A note in `02-pack-items.md` points here.
9. Slice 3, taking a prepared item back to planned (Q6): a prepared row has a small back button at its end, beside the row's tap area. It shows an icon, not ✕, which on the trip page means "remove from the trip"; its spoken label names the step ("Back to planned: Swimsuit"). Planned and packed rows have no back button. Added after testing slice 2 (Q10, Q11).

**Out:**
1. Categories within a person's list (idea 6 in `../user-flows.md`). Its own milestone after this one.
2. Needs buying and bought (step 5, Review what is missing). The lifecycle keeps room for them, nothing is built.
3. Grouping by where things are in the house (idea 4).
4. Showing the prepared or packed state on the trip page (planning). The trip page stays as it is.
5. Remembering which collapsed sections were open between visits.
6. Moving several items at once (for example "all of Child 1's things are prepared").
7. (Taking a prepared item back to planned moved In as slice 3 on 2026-10-11, number kept free.)
8. A dedicated guard against an accidental double tap (Q7). Mostly covered by Scope 6: a second tap that lands before the page updates repeats "prepare" and changes nothing; a tap after the update packs, as intended.

## How we will know it worked
Same shape as a product hypothesis, sized for two users: the evidence is our own experience on the next real trip, not tracked numbers.

- **Hypothesis:** with prepared items marked and packed items kept under each person, gathering and packing take less mental effort.
- **Baseline (2026-10-10):** packing is the worst step; what is already on the bed lives in our heads; checking a bag means looking in two places.
- **Target, after the next trip:** we no longer keep what is prepared in our heads, and we check a person's bag by looking at their section alone. The step 4 and step 6 rows in `../user-flows.md` get updated with the result.
- **Guardrail:** every item now takes two taps instead of one. If that feels slower than the effort it saves, the strict "always prepare first" rule is the first thing to revisit.
- **Guardrail:** items marked as prepared that were not on the bed. Slice 3 lets a parent take them back (Q6). Watch for the opposite slip: aiming for the back button and packing the item instead.
- **Watch:** finding what is still to gather means scanning each person's list for unprepared marks (Q8). If that feels like the same head-work in a new place, a per-person "N still to gather" line is the first fix.

## Acceptance conditions (BDD)
<!-- Written so anyone in the family can read and challenge them: no code terms. How each one is tested lives under Test notes. -->

### S1: Packed items are shown under each member
- Given a trip where Child 1 has three items and one of them is packed
- And Child 2 has items but none of them is packed
- When the parent opens the packing page
- Then Child 1's section lists the two unpacked items
- And Child 1's section has a closed "Packed (1)" part holding the packed item
- And Child 2's section has no packed part
- And there is no shared Packed section at the end of the page

### S2: A fully packed member has all their items in their packed list
- Given a trip where all of Child 1's items are packed
- When the parent opens Child 1's packed part
- Then every item in Child 1's bag is listed there
- And Child 1's section says that Child 1 is packed

### S3: Parent prepares an item
- Given a trip where Child 1 has "Swimsuit", "Towel" and "Sun hat", all planned
- When the parent taps "Swimsuit"
- Then "Swimsuit" shows as prepared, with a mark shaped differently from a planned item's
- And the row says "Prepared" below the item name
- And Child 1's list keeps its order: "Swimsuit", "Towel", "Sun hat"
- And the row now offers to pack "Swimsuit" instead of preparing it
- And it is not counted as packed

### S4: Parent packs a prepared item
- Given Child 1's "Swimsuit" is prepared
- When the parent taps "Swimsuit"
- Then "Swimsuit" is packed
- And it is in Child 1's packed part
- And the progress counts it as packed

### S5: Parent unpacks an item
- Given Child 1's "Swimsuit" and "Towel" are packed
- When the parent opens Child 1's packed part and taps "Swimsuit"
- Then "Swimsuit" is prepared again, not back to planned
- And it is back in Child 1's list of planned and prepared items, in the same position as before it was packed
- And Child 1's packed part is still open, so "Towel" can be unpacked without reopening it
- And Child 1's packed count has gone down by one

### S6: Parent takes a prepared item back to planned
- Given Child 1 has "Swimsuit", "Towel" and "Sun hat", and "Swimsuit" is prepared
- And Child 1 has something packed, with the packed part open
- When the parent taps the back button on the "Swimsuit" row
- Then "Swimsuit" is planned again: the row no longer says "Prepared" and offers to prepare it
- And Child 1's list keeps its order: "Swimsuit", "Towel", "Sun hat"
- And Child 1's packed part is still open
- And only prepared items have a back button

### S7: Family items are packed into the Family section
- Given the family item "Sunscreen" is prepared
- When the parent taps "Sunscreen"
- Then "Sunscreen" is packed
- And it is in the Family section's packed part

### S8: Parent taps an item the other parent has just prepared
- Given Parent 1's page shows Child 1's "Swimsuit" as planned
- And Parent 2 has since prepared it on their phone
- When Parent 1 taps "Swimsuit"
- Then "Swimsuit" is prepared, not packed
- And Parent 1's list comes back up to date

### S9: Parent taps an item the other parent has already packed
- Given Parent 1's page shows Child 1's "Swimsuit" as planned
- And Parent 2 has since prepared and packed it
- When Parent 1 taps "Swimsuit"
- Then "Swimsuit" stays packed
- And Parent 1's list comes back up to date, with "Swimsuit" in Child 1's packed part

### S10: Parent takes back an item the other parent has already packed
- Given Parent 1's page shows Child 1's "Swimsuit" as prepared
- And Parent 2 has since packed it
- When Parent 1 taps the back button on "Swimsuit"
- Then "Swimsuit" stays packed
- And Parent 1's list comes back up to date, with "Swimsuit" in Child 1's packed part

### Test notes
- Slices: S1 and S2 are slice 1 (a tap still packs directly while slice 1 stands alone); S3 to S5 and S7 to S9 are slice 2; S6 and S10 are slice 3.
- S3: "offers to pack" is checked through the row's spoken label ("Prepare Swimsuit" becomes "Pack Swimsuit"), and the "Prepared" text is checked in the reply; the mark's shape is checked by eye on a phone.
- S5: the packed part staying open is a manual phone check, as in 02/S3. Automated: the reply holds the updated list, the packed part's new count and the progress, and the item is stored as prepared.
- S8 and S9 send the step the out-of-date page would send (prepare) for an item stored in a later state. The same rule covers a repeated unpack or pack; 02/S10 (an item the other parent removed) still holds unchanged.

**Slice 1, as tested** (written 2026-10-11, Opus, before implementation; `cmd/web/packed_per_member_test.go` and the rewritten tests in `pack_toggle_test.go`, `family_pack_test.go`, `pack_filter_test.go`):
- Each member section holds a wrapper `packed-member-<id>` that is always on the page. Inside it, only when the member has something packed, a closed `<details>` whose `<summary id="packed-summary-member-<id>">` says "Packed (n)" and whose `packed-list-member-<id>` lists the packed items. The shared `packed`, `packed-summary` and `packed-list` ids are gone.
- What a tap's reply swaps for the owner's packed part, besides the owner's `topack-member-<id>` list and `pack-progress`: when the part appears (first item packed) or goes (last item unpacked), the whole wrapper; otherwise only the summary and the list, never the `<details>` itself, which would close it (the 02/S3 hazard, now once per member). The server knows which case it is from the step and the new count. (Replaced in code review round 1: the page now says whether it shows the part, see the Decisions log.)
- 02/S2 and 02/S3 still hold their milestone 02 meaning in slice 1: a tap on a planned item packs it, and unpacking returns it to planned. Slice 2 changes both.
- Deleted: the test for one shared section's count (`TestPackedSectionShowsItsCount`); S1 and the per-member reply tests cover the count now.

**Slice 2, brief for whoever writes its tests** (red first, before implementation; follow the style of the slice 1 tests above):
- Steps and addresses: add `/trips/<t>/members/<m>/items/<i>/prepare` beside the existing `pack` and `unpack` ones (helpers `packItemURL`, `unpackItemURL` in `pack_toggle_test.go`; add `prepareItemURL` and a `prepareItem(t, db, id)` setter beside `packItem` in `pack_page_test.go`). The stored value is `prepared`.
- Scope 6 is the rule to pin down: prepare moves only a planned item, pack only a prepared item, unpack only a packed item (to prepared). Anything else stores nothing and answers 200 with the owner's list, packed part and progress as they really are (S8, S9). Test each "wrong state" pair at least once: pack on planned, prepare on prepared and on packed, unpack on prepared.
- S3: assert the row's spoken label (`aria-label`) changes from "Prepare Swimsuit" to "Pack Swimsuit", the "Prepared" text is in the row, the list order is unchanged (index of each name in the `topack-member-<id>` fragment), and the progress does not count it. Planned rows must not say "Prepared".
- S5: rewrite `TestUnpackItemInPlace` so the stored status is `prepared`, and the item reappears in its original position among the member's planned and prepared items (items are listed in the order they were added).
- Existing tests that pack a planned item through the address and must now start from prepared (or prepare first): `TestPackItemInPlace`, `TestPackRepliesNeverSwapAnExistingPackedPart`, `TestPackedPartAppearsWithTheFirstPackedItem`, `TestPackItemWithoutHTMXRedirects`, `TestPackItemRerendersTheItemsOwner` (all `pack_toggle_test.go`), `TestPackFamilyItem` (becomes S7), `TestPackingWhileNarrowedStaysNarrowed`, and `basics_test.go` around line 255. `TestPackItemThatIsGone`, `TestPackItemFromAnotherTrip` and `TestPackItemWithUnknownMemberSavesNothing` should also be run against the prepare address.
- Scope 7 is documentation, not tests: the `domain-model.md` note and the comment in `sql/schema/20260920120000_add_item_status.sql`.

**Slice 2, as tested** (written 2026-10-11, Opus, before implementation; `cmd/web/prepare_test.go` and the rewritten tests named in the brief):
- A row's tap target follows its state: a planned row posts to `prepare` and is labelled "Prepare <name>"; a prepared row posts to `pack`, is labelled "Pack <name>" and says "Prepared". The prepared mark survives a reload.
- S5 is tested with a planned and a prepared item added after the unpacked one, so an implementation that puts an unpacked item at the end of the list fails.
- Wrong-state taps are covered by S8, S9 and a table of the rest (pack on planned, unpack on prepared, unpack on planned). Each answers 200, stores nothing, and shows the item as it really is.
- S9 needs the whole packed part in the reply: the out-of-date page had no packed part for the member, so a count and list alone would have nowhere to go. Every wrong-state reply therefore sends the whole part (decided below, 2026-10-11).
- The removed-item, other-trip and unknown-member tests run against both `prepare` and `pack`; their prepare runs pass before implementation, since a missing route saves nothing either.

**Slice 3, as tested** (written 2026-10-11, Opus, before implementation; `cmd/web/unprepare_test.go`, plus one row each in the wrong-state table in `prepare_test.go`, the no-JavaScript test and the unknown-member test in `pack_toggle_test.go`):
- The step's address is `/trips/<t>/members/<m>/items/<i>/unprepare`, stored value `planned`. A prepared row holds two forms: the row's tap (pack) and the back button (unprepare, labelled "Back to planned: <name>", serialised like every other tap with `hx-sync`, and keeping the narrowing on a narrowed page). Planned rows and the packed part have no unprepare form.
- S6: the reply shows the row as planned again ("Prepare Swimsuit", no "Prepared", no back button), the list order is unchanged, the progress is unchanged, and the reply carries no `<details>`, so an open packed part stays open. That it stays open on screen is the usual manual phone check.
- S10 and the table row (take back a planned item) are the wrong-state cases: 200, nothing stored, the item shown as it really is; S10 needs the whole packed part, as in S9.
- A removed item and an item on another trip (stored prepared, so a missing state check would show) save nothing.
- The back button's icon and its size as a thumb target are checked by eye on a phone.

## Open questions
- None open.

## Decisions log

Kickoff (Human-PM, 2026-10-10), from a PM review of the packing step after testing (the problems are recorded in `../user-flows.md`, steps 4 and 6):

- This milestone covers two problems on the same page, built as two slices: first, each person's packed items stay in that person's section (collapsed) instead of one section at the end of the page; second, the `prepared` state from `../domain-model.md` (the pile on the bed), between planned and packed.
- Every item goes through prepared: planned → prepared → packed. Mistakes can be undone step by step back: packed → prepared → planned. `../domain-model.md` lists only packed → prepared today, so it gains prepared → planned.
- Explore how to show the three states, for example a Gather / Pack switch beside the person filter where each phase shows only what still needs attention. No reference designs chosen yet.
- Out: categories (idea 6 in `../user-flows.md`), its own milestone after this one. Needs-buying and bought (step 5, Review what is missing).

Open questions resolved (Human-PM, 2026-10-10):

- Q2 Progress line: counts only packed items. Packing is the goal the parents work towards, so the pile gets no number of its own.
- Q4 Existing trips: no change to existing items. Packed stays packed, planned stays planned, nothing starts on the pile. Nothing needs migrating either, since `prepared` is a new value in the status field items already have.
- Q5 Page length: collapsed by default is enough until categories arrive. A closed packed part adds one line per person, not the whole list.
- Two goals in one milestone, intentionally. `../README.md` treats a kickoff joining two goals with "and" as two milestones; this one is an exception because both slices change the same page and serve one aim, less mental effort while gathering and packing. Splitting them would mean redesigning the same sections twice. The cost is size: expect more than the rough ceiling of 8 scenarios once Q1 adds the scenarios for showing the three states.

Showing three states (Human-PM, 2026-10-10), after comparing four clickable sketches (Gather / Pack switch, one list where a tap moves forward, three parts per person, buttons on every row; https://claude.ai/artifact/Rxq2HYaVHKNotTEoonCz8t, private):

- Q1: one list per person, where a tap moves an item one step forward (option B). Planned items and items on the pile stay in place in that list, each with its own mark; packed items move to the person's collapsed packed part, as in slice 1. Chosen as the best compromise: the page stays as short as today and works like today's page, with no mode to switch. No consumer packing app found has a middle step, so there was no established pattern to copy.
- Q3: settled by Q1. One tap on the row moves forward; a tap on a row in the packed part takes it back to the pile, as unpacking works today. Taking an item off the pile is left open as Q6.
- Q6 Taking an item off the pile: later, not in this milestone. S6 and the prepared → planned half of the kickoff's undo rule move to Out, so `../domain-model.md` keeps only packed → prepared for now. Cost accepted: an item put on the pile by mistake stays marked as gathered until it is packed.
- Q7 Accidental double tap: deferred. Two quick taps on a planned item will pack it; revisit if it happens in real use.

Terms (Human-PM, 2026-10-11): the middle state is **prepared** and the action is **prepare**: items are prepared on the bed to be reviewed, then packed. Entries above written before this say "the pile" or "on the pile" for the same thing.

Red-team pass resolved (Human-PM, 2026-10-11). The red-team raised 11 findings (R1 to R11); R7 and R8 are open as Q8 and Q9.

- R1 Each tap names its step: accepted. A tap moves the item to the state it names, so a parent whose page is out of date repeats a step instead of skipping one. Human-PM accepted the rarer case (the other parent has moved the item two steps) as tolerable but wanted a guard if cheap; the main session proposed one with no extra cost: each step only moves an item from the state it starts from (Scope 6, S8, S9). Revisit if that rule gets in the way.
- R2 Replaced 02 scenarios: accepted. Tests move with the business rules, so 02/S1, S2 and S3 tests are rewritten or removed (Scope 8).
- R3 Unpacking keeps the packed part open and its count drops: accepted, added to S5.
- R4 A person with nothing packed has no packed part: accepted (Scope 1, S1). Cost: the section grows by one line when its first item is packed.
- R5 Prepared differs from planned in shape, not only colour, and the row's spoken label names its next step: accepted (Scope 5, S3).
- R6 The list keeps its order, and an unpacked item returns to its usual place: accepted (S3, S5).
- R9 S7 split to one step with one action: accepted. Preparing a family item works as in S3.
- R10 Scenarios tagged by slice: accepted (Test notes).
- R11 The domain model note and the status migration comment both get updated: accepted (Scope 7).
- R7 (now Q8), the finding: with planned and prepared items mixed in one list, answering "what have I not fetched yet?" means scanning every row's mark, and the progress line counts only packed items (Q2). The risk is moving the head-work from remembering the bed to scanning the list. Options: watch it on the next trip, or add a small per-person line such as "3 still to gather".

Remaining red-team questions (Human-PM, 2026-10-11):

- Q8 (R7) Seeing what is still to gather: watch it on the next trip rather than build a count now; easier to learn with the plain version to test on. Recorded in the success check.
- Q9 (R8) Second guardrail: added to the success check. A known problem, addressed later with Q6.
- R1's "only from the expected state" rule (Scope 6, S8, S9): confirmed.
- Human-PM wording edits to S1, S2, S3 and S5 (S3 gains the "Prepared" text under the item name); the main session fixed typos only.

Slice 2 tests (Human-PM, 2026-10-11):

- A tap that finds the item in another state answers with the member's whole packed part, not only its count and list. Accepted: simplest rule, and it is what brings an out-of-date page back (S9). Cost accepted: a stray repeated tap closes the packed part if it was open. (Removed in code review round 1, see below.)

Code review, round 1 (main session, 2026-10-11). 8 findings, 1 blocking. The blocking one: the server guessed from the database whether the page shows a member's packed part, so when the other parent packed that member's first item, the next tap from this page updated a count and a list that were not on it, and the packed part did not appear until a reload. Fixed: every tap form now says whether its page shows the part, and the reply sends the whole part unless the page shows it and the member still has something packed. This also removes the cost accepted above: a repeated tap no longer closes an open packed part. Tests: `TestTapFormsSayWhetherThePackedPartIsShown`, `TestTapFromAPageWithoutThePackedPartBringsItWhole`, `TestRepeatedTapKeepsThePackedPartOpen`. Declined, with reasons in `../allocation-log.md`: sending the open state from the browser (needs JavaScript; the page's field does the same job), the reply aimed at the wrong list for a hand-edited URL (from milestone 02, unreachable from the page), the quantity's dead spot (accepted at the phone check), re-rendering the whole trip per tap (two users, short lists), one URL helper for all steps (four named one-liners read better).

Slice 3, taking a prepared item back (Human-PM, 2026-10-11), after testing slice 2: not being able to take a prepared item back felt wrong in use, so Q6 comes back before the next trip rather than after it, as the guardrail planned.

- Q10 When the mistake is noticed: both, right after the tap and later while checking the bed. So the control stays on the row; a short-lived undo would miss the later case.
- Q6 How: a back button at the end of each prepared row (option A of five: back button, a tappable "Prepared" chip, an undo message, tapping the mark, long-press or swipe). It reads like the trip page's rows, works without JavaScript, and when needs-buying arrives, planned rows can take their side action in the same place. Cost accepted: a tap aimed at the back button that lands on the row packs the item, undone with unpack and back. No ✕ icon, since ✕ means remove on the trip page.
- Q11 Where: slice 3 of this milestone, not its own milestone. The page and its steps are fresh, and the work is small. Cost accepted: 08 grows to 10 scenarios and acceptance moves later.
- Domain note for the needs-buying milestone: back from prepared is unambiguous only while planned is the one way into prepared. Once bought also leads there, the single status field does not remember which one an item came from.

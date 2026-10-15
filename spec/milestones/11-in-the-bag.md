# Milestone 11: In the bag

**Status:** done · **Branch:** feature/in-the-bag (merged 2026-10-15)
<!-- Status: draft → red-teamed → in-progress → done -->

## Why
Flow: step 6, Pack (`../user-flows.md`), the step Human-PM rates weakest.

Each person's packed things sit at the very end of their list today, behind a closed "Packed (n)" line. To check whether something is already in the bag, a parent scrolls past everything still to pack and opens it. This milestone moves the packed part to right under the person's heading and shows the packed item names in it while it is closed, so "is it in the bag?" is answered at a glance, as on the design canvas ([`../mobile-redesign.md`](../mobile-redesign.md), R13, R15, R21).

## Scope
**In:**
1. Slice 1, the move: each person's packed part sits right under their heading and bar, above the things still to pack. It keeps the title "Packed (n)" and gets the canvas look: a green tick mark before it, the title in packed green, a chevron on the right that turns when the part is open. Opened, it says "Tap an item to take it out" under the title (R21). Taking an item out still returns it to prepared, in its usual place in the list (08/S5, R15).
2. Slice 2, the preview: while the part is closed, the names of the packed items show under the title, without quantities, cut short after two lines (R13). The preview follows every tap.

Each slice is phone-checked before the next.

**Out:**
1. Renaming the part "In the bag", as the canvas has it: the title stays "Packed (n)" (R21).
2. Making each person's heading sticky while scrolling their list (10/Out 3).
3. The "Next: Parent 2" button on the canvas: dropped (R14).
4. Changing what a tap does: packing, preparing and taking out work as in milestone 08.
5. The trip page: it shows no packed part.

## Acceptance conditions (BDD)

### S1: The packed part sits right under a person's heading
- Given Parent 1 has 3 items: Socks is packed, Hat and Towel are not
- When a parent opens the packing page
- Then Parent 1's section shows, in this order, their heading and bar, "Packed (1)", and then Hat and Towel
- And the packed part is closed
- And a person with nothing packed shows no packed part, as today (08/S1)

### S2: The packed part appears in its new place on the first tap
- Given Parent 1 has nothing packed yet, and Hat is prepared
- When a parent packs Hat
- Then "Packed (1)" appears right under Parent 1's heading, without reloading
- And Hat is no longer among the things still to pack

### S3: Opened, the packed part says how to take an item out
- Given Parent 1 has Socks packed
- When a parent opens Parent 1's packed part
- Then it reads "Tap an item to take it out" under the title
- And Socks is listed in it with a tick, its name struck through

### S4: Taking an item out from under the heading returns it to prepared
- Given Parent 1 has Socks and Hat packed, and their packed part is open
- When a parent taps Socks in the packed part
- Then Socks is back among the things to pack, marked prepared, in its usual place in the list
- And the packed part reads "Packed (1)" and stays open
- And when the parent also takes Hat out, the packed part goes away

### S5: The closed packed part shows what is in the bag
- Given Parent 1 has packed Socks (3 pieces) and Hat, in that order, and still has Towel to pack
- When a parent opens the packing page
- Then Parent 1's closed packed part reads "Packed (2)" with "Socks, Hat" under it
- And the preview shows no quantities
- And Towel is not in the preview

### S6: The preview follows every tap
- Given the trip from S5 is open on the packing page
- When a parent packs Towel
- Then the preview reads "Socks, Hat, Towel", without reloading
- And when the parent then takes Hat out, the preview reads "Socks, Towel"

### S7: A person who is all packed
- Given Child 1 has 2 items, both packed
- When a parent opens the packing page
- Then Child 1's section shows their heading and bar, "Packed (2)" with its preview, and then "All packed."

### S8: The packed part looks like the design canvas
- Given the packing page on a phone, with one person part way through packing and many items packed
- When a parent looks at it, then opens and closes their packed part
- Then the closed part shows a green tick mark, "Packed (n)" in green, and the names in grey, cut short with "…" after two lines
- And the chevron on the right points down when closed and up when open
- And opened, the names give way to "Tap an item to take it out"
- And the opened part's rows are tinted pale green, so it is clear where the packed items end and the items still to pack begin
- And packing an item while the part is open or closed does not make a quick second tap land on the wrong row (Q1)
- Verification: manual

## Test notes
- S1, S7: tests read the order of the section's parts by the offsets of their ids: the heading (`pack-head-member-N`), the packed part, then the list still to pack.
- S2, S4, S6: the htmx reply to a tap is what changes the page, so the tests read the reply. Its swap targets do not change in this milestone: the packed part's wrapper is always on the page, now under the heading, so swapping it in place puts it in its new position. S4 reruns 08/S5 from the new place: the existing tests (`TestUnpackItemInPlace` and the 08 tests around it) keep passing unchanged, and the part staying open is a manual check, as in 02/S3.
- S3: the hint is in the part's summary. That it replaces the preview only while the part is open is done in CSS, so a test checks only that both are present; the swap itself is S8's manual check.
- S5, S6: the preview lives in the part's summary, which every tap already swaps (`PackedSummary`), so no new swap target is needed. The test reads the summary alone, since the list inside the part does show "3 pcs".
- Wording that changes: the summary's "Show" and "Hide" go (Q2). No test reads either. "Packed (n)" and the rows' spoken "Unpack Socks" stay, so the tests reading them do not change.
- S7 reverses where "All packed." sits, so `TestFullyPackedMemberHasAllItemsInTheirPackedPart` (`packed_per_member_test.go`, 08) now looks for it after the packed part, not before (09/R1 rule).
- The tests are in `in_the_bag_test.go`, written before the code. Slice 1 turns S1 to S4 and S7 green; slice 2, S5 and S6.

## Design reference
From the canvas (artboards "Packing", "PackingPerson", "PackingDone" and their shared stylesheet), so implementing sessions do not need to open it.

- **Place:** right after the heading block, before the list still to pack, inside the section. A line above it (the heading's bottom edge), none doubled where the list below starts.
- **Summary:** at least 56px high, padding 12px top and bottom, 16px left, 4px right, items aligned to the top, 12px gaps. From the left: a 24px packed mark (green fill, white tick), then the text column, then a 48px-wide chevron column.
- **Title:** "Packed (n)", semibold, packed green.
- **Preview:** under the title, 14px on 20px lines, muted, the names joined with ", ", at most two lines, then cut with an ellipsis. Shown only while closed.
- **Hint:** "Tap an item to take it out", muted, in the preview's place, shown only while open.
- **Chevron:** pointing down, muted; turned 180 degrees while open.
- **Rows inside:** as today (52px, a 24px green tick mark, name struck through and muted, quantity muted). The canvas gives the quantity semibold; adopt that.
- **All packed:** the "All packed." note comes after the packed part, as on the "PackingPerson" artboard, so the part keeps its place under the heading in every state. ("PackingDone" has it the other way round; the two artboards disagree.)

## Open questions
None open.

## Decisions log
Light red-team of this draft, run inline by the main session (R19 in `../mobile-redesign.md`), 2026-10-15:
- **R1 Rows moving under the finger.** Raised as Q1.
- **R2 The 02 affordance.** Dropping "Show"/"Hide" undoes a fix from 02. Raised as Q2.
- **R3 The two "All packed" artboards disagree.** Decided: the note after the part, so the part never moves (S7, Design reference).
- **R4 The spoken label of a packed row.** The canvas says "Take out of the bag: Socks"; the app says "Unpack Socks". Kept as "Unpack", since R21 kept the "Packed" naming and a new label would rewrite the 02 and 08 tests for no change a parent sees.
- **R5 Preview order.** The names in the order the items are listed, which is the order they were added, the same as the rows inside the part. Most recently packed first was considered and left out: the list has no packing time to sort by, and the part's own rows would then disagree with its preview.
- **R6 A long preview and screen readers.** The two-line cut is visual only; a screen reader hears every name in the summary. Accepted: a parent using a screen reader gets the full answer, at the cost of a longer summary.
- **R7 Swap targets.** None added. The preview rides in the summary every tap already swaps, and the part's wrapper keeps its id in its new place.
- **R8 Slices share files.** Both touch `pack_page.templ`, so they run one after the other. Each is a few dozen lines, so both run inline in the main session ("Size before spawning", `../constitution.md`).

Open questions resolved (Human-PM, 2026-10-15), each as recommended:
- **Q1 Rows moving under the finger:** accepted as designed, judged at slice 1's phone check (S8). Fixed only if it causes mis-taps. Closing the part on every pack was rejected: the list would jump up by the part's whole height instead.
- **Q2 "Show" and "Hide":** dropped, following the canvas: the chevron, the preview and the opened part's hint say what the part is.

Decided at the phone check of both slices (Human-PM, 2026-10-15):
- **Where the open part ends.** Opened, the last packed row and the first row still to pack were split only by the line between any two rows, which was confusing. The packed rows are now tinted pale green, as prepared rows are tinted amber. A grey gap was the alternative, but grey gaps already separate one person from the next. Added as a tweak to S8, not a new scenario (workflow step 6e); a manual check.

Code review, round 1 (2026-10-15, Opus sub-agent, `/code-review medium`): no blocking findings, 7 non-blocking.
- **Fixed:** two tests that could not tell "All packed." inside the packed part from after it (`TestFullyPackedMemberHasAllItemsInTheirPackedPart` and S7's) now read the part and the list separately, and S7's test checks the preview too. The green tick mark is one shared piece (`packedMark`) instead of two copies. The preview breaks a very long name instead of cutting it at the edge. The chevron is centred on the title line. A comment rewrapped.
- **Declined:** drawing the down chevron by rotating the forward one; a separate icon reads more clearly.

Allocation (Human-PM, 2026-10-15): both slices inline in this main session, on Opus, tests written first.

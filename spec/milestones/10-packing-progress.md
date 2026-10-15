# Milestone 10: Packing progress

**Status:** done · **Branch:** feature/packing-progress (merged 2026-10-15)
<!-- Status: draft → red-teamed → in-progress → done -->

## Why
Flow: step 6, Pack (`../user-flows.md`), the step Human-PM rates weakest.

While packing, the only sign of progress today is one line of small text in the header, and it changes meaning when the page is narrowed to one person. A parent cannot see at a glance whose things are still far from done without opening each person's filter. This milestone shows progress three ways, as on the design canvas: a bar across the bottom of the header for the whole trip, a bar under each person's heading, and a "9/16" count in every filter pill, so the parent can pick the next person to pack for without scrolling ([`../mobile-redesign.md`](../mobile-redesign.md), R11, R12, R17, R22).

## Scope
**In:**
1. Slice 1, the trip bar: a thin bar along the bottom of the packing page's sticky header, green for packed and amber for prepared (R22, reversing 08/Q2 for the bars only), the rest grey. The "Packing · 38 of 75 packed" line stays beside it (R12). Both always count the whole trip, also when the page is narrowed to one person (Q1). Every trip has it, past trips too (R17).
2. Slice 2, each person's heading: the count changes from "16 items" to "9 of 16 packed", and a bar under it, green and amber like the trip bar (Q2). Both update with every tap.
3. Slice 3, the filter pills: each pill shows its name with "9/16" under it, Everyone and Family included (R11). A finished pill's count is green with a tick. The pills update with every tap.

Each slice is phone-checked before the next.

**Out:**
1. Animating the bars. The canvas has no animation, and htmx settling is off app-wide (09, Decisions log, Code review), so a width change cannot simply be a CSS transition. The bars jump to their new width (Q4).
2. "In the bag" under the heading, and its preview: milestone 11. The packed part keeps its place and title "Packed (n)".
3. Making each person's heading sticky while scrolling their list. The canvas does it, but it is not in the plan and needs its own look at how two sticky bars stack.
4. The "done" notes ("Parent 1 is packed. 3 of 5 packed on the trip.") keep their wording (Q3).
5. Progress on the trips list and the bottom packing button: milestones 14 and 12.
6. Counting prepared items in any text or pill: those stay packed-only ("61 of 75 packed", "9/16"), as 08/Q2 decided. Only the bars show prepared.

## Acceptance conditions (BDD)

### S1: The header shows how far the whole trip has come
- Given a trip with 4 items: 1 packed, 1 prepared and 2 planned
- When a parent opens its packing page
- Then the header reads "Packing · 1 of 4 packed"
- And a bar along the bottom of the header is one quarter green, one quarter amber, and grey for the rest

### S2: The trip bar follows every tap
- Given the trip from S1 is open on the packing page
- When a parent prepares one of the planned items
- Then the bar's amber part grows to half, without reloading
- And when the parent then packs a prepared item, its green part grows and its amber part shrinks
- And the header line counts the newly packed item

### S3: Narrowed to one person, the header still shows the whole trip
- Given a trip with 5 items, 1 of them packed, where Parent 1 has 2 items, neither packed
- When a parent narrows the packing page to Parent 1
- Then the header still reads "Packing · 1 of 5 packed"
- And the trip bar still shows the whole trip
- And Parent 1's own heading reads "0 of 2 packed"
- This replaces 02/S5, where the header counted only the person shown.

### S4: Each person's heading shows their own progress
- Given Parent 1 has 4 items: 1 packed, 1 prepared and 2 planned
- When a parent opens the packing page
- Then Parent 1's heading reads "Parent 1" and "1 of 4 packed"
- And a bar under the heading is one quarter green, one quarter amber, and grey for the rest
- And a person with nothing planned shows "Nothing planned." as today, with no count and no bar in their heading

### S5: A person's heading follows every tap
- Given Parent 1 has 2 planned items
- When a parent prepares and then packs one of them
- Then Parent 1's heading reads "1 of 2 packed", without reloading
- And their bar is half green
- And no other person's heading changes

### S6: Every filter pill shows packed out of total
- Given a trip where Parent 1 has 1 of 3 items packed, Family 0 of 1, and Parent 2 has no items
- When a parent opens the packing page
- Then the Everyone pill shows "1/4"
- And the Family pill shows "0/1", Parent 1 "1/3" and Parent 2 "0/0"
- And a screen reader hears each count as "1 of 3 packed", not as "1/3"
- And Parent 2's pill is dimmed, so it is easy to see who is not on the trip, but it can still be tapped
- And the pill being viewed is never dimmed, even when it shows "0/0"

### S7: A finished pill turns green, and every pill follows each tap
- Given Family has 1 item, prepared but not packed
- When a parent packs it, on the page showing everyone
- Then the Family pill shows "1/1" in green with a tick, without reloading
- And the Everyone pill counts it too
- And an empty person's "0/0" pill is not green

### S8: A past trip's packing page shows its progress
- Given a trip that came home last week, with 2 of its 3 items packed
- When a parent opens its packing page
- Then the header reads "Packing · 2 of 3 packed"
- And the trip bar, the person bars and the pill counts show as on any other trip

### S9: The bars look like the design canvas
- Given the packing page on a phone, part way through packing
- When a parent looks at it
- Then the trip bar is a thin line flush along the bottom of the header, edge to edge
- And each person's bar is a little thicker, with rounded ends, between their heading and their list
- And the filter pills are still two rows of three, each name with its count under it, the active pill's count in a pale blue
- Verification: manual

## Test notes
- Bars are decoration for the eye (hidden from screen readers, as on the canvas); the text beside them carries the numbers. Tests read a bar's two segments by their width. Pick item counts that give clean fractions (quarters, halves) so a test never depends on rounding.
- S2, S5, S7: the htmx reply to a tap is what changes the page, so the tests read the reply: the header's progress, the tapped person's heading and the filter row must each come back as an out-of-band swap, and no other person's heading may.
- S3 rewrites the tests that expected the narrowed header to count only the shown person (09/R1 rule): `TestPackPageNarrowedToOneMember` and `TestPackingWhileNarrowedStaysNarrowed` (`pack_filter_test.go`, 02/S5) and `TestPackPageNarrowedToFamily` (`family_pack_test.go`, 03). Each now reads the person's count from their heading, and the header's from the whole trip.
- S4 and S6 change visible wording: the packing heading's "16 items" (09/Q2) becomes "1 of 4 packed", and each pill's text gains its count. Tests that read either are listed here when found.
- S6: the pill's visible "1/3" is hidden from screen readers and a hidden "1 of 3 packed" is read instead, so the link is heard as "Parent 1, 1 of 3 packed".
- S8 uses `fixedNow` to place the trip in the past.
- The tests are in `packing_progress_test.go`, written before the code. Which slice turns each green: slice 1, S1 and S2; slice 2, S3, S4, S5 and the three rewritten tests (they read the heading); slice 3, S6, S7 and S8 (it checks all three places).
- Ids the tests rely on: `pack-trip-bar` (inside the header), `pack-head-member-N` (the person's heading and bar), `pack-filter` (the filter row). They are added to `packFragment`'s list of swap targets. Bar segments carry `data-bar="packed"` and `data-bar="prepared"` with their width as a percentage in an inline style.

## Design reference
From the canvas (artboards "Packing", "PackingPerson", "PackingDone" and their shared stylesheet), so implementing sessions do not need to open it.

- **Trip bar:** the last thing inside the sticky header, full width, 4px high, no rounding, line colour as its background. Two segments from the left: packed (green), then prepared (amber). Hidden from screen readers. It lives inside `#pack-progress`'s swap, or in its own swapped element.
- **Person bar:** under the heading row, inside the section's 16px gutter, 6px high, 3px radius, overflow hidden, line colour background, the same two segments. 10px above it, 12px below it. The heading row keeps the name left and the count right ("9 of 16 packed", muted, tabular figures), still inside the `<h2>`, which tests read as the heading.
- **Filter pill:** 56px high as today, the name (15px semibold) above the count (14px, medium, muted, tabular figures), stacked and centred. Active pill: count in `#D7E1EE` on the blue. Finished (packed equals total, total above zero): count in packed green, semibold, with a 14px tick (stroke 3) before it.
- Widths are percentages of the person's or the trip's total, packed first, prepared after it.

## Open questions
None open.

## Decisions log
Light red-team of this draft, run inline by the main session (R19 in `../mobile-redesign.md`), 2026-10-15:
- **R1 Header text and bar could disagree when narrowed.** Raised as Q1.
- **R2 Pill counts go stale after a tap.** The filter row is not part of any tap's reply today, so its counts would show the page's first load. It becomes an out-of-band swap on every tap (S7). It is a handful of links, so the extra bytes do not matter.
- **R3 The person's heading is not swapped today.** Its "16 items" never changed while packing (09/Q2); "1 of 4 packed" does. Each tap's reply now carries the tapped person's heading too, and only theirs (S5).
- **R4 "9/16" is read aloud badly.** A screen reader says "9 slash 16". The pill speaks "9 of 16 packed" instead (S6).
- **R5 A person with no items.** "0/0" must not count as finished, so it gets no green and no tick (S7); their heading shows no bar (S4).
- **R6 Sticky person headings on the canvas.** Not in the plan; left out (Out 3).
- **R7 Slices share files.** All three touch `pack_page.templ` and the same tap reply, so they run one after another, never in parallel.

Open questions resolved (Human-PM, 2026-10-15), each as recommended:
- **Q1 Header when narrowed:** the header line and the trip bar always count the whole trip; the person's own count is in their heading (S3, replacing 02/S5).
- **Q2 Prepared on the person bars:** yes, on both bars, so a bar means the same thing wherever it appears.
- **Q3 The "done" notes:** wording unchanged for now, judged at the slice 1 phone check.
- **Q4 Animation:** none in this milestone; the bars jump to their new width.

Decided at slice 3's phone check (Human-PM, 2026-10-15):
- **Empty pills dimmed:** a pill whose view holds nothing ("0/0") is dimmed, so the pills also show who is coming on the trip (S6). Added as a tweak to S6, not a new scenario, under step 6e of the workflow; the code review that follows covers it. The look is a manual check; the test checks only that the empty pill is styled apart from the others.

Code review, round 1 (2026-10-15, Opus sub-agent, `/code-review medium`): 1 blocking, 9 non-blocking.
- **Fixed, blocking:** with nothing left on the trip the trip bar was not rendered, so a tap whose reply found the trip emptied (the other parent removed its last item) left the old bar on screen. The bar's element is now always there. Settled by a test (`TestTripBarClearsWhenTheTripEmpties`), which failed before the fix.
- **Fixed:** the dimmed pill used half opacity, which took its text below a readable contrast (about 3:1, and 2:1 for the count); it is now dimmed by a lighter border and muted text. Not dimmed enough at Human-PM's check, so it became the canvas grey as fill, no visible border, muted text: it sinks into the page and keeps a 5.4:1 contrast. One "done" rule (`isDone`) for the pill's colour and its tick. One definition of "prepared" (the trip total sums each member's). A colour token for the active pill's count instead of a hex value. A stale comment. The Status line.
- **Added test:** a tap on a narrowed page keeps its pill current (`TestNarrowedTapKeepsItsPillCurrent`); it already passed, so this was a gap in coverage, not a bug.
- **Declined:** the grey line under the trip bar (the canvas has the same line under its header); a test for the finished count's green (a colour is the manual check S9, not a class name); the done note's repeated count (Q3, kept at the phone check).

Allocation (Human-PM, 2026-10-15): Option B. Each slice is implemented by a Sonnet sub-agent at high effort, against the tests the Opus main session wrote first; the main session reviews each slice before the phone check.

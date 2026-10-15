# Mobile redesign

The plan for a body of work spanning several milestones: a new look for every screen, then the packing features the redesign introduced. This file owns the order and the scope split between those milestones. Each milestone's own spec owns its status, scenarios and decisions, so this file links to them instead of repeating them.

## Sources

- Brief: [`docs/claude-design-prompt.md`](../docs/claude-design-prompt.md), the design brief with the usability rules and known pain points.
- Design canvas: https://claude.ai/artifact/MhRgkSTt11nyBjsGZL8pwD (private to Human-PM). Artboards for all five screens, the states, the design tokens and the colour options.
- Serves mainly step 6, Pack, and step 4, Gather the items, in [`user-flows.md`](user-flows.md). Packing is the step Human-PM finds weakest.

## Decisions made on the canvas (2026-10-11)

- Colour scheme: Fjord blue (`#1C3A5E`) as the action colour. Charcoal and Plum were compared; blue won because it already reads as "tappable" and "editing", which the other two do not.
- State colours keep fixed meanings: amber is prepared, green is packed, red is remove or error. The action colour may not resemble any of them.
- Trip page and basics: the whole row opens the edit row. A dotted underline on the quantity is the hint that rows are tappable.
- The add form keeps its − / + stepper, since added items are usually 1 to 4 pieces.
- "In the bag" sits right under each person's heading and previews the packed item names, without quantities.
- Taking an item out of the bag returns it to prepared, not planned.
- The packing button moves to a full-width bar at the bottom of the trip page, labelled "Start packing" or "Continue packing 38/75".
- New trip: departure date, then a nights slider (1 to 13 nights, see R2), then the computed return date shown for checking.
- A trip counts as past from the day after its return date.

## Milestones, in build order

The visual milestone comes first, so every feature after it is built once, in its final style.

| # | Milestone | Contains | Depends on | Spec |
|---|---|---|---|---|
| 09 | Visual refresh | Design tokens in Tailwind, the new look on all five screens, and the whole row opening the edit. No new route, handler or query. Slices below. | nothing | [09-visual-refresh.md](milestones/09-visual-refresh.md) |
| 10 | Packing progress | Packed count per person, shown three ways: the trip bar in the header (the "12 of 75 packed" text stays beside it, R12), a bar under each person's heading, and "9/16" in every filter pill (R11). The trip bar also shows prepared items in amber (R22). Also on past trips, so every trip's packing page shows its progress (R17). An animated bar must allow for htmx settling being off (09, Decisions log, Code review). | 09 (header, filter grid) | not started |
| 11 | In the bag | Each person's packed part moves from the end of their list to right under their heading, keeps the title "Packed (n)" with the hint "Tap an item to take it out" (R21), and previews the packed item names (R13). Unpacking already returns an item to prepared (08/S5); this milestone keeps that working from the new place (R15). | 09; after 10, since both change the space under the heading | not started |
| 12 | Packing button at the bottom | Slice 1: the trip page's packing button becomes a bottom bar, labelled with the trip's packed count, and "All packed" when everything is packed (R16). Slice 2: the bar hides while an input has focus or an edit row is open, so the keyboard keeps its space (R4). | 09; reuses the trip-wide count the packing page already shows, not 10 | not started |
| 13 | Nights slider | New trip asks for departure and nights instead of a return date, and shows the return date. | 09 (new trip screen) | not started |
| 14 | Trip list progress and past trips | Packing progress on upcoming trips in the trips list (R17), and past trips in their own group, open by default. | 09 (trips list); same count as 12 | not started |

Order: 10 and 11 come first after 09 because they work on the packing page, the weakest step. 12 to 14 depend only on 09, so their order among themselves is free and can change without touching the others.

When a milestone starts, its spec file replaces "not started" with a link, and from then on the spec's Status line is the only place its status is tracked.

## The line around milestone 09

Milestone 09 changes only how things look. The test for that: no new route, handler or database query is added, and tests change only where the visible wording changes, each such change listed in 09's spec (R1). Anything that fails this test belongs in a later milestone, not 09.

In 09: colours, type and spacing (tokens copied from the canvas into `cmd/web/tailwind.css`, which becomes their source from then on, R10); person lists running edge to edge on a phone; button, stepper and edit row styles; back chevrons; the Basics button on the trips page, in its new place (R8); the 3×2 packing filter grid; the slimmer sticky header; hiding "1 pc" on the packing page; the existing back-to-planned button on prepared rows, restyled (R6); the edit row in the blue action tint of the canvas's fifth frame, not amber (R5); the dotted underline and whole-row tap on the trip page and basics, where ✕ keeps its own separate 48px zone and, while a row is open, every other row is locked as a whole (R7).

Slices, each phone-checked before the next (R9):

1. Tokens and the trips page: the tokens in Tailwind, the shared styles (buttons, inputs, back chevron), and the trips page with its Basics button.
2. Trip page: edge-to-edge person lists, the slimmer sticky header, whole-row tap with the dotted underline, the blue edit row, the stepper, ✕ in its own zone.
3. Packing page: header, the 3×2 filter grid, the packing rows, the restyled back-to-planned button, hiding "1 pc".
4. Basics page: the trip page's patterns applied to basics.
5. New trip page: the form restyled as it works today (the slider is 13).

## Red-team pass (2026-10-11)

Findings R1 to R20, raised in a main-session red-team of this plan. Decided so far (Human-PM, 2026-10-14):

- R1 Wording changes in 09: accepted. Tests that check visible wording change with it, and 09's spec lists each one. Human-PM's concern: a test suite that breaks on a single UI change is too fragile. Improving how tests check the screen is a workflow question outside this plan, to look at separately.
- R2 Trip length: the slider runs from 1 to 13 nights, which is the existing limit of 14 days. Day trips are not supported, so a trip is at least 2 days. Storage stays in days (nights = days − 1).
- R4 Hiding the bar while typing: merged into 12 as slice 2, since 12 without it may cost list space while typing. The milestones after it moved up by one (nights slider 13, trip list 14); none had started.
- R5 Edit row colour: the blue action tint from the canvas's fifth frame, since amber now means prepared.
- R6 "Undo zone": the existing back-to-planned button on prepared rows (08 slice 3), restyled. Same behaviour.
- R7 Whole-row tap next to ✕: ✕ keeps its own 48px zone, and while a row is open every other row is locked as a whole.

- R8 Basics button: only on the trips page. Same button, new layout.
- R9 Slices: accepted, listed under "The line around milestone 09".
- R10 Tokens: copied into `cmd/web/tailwind.css` in 09 slice 1, so implementing sessions do not need to read the canvas.
- R11 Filter pills: every pill, Everyone and Family included, shows "9/16" (packed/total). Everyone is always shown, so a two-parent trip has three pills, which fill one row of the three-column grid.
- R12 Header: both the progress bar and the "n of m packed" text, as in the canvas.
- R13 In the bag: replaces the packed part at the end of each person's list and keeps the title "Packed (n)".
- R14 "Next: person" button: dropped.
- R15 Unpacking to prepared: already built in 08, kept.
- R16 Bottom bar when everything is packed: reads "All packed" and still opens the packing page. The count updates when an item is added or removed on the trip page, and the page has bottom padding so the bar never covers the last add form.
- R17 Progress: on the trips list, upcoming trips only. Every trip's packing page shows it, past ones too, through 10's header bar. The past trips group is open by default.
- R20 Dependencies: mapped in the table. 12 needs only the trip-wide count, which exists today, so it does not wait for 10.

- R3 Past trips: a trip counts as past from the day after its return date. Today a trip counts as past once its departure has gone by, so a trip in progress shows as past; 14 changes that and gets a scenario for it.
- R18 Baseline ratings before 09, and a real trip before 11: skipped.
- R19 Red-team for the small milestones: a light red-team pass for each milestone, run inline by the main session against that milestone's spec (as this pass was), instead of a spawned agent.
- R21 The packed part's title (raised from the canvas): stays "Packed (n)", with the canvas's hint "Tap an item to take it out".
- R22 Prepared in the progress bar (raised from the canvas): the trip bar shows prepared items as an amber segment beside the green packed one. This reverses 08/Q2 for the bar only; the counts in text and pills ("61 of 75 packed", "9/16") stay packed-only. 10's spec records the reversal.

All findings decided (Human-PM, 2026-10-14).

## Out of scope for this plan

- Item categories (outdoor, hygiene, clothes). The redesign leaves room for them, but they are a separate piece of work.
- Ideas raised while designing that are not planned here belong in the ideas list in [`product-notes.md`](product-notes.md).

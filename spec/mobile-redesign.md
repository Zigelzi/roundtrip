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
- New trip: departure date, then a nights slider (1 to 14 nights), then the computed return date shown for checking.
- A trip counts as past from the day after its return date.

## Milestones, in build order

The visual milestone comes first, so every feature after it is built once, in its final style.

| # | Milestone | Contains | Spec |
|---|---|---|---|
| 09 | Visual refresh | Design tokens in Tailwind, the new look on all five screens, and the whole row opening the edit. No new route, handler or query. | not started |
| 10 | Packing progress | Packed count per person, shown three ways: the trip bar in the header, a bar under each person's heading, the count in each filter pill. | not started |
| 11 | In the bag | "In the bag" under each person's heading with a name preview; taking an item out returns it to prepared; a "Next: person" button at the end of a one-person view. | not started |
| 12 | Packing button at the bottom | The trip page's packing button becomes a bottom bar, with the progress from 10 in its label. | not started |
| 13 | Hide the packing button while typing | The bottom bar hides while an input has focus or an edit row is open, so the keyboard keeps its space. | not started |
| 14 | Nights slider | New trip asks for departure and nights instead of a return date, and shows the return date. | not started |
| 15 | Trip list progress and past trips | Packing progress on the trips list, and past trips in their own group. | not started |

When a milestone starts, its spec file replaces "not started" with a link, and from then on the spec's Status line is the only place its status is tracked.

## The line around milestone 09

Milestone 09 changes only how things look. The test for that: every existing test passes without changes, and no new route, handler or database query is added. Anything that fails this test belongs in a later milestone, not 09.

In 09: colours, type and spacing; person lists running edge to edge on a phone; button, stepper and edit row styles; back chevrons; the Basics button in the top bar; the 3×2 packing filter grid; the slimmer sticky header; hiding "1 pc" on the packing page; the undo zone on prepared rows; the dotted underline and whole-row tap on the trip page and basics.

## Out of scope for this plan

- Item categories (outdoor, hygiene, clothes). The redesign leaves room for them, but they are a separate piece of work.
- Ideas raised while designing that are not planned here belong in the ideas list in [`product-notes.md`](product-notes.md).

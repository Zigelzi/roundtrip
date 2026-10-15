# Milestone 09: Visual refresh

**Status:** done · **Branch:** feature/visual-refresh (merged 2026-10-15)
<!-- Status: draft → red-teamed → in-progress → done -->

## Why
Flow: every step that has a screen, mainly Gather the items (`../user-flows.md` step 4) and Pack (step 6).

The screens were styled one milestone at a time, so colours carry no fixed meaning (the edit row is amber, prepared items are blue, − is red like remove) and each page has its own look. The redesign on the canvas gives every screen one look with fixed colour meanings: blue for actions and editing, amber for prepared, green for packed, red for remove and errors. It also makes the whole row the tap that opens an edit, where today only the small quantity is. This milestone builds that look on all five screens before any new feature, so milestones 10 to 14 are built once, in their final style ([`../mobile-redesign.md`](../mobile-redesign.md)).

## Scope
**In** (the plan's "line around milestone 09" holds: no new route, handler or query):
1. Slice 1, tokens and the trips page: the canvas's colours, radii and sizes as Tailwind tokens in `cmd/web/tailwind.css`, which is their source from then on (R10); the shared styles (buttons, inputs, stepper, back chevron); the trips page with Basics beside the title (R8) and the empty state's link to basics.
2. Slice 2, trip page: the header with the back chevron, pill-shaped jump links, person lists edge to edge with 8px between them, whole-row tap with the dotted underline, ✕ in its own 48px zone (R7), every other row locked as a whole while one is open, the edit row in the blue action tint (R5), the joined − / + stepper, the add form restyled.
3. Slice 3, packing page: the header with the back chevron, the 3×2 filter grid, the new marks (empty square, amber circle, green filled square), the amber prepared row with its back button at the row's end, restyled (R6), the packed part restyled where it is today, and "1 pc" hidden.
4. Slice 4, basics page: the trip page's patterns applied to basics, and the edit row's fields labelled Name, Per day and Extra, each number with a stepper.
5. Slice 5, new trip page: the form restyled as it works today (return date and days stay; the slider is milestone 13), the people as toggles with a tick box, errors in one alert box.

Each slice is phone-checked before the next (R9).

**Out:**
1. Progress bars and counts in the filter pills: milestone 10.
2. "In the bag" under each person's heading, with its preview: milestone 11. The packed part stays at the end of each list, titled "Packed (n)".
3. The packing button as a bottom bar: milestone 12. Until then "Start packing" stays in the trip page header, restyled.
4. The nights slider: milestone 13. Upcoming and Past groups on the trips page: milestone 14.
5. The "Next: person" button on the packing page (R14, dropped).
6. Locking the add forms, the back chevron and the jump links while a row is open. The canvas dims them, but today they stay usable on purpose (04, see `quantity_test.go`), and changing that is behaviour, not look. Only the item rows lock.
7. A dark theme, and any layout designed for desktop beyond today's width cap.
8. Any change to what a control does, other than the larger tap area of a row.

## Acceptance conditions (BDD)
<!-- Most of this milestone is how things look on a phone, which a Go test cannot see. Those scenarios are manual; the part a test can observe is still tested. -->

### S1: Every screen has the new look
- Given any of the five screens
- When a parent opens it on a phone
- Then its colours, type and spacing match the design canvas
- And blue appears only on things to tap or being edited
- And amber appears only on prepared items, green only on packed ones, red only on remove and errors
- Verification: manual, slice by slice. That the stylesheet is built with the new colours is tested.

### S2: Basics is beside the title on the trips page
- Given the trips page
- When a parent opens it
- Then the top bar shows "Roundtrip" with a Basics button beside it
- And New trip is the one solid, full-width button below them

### S3: A parent with no trips is pointed to the basics first
- Given there are no trips
- When a parent opens the trips page
- Then it says "No trips yet" and that a new trip starts with everyone's basics
- And it has a "Check basics" link to the basics page

### S4: Every inner page has a back button at the top
- Given the trip page, the packing page, the basics page or the new trip page
- When a parent opens it
- Then the top bar starts with a back arrow
- And on the packing page it goes back to the trip, on the others to the trips page
- And on the trip, packing and basics pages it stays in view while scrolling, inside the sticky header

### S5: Tapping anywhere on a row opens it for editing
- Given Parent 1 has Socks, 4 pcs, on the trip page
- When a parent taps the word "Socks"
- Then the Socks row opens for editing, as tapping the quantity does today
- And the quantity has a dotted underline, the hint that rows can be tapped
- And the ✕ at the end of the row is a separate button that removes Socks, never opens it
- And the basics page works the same way

### S6: While a row is open, the other rows rest
- Given the Socks row is open for editing
- When a parent taps another row, its name, its quantity or its ✕
- Then nothing happens
- And every other row is greyed out as a whole
- And the add forms still work, as today

### S7: The open row is blue and uses a stepper
- Given the Socks row is open for editing
- When a parent looks at it
- Then the row is tinted blue, not amber
- And the quantity sits between − and + in one stepper, neither of them red or green
- And Save is a solid blue button
- And Cancel is a plain blue text button
- And on the basics page the open row labels its fields Name, Per day and Extra, each number with its own stepper
- Verification: manual

### S8: The packing page header shows the trip and its progress
- Given a trip "Parainen" with 3 items, none packed
- When a parent opens its packing page and scrolls down
- Then the header shows a back arrow, "Parainen" as the title, and "Packing · 0 of 3 packed" under it
- And the header stays in view
- And after packing an item the line reads "Packing · 1 of 3 packed" without reloading

### S9: The packing filter is a grid of three columns
- Given a trip with Parent 1, Parent 2, Child 1, Child 2 and Family items
- When a parent opens the packing page on a phone
- Then the six filters sit in two rows of three: Everyone, Family, Parent 1, then Parent 2, Child 1, Child 2
- And the one being viewed is solid blue
- Verification: manual for the layout; the order is already tested (03/S9)

### S10: Packing rows show their state by shape and colour
- Given Parent 1 has one planned, one prepared and one packed item
- When a parent opens the packing page
- Then the planned item has an empty square
- And the prepared item has an amber circle with a dot, an amber row and the word "Prepared"
- And the prepared row's back button is at its very end, the full height of the row
- And the packed item, in the packed part, has a green square with a tick
- Verification: manual; what each tap does is already tested (08)

### S11: The packing page shows a quantity only when there is more than one
- Given Parent 1 has Underwear, 4 pcs, and Toothbrush, 1 pc
- When a parent opens the packing page
- Then Underwear shows "4 pcs"
- And Toothbrush shows no quantity
- And the trip page still shows Toothbrush as "1 pc"

### S12: The new trip form has the new look
- Given the new trip page
- When a parent fills it in and makes a mistake
- Then each field has its label above it, and each person is a toggle with a tick box, blue when going
- And the problems show together in one red box at the top, as today
- And Create trip is the one solid, full-width button
- Verification: manual; the form's behaviour is unchanged and already tested

## Test notes
- Tests check what a parent can see or a screen reader can hear, not class names, except the one stylesheet check that already exists (`handlers_test.go`), which moves from `text-3xl` (no longer on the trips page) to a class the new trips page uses and to one new colour token.
- R1, the tests that change because visible wording changes: `pack_page_test.go` expects "1 pc" on the packing page (S11 hides it). Slice 4 adds a second: `edit_basics_test.go` counts open rows by their "· edit basic" hint, which goes (R5), so it counts them by another mark. Any other found while implementing is added here before it is changed.
- Row labels for screen readers stay as they are ("Change quantity of Socks", "Remove Socks", "Pack Socks", "Back to planned: Socks"), so the many tests that find buttons by them keep working. The tap now covers the whole row, but its label still says what it does.
- S5: tested by checking that the item's name is inside the same button as its quantity, and that ✕ is outside it. S6's disabled buttons are already tested (04, 06); the greying is manual.
- S4: tested by finding a link with a spoken label ("Back to trips", "Back to trip") and the right address, inside the header where the page has one.

## Design reference
Taken from the canvas (https://claude.ai/artifact/MhRgkSTt11nyBjsGZL8pwD, the artboards and their shared stylesheet) so implementing sessions do not need to open it.

**Tokens:** the colours and radii live in `cmd/web/tailwind.css` (`@theme`), copied there from the canvas in slice 1, and are used as `bg-canvas`, `text-muted`, `rounded-control` and so on. That file says what each colour means.

Radii: 7px tick boxes, 8px small chips, 10px controls (buttons, inputs, pills), 12px section cards on a wide screen, 0 on a phone. Jump links take the 8px chip radius (Q4). Type: the system font stack; title 24/32 bold, page title in a bar 18/24 bold, section heading 18/24 bold, body 16/24, small 14/20; numbers use tabular figures. Sizes: tap target at least 48px, plan row 56px, packing row 64px, packed row 52px, filter pill 56px, screen gutter 16px, 8px between sections, content at most 672px wide (today's `max-w-2xl`).

**Shared pieces:**
- Buttons, 48px high, 10px radius, semibold: primary (solid primary, white text; one per screen), secondary (white, line-strong border), quiet (no border, primary text). Icon buttons are 48px squares with no border.
- Inputs: 48px, line-strong border, 10px radius, 16px text (stops iOS zooming).
- Stepper: one bordered box holding −, the number and +, each part 48px wide, dividers between them, ink-coloured symbols. The number stays a real input so the form works without JavaScript. Steppers never go below the field's lowest value: 1 for a trip quantity, 0 for a basic's Per day and Extra.
- Icons are inline SVG line icons (24 viewbox, 2px stroke, round ends): back chevron, ✕, +, −, arrow, the undo arrow, the tick. No icon library.
- Top bar: white, line under it. Back chevron as a 48px icon button at the left, then the title (truncated to one line) with a meta line under it.
- Sections: white, edge to edge on a phone, 8px apart on the canvas colour; rounded 12px cards from the `sm` breakpoint. Section heading row: name left, count right ("16 items", muted). The count stays inside the `<h2>`, which tests read as the heading.

**Trips page:** top bar with "Roundtrip" (24px bold) and Basics as a secondary button on the right; New trip as a full-width primary button with a + icon, inside the bar. The trips sit in one white section, rows at least 76px, divided by lines, each with the destination (17px semibold), the dates and days line, and a chevron at the right. Empty: a white card with a bag outline icon, "No trips yet", the line "A new trip starts with everyone's basics already on the list, so check those first." and a quiet "Check basics" link.

**Trip page:** sticky top bar: back chevron ("Back to trips"), destination and dates, and Start packing as a primary button at the right until milestone 12. Under it the jump links as lightly rounded buttons (8px, Q4), one line, equal width as today (09/Q1). The header should be no taller than today's, about 117px, so `anchorClass` keeps landing jumps below it; measure and adjust the two together if not. Rows: the name and the quantity are one button (the row's tap area, pressed state primary-soft), quantity in primary with a dotted underline (muted when it is "1 pc"); ✕ in danger as its own 48px icon button at the end. While a row is open, every other row is at about 40% opacity with its buttons disabled. Edit row: primary-soft fill, 2px primary line above and below, no rounded box; first line the item's name in bold (no "· change quantity", R5) and Cancel as a quiet button at the right; second line the stepper and Save (primary). Add form under the list, behind a line: the name field full width ("Add an item for Parent 1"), then the stepper and Add as a secondary button. The view transition styles in `tailwind.css` paint the section card on the transition group; they follow the new card colour and radius.

**Packing page:** sticky top bar: back chevron ("Back to trip"), destination as the title, "Packing · 38 of 75 packed" as the meta line (the count part is the swapped progress). The done messages stay in the header as today, restyled as packed-soft notes. Filter: a white band under the header, three equal columns, 8px gaps, pills 56px high; the active one solid primary with white text. Person sections: heading row as on the trip page. Rows 64px, divided by lines: mark (28px), name, and the quantity in bold at the end of the tap area, empty when it is 1. Marks: planned an empty square (2px faint border, 7px radius); prepared a circle with a 12px dot, both prepared colour, the row prepared-soft, "Prepared" under the name in 14px semibold prepared; packed a filled packed square with a white tick (24px in the packed part). Back button on a prepared row: 56px wide, full row height, the undo arrow in prepared colour, prepared-line on its left; it sits after the quantity, at the row's very end. The packed part keeps its place and title "Packed (n)"; its rows are 52px with the small packed mark, the name struck through in muted. "All packed." becomes a packed-soft note with a tick.

**Basics page:** as the trip page. Top bar: back chevron, "Basics", and "Every new trip starts with these." kept as the meta line. Rows: name and rule ("1 / day + 1 extra") as one button with the dotted underline. Edit row: primary-soft as on the trip page, "Name" label and field, then "Per day" and "Extra" labels, each over a stepper, then Cancel (quiet) and Save (primary) at the right. The add form keeps its fields, restyled.

**New trip page:** top bar with back chevron and "New trip". The form in one white panel, 16px padding, 20px between fields, labels 14px semibold above each field. Departure and Return side by side as today, Days under them. Who is going: two columns of 56px toggles; going is a 2px primary border, primary-soft fill and a filled primary tick box; not going is a line-strong border and muted text. Errors: one danger-soft box with a list. Create trip: full-width primary.

## Open questions
None open.

## Decisions log
Light red-team of this draft, run inline by the main session (R19 in `../mobile-redesign.md`), 2026-10-14. Findings and how the draft handles them:
- **R1 Start packing has nowhere to go in 09.** The canvas moves it to the bottom bar, which is milestone 12. Kept in the trip page header, restyled (Out 3).
- **R2 The canvas locks more than rows.** It dims the add forms, back chevron and jump links during an edit; 04 keeps the add forms usable on purpose. Rows only (Out 6).
- **R3 A larger tap area means more accidental opens while scrolling.** Opening a row is harmless (Cancel, or any other row is locked until then), so accepted; watched at slice 2's phone check.
- **R4 Tests that find buttons by spoken label.** Labels unchanged, so only S11 changes a test's wording (Test notes).
- **R5 The canvas drops "· change quantity" from the edit row.** It was added after 04's phone check because the open row was not obvious. Kept for slice 2 so its phone check could decide. Decided at that check (Human-PM, 2026-10-14): dropped, the blue fill and lines are enough. The basics edit row follows in slice 4: its "· edit basic" goes too, and the test that counts it (`edit_basics_test.go`, `TestBasicEditFormShowsZeroesEmpty`) changes with it (R1).
- **R6 Basics steppers can go below zero.** Today's basics fields have no − / +; the new steppers stop at 0. The server's checks are unchanged.
- **R7 The header could grow.** The back link moves into the header and the jump links become pills; `anchorClass` must match its height (Design reference, trip page).

Open questions resolved (Human-PM, 2026-10-14):
- **Q1 Jump links:** the canvas shows a count in each jump link ("Parent 1 16"), in a row that runs off the right edge. With five people the counts do not fit on one line at phone width, and the header is not part of what changes when an item is added, so the count there would go stale until the page reloads. Decided: pill-shaped links without counts, equal width on one line as today; the count stays in each section's heading, where it updates.

Decided at slice 3's phone check (Human-PM, 2026-10-15):
- **Q2 Packing section heading count:** "heading row as on the trip page" left open what to count. Decided: the member's total items ("16 items"), which packing never changes, so it is not swapped and cannot go stale. A "left to pack" count belongs with milestone 10's progress.

Decided at slice 4's phone check (Human-PM, 2026-10-15):
- **Q3 The basic's name in its open row:** the first build kept the name as a bold title above the Name field, as the trip page's open row has. Decided: shown only in the field, since twice is redundant. Tests find an open basic row by its Name field's spoken label ("Name of Underwear"), which names the stored basic even after a refused rename.

Decided at slice 5's phone check (Human-PM, 2026-10-15):
- **Q4 Jump link shape:** the canvas's full-round pills read differently from every other button. Decided: lightly rounded like the packing filter's links, but smaller (44px high, 8px radius), on the trip and basics pages.

Code review (2026-10-15), for later milestones:
- **htmx settling is off app-wide** (`attributesToSettle: []` in `cmd/web/view/layout.templ`), because it drew each swapped row in its old class for a frame (a flicker). So a CSS transition on the class or style of an element htmx swaps by id no longer animates: milestone 10's progress bars, if they animate their width inside `#pack-progress`, need another way (a transition on an element htmx does not replace, or settling turned back on for that one swap).

Allocation (Human-PM, 2026-10-14): the whole milestone runs in one Opus main session, tests first and then each slice. The work is mostly markup with little logic, and one context that holds the canvas reading should follow it more accurately than handing slices on.

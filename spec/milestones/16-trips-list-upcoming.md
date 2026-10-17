# Milestone 16: Upcoming heading and progress only once packing starts

**Status:** done · **Branch:** feature/trips-list-upcoming
<!-- Status: draft → red-teamed → in-progress → done -->

## Why
Flow: step 6, Pack (`../user-flows.md`), seen from the trips list, the first screen of the app.

Milestone 14 split the list into upcoming and past trips, but only the past group has a heading. The upcoming trips are an unlabelled run, and when every trip is past the list shows the past group alone, which reads as if the upcoming trips had gone missing. Each upcoming trip with items also shows a progress bar and "0 of 40 packed" before any packing has begun, which is noise on a trip that is only planned. This milestone labels the upcoming trips, says so when there are none, and shows progress only on a trip where packing has started.

## Scope
**In:**
Two slices, each phone-checked before the next.

1. Upcoming heading: whenever there is at least one trip, past or upcoming, the upcoming trips sit under a heading "Upcoming trips". The list is always visible, not collapsible. When every trip is past, the heading is followed by a short line saying there are no upcoming trips. This reverses 14/Q3 ("only past trips: no upcoming section"). S1 to S4.
2. Progress once started: a trip's row shows its bar and count only when at least one of its items is prepared or packed. A trip with items but none prepared or packed shows the destination and dates only. S5 to S8.

**Out:**
1. Showing who is going on each trip's row. That changes what the list loads, so it is its own milestone.
2. Any change to the past trips group, the trip page, the packing page or the new trip form.
3. Remembering anything between visits.
4. A button or other action in the "no upcoming trips" line: New trip is already at the top of the page.

## Acceptance conditions (BDD)

### S1: Upcoming trips have a heading
- Given today is 19.9.2026
- And trips to Parainen (20.–22.9.2026) and Levi (1.–5.12.2026)
- And a past trip to Kuopio (10.–11.6.2026)
- When a parent opens the trips list
- Then a heading "Upcoming trips" is shown above Parainen and Levi
- And Parainen and Levi are visible without tapping anything

### S2: The heading shows without past trips
- Given only upcoming trips
- When a parent opens the trips list
- Then the heading "Upcoming trips" is shown above them
- And there is no past trips group

### S3: Only past trips
- Given only past trips
- When a parent opens the trips list
- Then the heading "Upcoming trips" is shown
- And under it a line says there are no upcoming trips
- And below that the past trips group is shown, closed
- And the list does not say "No trips yet"

### S4: No trips at all
- Given no trips
- When a parent opens the trips list
- Then the list says "No trips yet", as before
- And there is no "Upcoming trips" heading
- And no line saying there are no upcoming trips

### S5: A planned trip shows no progress
- Given an upcoming trip to Levi with 40 items, none prepared and none packed
- When a parent opens the trips list
- Then Levi's row shows the destination and dates
- And no bar
- And no count

### S6: Preparing an item starts the progress
- Given an upcoming trip to Levi with 40 items, 6 prepared and none packed
- When a parent opens the trips list
- Then Levi's row reads "0 of 40 packed"
- And its bar is amber for the 6 prepared, the rest grey

### S7: Packing an item starts the progress
- Given an upcoming trip to Levi with 40 items, 18 packed and none prepared
- When a parent opens the trips list
- Then Levi's row reads "18 of 40 packed"
- And its bar is green for the 18 packed, the rest grey

### S8: Taking everything back hides the progress again
- Given an upcoming trip to Levi with 40 items, 6 prepared and 18 packed
- And the parent then moves every item back to planned
- When the parent opens the trips list
- Then Levi's row shows the destination and dates
- And no bar
- And no count

## Test notes
- The tests are in `trip_list_test.go`, beside milestone 14's, written before the code.
- Every test fixes today with `fixedNow`, as 14 does.
- S1 to S4: the heading and the line are read from the page's text. "Visible without tapping" is the upcoming trips not being inside a `<details>`. S3 and S4 also check the other state's wording is absent.
- S5 to S8: a bar is present when the row holds the bar's segments, as in 14/S7, so absence is the row having none and no "of 40 packed" text. S6 and S7 read the segment widths as 14/S7 does. S8 sets the state through the same packing requests a parent uses, so the test fails if "started" is stored and not counted from the items.
- 14/S8 (all packed reads "All packed") and 14/S10 (past trips show no progress) hold unchanged.

Older tests whose behaviour changes (09/R1). Each keeps its intent:
- 14/S5 (only past trips): its last line, "there is no empty upcoming section", is reversed by S3. The test now expects the heading and the line.
- 14/S7 and 14/S9: still hold. Any test that gave a trip items but none prepared or packed and expected a bar fails on S5 and is changed to start packing first. Running the suite shows which.

## Design reference
Built from the tokens in `cmd/web/tailwind.css` and judged at the phone check (Q1).

- **Heading:** "Upcoming trips" as a small heading above the white list, in the muted text size of the dates line, semibold, like the group label on the canvas's first version of the Trips board. It is a real heading for screen readers.
- **Empty line:** inside a white section under the heading, "No upcoming trips." in the muted body text, with the row padding of the lists around it. No icon, no button.
- **Row without progress:** the row of today with its third and fourth lines left out, so it is two lines, destination and dates.

## Open questions
- **Q1 Heading look.** The past group's heading is a tappable row inside its white section, while "Upcoming trips" is a label above its list. They could look alike (same size and weight, inside the section) or differ because one is a control and one is a label. Starting point is the label above, decided when looking at it on the phone (slice 1).

## Decisions log
Human-PM decisions in discussion (2026-10-17), recorded here as they shaped the scenarios:
- **R1 Heading whenever there are trips.** "Upcoming trips" shows whenever any trip exists, so it also needs an empty state when all are past (S3). This reverses 14/Q3, which chose no upcoming section. 14's decisions log points here.
- **R2 Upcoming is not collapsible.** Only the past group is a `<details>`.
- **R3 "Started" means any item prepared or packed.** Not only packed: gathering the first item into the pile counts, because that is when a parent has begun (S6). It is counted from the items each time, not stored (S8).
- **R4 Nothing in place of the bar.** No item count or other text takes the bar's place (S5).
- **R5 Empty state is a small line.** "No upcoming trips." with no button or icon; the page already has New trip at the top. Out 4.
- **R7 Heading reads "Upcoming trips" (Q1 settled, 2026-10-17).** Human-PM changed the wording at the phone check to match the past group's "Past trips (n)", and aligned it with the row text. It stays a label above the list, not inside the section.
- **R6 Two slices, one screen.** Heading and progress are two small changes to the same list template, so they are one milestone with two slices, each phone-checked, as in 14. Showing who is going is left out because it changes the data the list loads.

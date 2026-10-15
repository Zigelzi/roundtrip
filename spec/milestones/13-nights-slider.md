# Milestone 13: Nights slider

**Status:** done · **Branch:** feature/nights-slider (merged 2026-10-15)
<!-- Status: draft → red-teamed → in-progress → done -->

## Why
Flow: step 3, Plan what to pack (`../user-flows.md`), at its entry: creating the trip.

The new trip form asks for the same thing three ways: a departure date, a return date and a number of days, kept in step by a script. On a phone that is two date pickers and a number field for one fact, and "days" is not how the family thinks about a trip: cabins and hotels are booked by the night. This milestone replaces the return date and days with a nights slider under the departure date, and shows the return date it works out to, so a parent can check it against the booking ([`../mobile-redesign.md`](../mobile-redesign.md), R2).

## Scope
**In:**
One slice, phone-checked as a whole.

1. The form: destination, departure date, then a slider from 1 to 13 nights, starting at 2 nights (the 3 days the form starts at today). The return date and days fields go.
2. The return date: shown under the slider and kept up to date while the parent moves the slider or changes the departure, with its weekday (Q1). Before a departure is chosen, the same line holds a hint (Q2).
3. Saving: the trip is stored as today, in days, one more than the nights (R2): 7 nights is 8 days. A trip is 2 to 14 days, so a day trip can no longer be created (R2).

**Out:**
1. Changing a trip's dates after it is created: not possible today, and not added here.
2. Existing trips: a 1-day trip already saved keeps working and showing as it does.
3. Trips longer than 13 nights: the 14-day limit (04/S11) stays.
4. The past trips rule (a trip is past from the day after its return): milestone 14.

## Acceptance conditions (BDD)

### S1: The new trip form asks for nights
- Given the family with Parent 1 and Parent 2
- When a parent opens the new trip form
- Then it asks for a destination, a departure date and a number of nights
- And the nights start at 2
- And there is no return date or days field to fill in

### S2: The form shows the return date
- Given the new trip form, with the departure set to Sunday 20.12.2026
- When the parent moves the slider to 7 nights
- Then the form reads "7 nights"
- And it shows the return date "Sun 27.12.2026"
- And moving the departure to 21.12.2026 moves the return date to "Mon 28.12.2026"
- Verification: manual, plus the observable part in a test (see Test notes)

### S3: A trip is saved with the nights chosen
- Given the new trip form
- When a parent creates a trip to Levi departing 20.12.2026 for 7 nights
- Then the trip shows 20.–27.12.2026
- And a basic of one per day with no spare, such as Child 1's hair tie, comes to 8

### S4: A trip is 1 to 13 nights
- Given the new trip form
- When a parent sends 0 nights, or 14 nights (only possible without the slider, for example a stale page)
- Then no trip is saved
- And the form explains "Choose 1 to 13 nights."
- And it keeps the destination and departure the parent typed
- And the slider is back within 1 to 13 nights, with its words and return date matching where it sits

### S5: The form keeps the nights after a mistake
- Given a parent filling in the form with 7 nights and a departure date, but no destination
- When the parent taps Create trip
- Then the form asks for a destination
- And the slider is still at 7 nights
- And the return date for those 7 nights is still shown

### S6: The slider is easy to use with a thumb
- Given the new trip form on a phone
- When a parent drags the slider, or taps along it
- Then each night is a step the thumb can stop on
- And a screen reader hears "7 nights", not just "7"
- Verification: manual

## Test notes
- The tests are in `nights_slider_test.go`, written before the code.
- S2: the live update runs in the browser and cannot be observed by a Go test. The test checks the return date the server draws when it already knows the departure and nights (the page after a mistake, as in S5), with its weekday.
- S3: the trip's dates are read from the trips list; the per-day count from the trip page, as in 04/S2.
- S4: a table of cases: 1 and 13 nights are created as 2 and 14 days; 0, 14, none and 2.5 are rejected.
- S6: the spoken "7 nights" is checked on the slider in S5's test; how the slider feels under a thumb is the manual part.

Older tests whose wording or fields change (09/R1). Each keeps its intent:
- 01/S3 (form defaults): the nights start at 2, instead of days at 3 and an empty return date.
- 01/S6 (invalid input): the "return before departure" and "duration below 1" cases go, since those fields no longer exist; S4 here covers the limits.
- 01/S17 (return date instead of duration): retired, the return date is no longer typed. The return date is still worked out the same way (departure day counted), now tested by S2 and S3 here.
- 04/S11 (at most 14 days): becomes 13 nights, its message becomes S4's.
- The helpers that create trips in other tests send nights instead of days. A trip of N days is sent as N − 1 nights, so their numbers keep their meaning.
- 04/S2 (per-day items sized to the trip) and 06/S11 (a fixed basic on the shortest and longest trip): their 1-day case becomes 2 days, the shortest trip now.

## Design reference
Not taken from the canvas, built from the tokens in `cmd/web/tailwind.css` and judged at the phone check (Q4).

- **Departure:** the date field as today, full width, now that Return no longer shares its row.
- **Nights:** a label "Nights" with the value ("7 nights") on the right, and under it a full-width slider in the action blue, with a thumb at least 28px across.
- **Return date:** one muted line under the slider, "Return Sun 27.12.2026". Before a departure is chosen: "Pick a departure date to see the return date." (Q2).

## Open questions
None open.

## Decisions log
Light red-team of this draft, run inline by the main session (R19 in `../mobile-redesign.md`), 2026-10-15:
- **R1 Day trips.** Today a trip can be 1 day. After this milestone the shortest is 2 days (1 night), as decided in the plan (R2). Already saved 1-day trips are untouched (Out 2).
- **R2 Without the script.** If the page's script does not run, the slider still works and the trip is saved right; only the return date line does not follow the slider. Accepted: the script already runs on this form today.
- **R3 What the server trusts.** Only the nights. The return date is worked out from them, never sent, so the two cannot disagree (today's form needs a rule for when they do, 01/S17).
- **R4 Spoken value.** A slider is read out as its number, "7". The form gives it the words "7 nights" (S6), as the filter pills and the packing bar do for their counts.
- **R5 Dates across a month or year.** 30.12.2026 for 3 nights returns on 2.1.2027. The return date uses the same day arithmetic as the trip page's date range, so the two cannot disagree; S3 checks one case, and the existing date range tests cover the edges.
- **R6 Old tests.** Five older scenarios change wording or fields, each listed under Test notes with the reason. One (01/S17) is retired.
- **R7 Size.** One slice: the form, the validation and the tests, mostly changes to existing code in two files plus the tests. Small enough to run inline in the main session.

Open questions resolved (Human-PM, 2026-10-15), each as recommended:
- **Q1 Weekday in the return date:** shown, "Sun 27.12.2026". The line is there to check against the booking.
- **Q2 Before a departure is chosen:** a hint holds the return date's place, so the line does not jump in later.
- **Q3 Slider or stepper:** the slider, as on the canvas.
- **Q4 Design reference:** built from the proposal, without reading the canvas.

Allocation (Human-PM, 2026-10-15): one slice, inline in this main session, on Opus, tests written first.

Phone check (Human-PM, 2026-10-15): accepted as built.

Code review, round 1 (2026-10-15, Opus sub-agent, `/code-review medium`): no blocking findings, 7 non-blocking.
- **Fixed:** a refused nights value outside the slider (0, 14, none) was drawn as sent, while the browser moved the slider into range, so the label, the spoken value and the return date disagreed with it. The form now draws the slider at the nearest value it can show, and everything reads from that; S4 says so and its test checks it. The return date builds on `FormatDate` instead of its own copy of the format. The form's date layout lives in one place (`view.DateLayout`). The departure field listens for `input` only, as the old form did.
- **Declined:** moving the 13-night cap out of the view package. It sits beside `MaxBasicQuantity`, the same pattern, and one constant read by both the slider and the validation is what keeps them from drifting apart. The Status line stays `in-progress` until the merge (step 7).


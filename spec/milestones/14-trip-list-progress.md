# Milestone 14: Trip list progress and past trips

**Status:** done · **Branch:** feature/trip-list-progress (merged 2026-10-15)
<!-- Status: draft → red-teamed → in-progress → done -->

## Why
Flow: step 6, Pack (`../user-flows.md`), seen from the trips list, the first screen of the app.

Today the trips list shows each trip's destination and dates, and nothing about how far packing has got: to see that, a parent opens the trip. The list also mixes upcoming and past trips into one run, and a trip counts as past as soon as its departure has gone by, so a trip the family is on right now sorts among the past ones. This milestone shows packing progress on each upcoming trip's row, and moves past trips into their own group under the upcoming ones, closed by default (R9, [`../mobile-redesign.md`](../mobile-redesign.md), R3, R17).

## Scope
**In:**
Two slices, each phone-checked before the next.

1. Past trips: a trip counts as past from the day after its return date (R3), so a trip in progress stays with the upcoming ones. Past trips sit in their own group below them, headed "Past trips (n)", closed by default, and a parent can open it (R9 below, reversing R17's "open by default"). S1 to S6.
2. Progress: each upcoming trip's row shows a bar, green for packed and amber for prepared, like the packing page's header bar (10/S1), and the words "18 of 40 packed". It counts the same items as the trip page's packing bar (12), so the two always agree. A trip with everything packed reads "All packed". S7 to S10.

**Out:**
1. Progress on past trips in the list (R17). Their packing page still shows it (10/S8).
2. Remembering whether the past group was opened: it is closed again on every visit (R5 below).
3. Any change to the trip page, the packing page or the new trip form.
4. Deleting or archiving trips.

## Acceptance conditions (BDD)

### S1: Past trips have their own group
- Given today is 19.9.2026
- And trips to Parainen (20.–22.9.2026), Levi (1.–5.12.2026), Kuopio (10.–11.6.2026) and Turku (1.–3.8.2026)
- When a parent opens the trips list
- Then Parainen and Levi are listed first, nearest first
- And below them a group "Past trips (2)" lists Turku, then Kuopio, the most recent first
- And the group is closed

### S2: A trip in progress is not past
- Given a trip to Parainen from 20.9 to 22.9.2026
- When a parent opens the trips list on 22.9.2026, the return day
- Then Parainen is with the upcoming trips

### S3: A trip is past from the day after its return
- Given a trip to Parainen from 20.9 to 22.9.2026
- When a parent opens the trips list on 23.9.2026
- Then Parainen is in the past trips group

### S4: No past group without past trips
- Given only upcoming trips
- When a parent opens the trips list
- Then there is no past trips group

### S5: Only past trips
- Given only past trips
- When a parent opens the trips list
- Then the past trips group is shown, closed
- And the list does not say "No trips yet"
- And there is no empty upcoming section (Q3)

### S6: A parent can open the past trips group
- Given the trips list with the past trips group closed
- When the parent taps "Past trips (2)"
- Then the past trips are shown
- And tapping it again hides them
- Verification: manual, plus the observable part in a test (see Test notes)

### S7: An upcoming trip shows how far packing has got
- Given a trip to Levi with 40 items, 18 packed and 6 prepared
- When a parent opens the trips list
- Then Levi's row reads "18 of 40 packed"
- And its bar is green for the 18 packed and amber for the 6 prepared, the rest grey
- And the count is the same one the trip page's packing bar shows

### S8: A fully packed trip reads "All packed"
- Given an upcoming trip with every item packed
- When a parent opens the trips list
- Then its row reads "All packed"
- And its bar is all green

### S9: A trip with nothing on it shows no progress
- Given an upcoming trip with no items
- When a parent opens the trips list
- Then its row shows the destination and dates
- And no bar or packed count

### S10: Past trips show no progress
- Given a past trip with items, some of them packed
- When a parent opens the trips list
- Then its row shows the destination and dates
- And no bar or packed count

## Test notes
- The tests are in `trip_list_test.go`, written before the code.
- Every test fixes today with `fixedNow`, as 01/S2 does.
- S1 and S2: the order and the group are read from the page's text; "closed" is the group having no `open` attribute.
- S2 and S3: one table of days, 19.9 to 23.9, checking which group Parainen is in on each. 22.9 is the edge R3 moves.
- S6: opening and closing is the browser's own behaviour for a `<details>` group and cannot be observed by a Go test. The test checks the group is a `<details>` whose summary reads "Past trips (2)".
- S7: the bar widths are read from the segments' inline widths, as in 10/S1. The agreement with 12 is checked by reading the same trip's page and comparing its packing bar's count.
- S7 to S10: the screen reader text is the words "18 of 40 packed"; the bar is hidden from it, as on the packing page.

Older tests whose wording or behaviour changes (09/R1). Each keeps its intent:
- 01/S2 (upcoming first): still holds. Its past trip moves into the past group, which is still below the upcoming ones, so the order check passes unchanged.

## Design reference
Built from the tokens in `cmd/web/tailwind.css` and judged at the phone check (Q2).

- **Upcoming row:** as today (destination, then "20.–22.9.2026 · 3 days"), with a third line: a thin bar the full width of the text, and under it "18 of 40 packed" in the muted text size of the dates line. "All packed" is in the packed green.
- **Past group:** a second white section under the upcoming one, with its own gap above it. Its heading row "Past trips (2)" has a chevron that turns when the group closes, the same look as the packing page's "Packed (n)". The rows inside are the plain rows of today.

## Open questions
None open.

## Decisions log
Light red-team of this draft, run inline by the main session (R19 in `../mobile-redesign.md`), 2026-10-15:
- **R1 Two goals in one milestone.** README says a kickoff joining two goals with "and" is two milestones. The plan kept them together since both change one screen and each is small; they are split into two slices instead, each phone-checked, so feedback on the past group arrives before progress is built on it.
- **R2 Trip in progress.** Today a trip moves to past the morning after it departs. From now it stays upcoming through its return day, and shows its progress there, which is when packing for the way home happens. The first upcoming trip may therefore be one already under way; it is listed first, by departure, as today.
- **R3 What "today" is.** The Finnish date, as everywhere else (01/S5). A trip returning 22.9 is past from 00:00 on 23.9 Finnish time.
- **R4 Same count as the trip page.** The list counts every item on the trip, packed out of all, the same items the trip page's packing bar counts (12/S2). S7 checks the two agree, so a later change to one cannot quietly drift from the other.
- **R5 Opening is not remembered.** The group is a plain `<details>`, closed on every page load (R9). Remembering would need storage per phone for a choice that costs one tap. Out 2.
- **R6 Cost of counting.** The list needs each trip's counts. One query that counts per trip is simpler than reading every trip's items. At a family's few dozen trips either is instant; it would only matter at thousands of trips with thousands of items, which this app will not have.
- **R7 Back after packing.** A parent who packs and goes back to the list must see the new count. Pages are already sent with `no-store` (12, Code review), so Back fetches the list again; no new work.
- **R8 Size.** Each slice is a few dozen lines in the trips handler, one query and the list template, plus tests. Small enough to run inline in the main session.

Open questions resolved (Human-PM, 2026-10-15), each as recommended:
- **Q1 Where the count sits:** a third line under the dates, where the bar is too.
- **Q2 Design reference:** built from the proposal, without reading the canvas.
- **Q3 Only past trips:** no upcoming section, just the past group.

Allocation (Human-PM, 2026-10-15): both slices inline in this main session, on Opus, tests written first.

Phone check, slice 1 (Human-PM, 2026-10-15):
- **R9 Past trips closed by default.** Changed from open: past trips are rarely what a parent opens the app for, so the group starts closed and a tap opens it. This reverses the plan's R17 ("the past trips group is open by default") for this milestone; `../mobile-redesign.md` points here. With only past trips (S5), the list is the header and one closed "Past trips (n)" row.

Code review, round 1 (2026-10-15, Opus sub-agent, `/code-review medium`): no blocking findings, 7 non-blocking.
- **Fixed:** past trips were sorted by departure, so a long trip that came back last could sit below a short one inside it; they are now sorted by return day, most recent first, with a test. The Why section still said "open by default". The list query copied the trip's columns by hand; it now embeds the stored trip whole, so a new column cannot be silently dropped. The row's "show progress" switch repeated which group a trip is in; past trips now carry no counts, so the row needs no switch. S4's test failed on any `<details>` on the page, not only a past group; that check went.
- **Declined:** a separate type for the list's trips, so the counts are not on the shared trip. They are zero everywhere else and the field comment says so; a second type is more code for a risk the comment covers. Counting in SQL beside the trip page's count in Go: S7's test checks the two agree, which is the guard R4 asked for.

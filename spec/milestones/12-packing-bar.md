# Milestone 12: Packing button at the bottom

**Status:** done · **Branch:** feature/packing-bar (merged 2026-10-15)
<!-- Status: draft → red-teamed → in-progress → done -->

## Why
Flow: step 6, Pack, entered from step 3, Plan what to pack (`../user-flows.md`).

The trip page's "Start packing" button sits in the header's top corner, out of thumb reach on a phone, and it says nothing about how far packing has got. This milestone moves it to a full-width bar at the bottom of the trip page, labelled with the trip's packed count, as on the design canvas ([`../mobile-redesign.md`](../mobile-redesign.md), R4, R16, R20). It reuses the trip-wide packed count the packing page already shows, so it does not wait for milestone 10's bars.

## Scope
**In:**
One slice (Human-PM, 2026-10-15: the hiding is too thin to stand alone), phone-checked as a whole.

1. The bar: the packing button leaves the header and becomes a bar fixed to the bottom of the trip page. It reads "Start packing" while nothing is packed (also on a trip with no items, Q1, and while items are only prepared, Q2), "Continue packing 1/3" part way through, and "All packed" when everything is packed, and always opens the packing page (R16). The count follows items added or removed on the trip page, without reloading. The page leaves room at the bottom so the bar never covers the last add form.
2. Out of the keyboard's way: the bar is not shown while a text or number field on the trip page has focus, or while a row is open for editing, so the phone keyboard and the edit row keep their space (R4).

**Out:**
1. The bar on any other page: the packing page keeps its own header and progress (10), the basics page has no packing.
2. Progress bars or prepared counts in the bar: the label counts packed items only, like the "n of m packed" text (10, R22).
3. Live updates when the other parent packs on their own phone: the count is right as of page load and after this parent's own adds and removes (no live sync in v1).
4. Progress on the trips list: milestone 14.
5. Room for the iPhone home indicator when the app is saved to the home screen: it runs in the browser, whose own toolbar keeps that space.

## Acceptance conditions (BDD)

### S1: The packing button is at the bottom of the trip page
- Given a trip where Parent 1 has 3 items and none is packed
- When a parent opens the trip page
- Then a bar at the bottom of the page reads "Start packing"
- And tapping it opens the trip's packing page
- And the header no longer has a packing button

### S2: Part way through, the bar shows how far packing has got
- Given a trip with 3 items, 1 of them packed
- When a parent opens the trip page
- Then the bar reads "Continue packing 1/3"
- And a screen reader hears "1 of 3 packed"

### S3: When everything is packed, the bar says so
- Given a trip with 3 items, all packed
- When a parent opens the trip page
- Then the bar reads "All packed"
- And tapping it still opens the packing page

### S4: The count follows adding and removing items
- Given a trip with 3 items, 1 of them packed (Socks), and the trip page open
- When a parent adds Sunscreen
- Then the bar reads "Continue packing 1/4", without reloading
- And when the parent then removes Socks, the bar reads "Start packing"
- And when every remaining item is packed and the parent removes the last one still to pack, the bar reads "All packed"

### S5: The bar never covers the end of the page
- Given a trip page on a phone, long enough to scroll
- When a parent scrolls to the very bottom
- Then the last person's add form, with its Add button, sits fully above the bar
- And the bar is a full-width button in the action blue, with a line above it separating it from the list
- Verification: manual

### S6: The bar steps aside while typing
- Given the trip page on a phone
- When a parent taps a person's add field
- Then the bar is not shown while the keyboard is up
- And when the parent taps outside the field, the bar comes back
- And after adding an item, with the field still focused for the next one, the bar stays hidden
- Verification: manual

### S7: The bar steps aside while a row is open
- Given the trip page with the bar showing
- When a parent opens a row to change its quantity
- Then the bar is not shown
- And when the parent saves or cancels, the bar comes back
- Verification: manual, plus the observable part in a test (see Test notes)

## Test notes
- The tests are in `packing_bar_test.go`, written before the code.
- S1 to S3: the bar has its own id (`packing-bar`), so tests read it alone. "Header no longer has a packing button": the header (`<header>` to `</header>`) contains no link to the packing page.
- S2: the spoken count follows the filter pills (10): the visible "1/3" is hidden from screen readers and "1 of 3 packed" is read instead.
- S4: the htmx reply to adding or removing an item is what changes the page, so the test reads the reply: it carries the bar as a second, out-of-band fragment beside the person's list. Without JavaScript, adding and removing already redirect to the full page.
- S5, S6: how the bar looks and when the keyboard is up cannot be observed by a Go test.
- S7: the open row is drawn by the server, so the test checks that the reply opening a row carries the mark the hiding rule looks for, and the reply after saving does not. That the browser then hides the bar is the manual part.
- No existing test reads "Start packing", so no wording change ripples into older tests (09/R1).

## Design reference
Not taken from the canvas (Q3). From the tokens already in `cmd/web/tailwind.css`:

- **Bar:** fixed to the bottom of the screen, as wide as the page column, surface background, a line on top, 12px padding around the button.
- **Button:** the existing primary button, full width, 48px high. The label centred, the count after it in the same weight.
- **All packed:** same blue button (it is still an action), label "All packed".
- **Page bottom:** padding equal to the bar's height, so scrolling to the end shows the last add form above the bar.

## Open questions
None open.

## Decisions log
Light red-team of this draft, run inline by the main session (R19 in `../mobile-redesign.md`), 2026-10-15:
- **R1 Empty trip.** Raised as Q1.
- **R2 Prepared but not packed.** Raised as Q2.
- **R3 Stale count across phones.** The other parent packing on their own phone does not update this page's bar. Accepted, Out 3: the same holds for every count in the app today.
- **R4 Where the bar sits in the page.** Outside the items container that opening, saving and cancelling a row replace (04), so those swaps leave it alone. Adding and removing items send the bar back as a second fragment. Changing a quantity does not change the count, so it needs nothing.
- **R5 Covering content.** The page gets bottom padding the bar's height (S5). Without it the last Add button would sit under the bar.
- **R6 Spoken count.** "1/3" reads as "1 slash 3", so the bar follows the filter pills' pattern (S2).
- **R7 How the bar hides (slice 2).** One CSS rule: while any field on the page has focus, or an open edit row exists, the bar is not shown. No JavaScript and no Alpine. It relies on the CSS `:has()` selector, supported since Safari 15.4 (2022) and Chrome 105 (2022), see https://caniuse.com/css-has. An older browser keeps the bar showing, which is slice 1's behaviour, not a breakage.
- **R8 Stepper buttons.** Tapping − or + takes focus from the name field, the phone keyboard closes and the bar comes back while the parent is still in the add form. Accepted: with the keyboard gone the space is free again. Judged at the phone check.
- **R9 Between the slices.** Moot: one slice (see Scope).
- **R10 The header.** Without the button, the destination and dates get the header's full width; nothing replaces it there.
- **R11 Slices share files.** Moot: one slice, a few dozen lines, inline in the main session ("Size before spawning", `../constitution.md`).

Open questions resolved (Human-PM, 2026-10-15), each as recommended:
- **Q1 A trip with no items yet:** the bar shows "Start packing"; the empty packing page explains itself.
- **Q2 Items prepared but none packed:** "Start packing" until something is in the bag, since the label counts packed items only (R22).
- **Q3 Design reference:** built from the proposal, judged at the phone check, without reading the canvas.

Allocation (Human-PM, 2026-10-15): one slice, inline in this main session, on Opus, tests written first.

Phone check (Human-PM, 2026-10-15): accepted as built, and again after the review fixes (the Back gesture count and the bar alignment).

Code review, round 1 (2026-10-15, Opus sub-agent, `/code-review medium`): no blocking findings, 8 non-blocking.
- **Fixed:** coming back with the browser's Back gesture showed the count from before packing. The first fix (reload when the page is restored from the back/forward cache, the `pageshow` pattern in `trip_new.templ`) did not help at the phone check. Measured in headless Chrome (`../../docs/visual-debugging.md`): Back re-showed the page from the browser's download cache, not the back/forward cache, so the reload never ran. Every page the app renders now says `Cache-Control: no-store`, and the probe then showed the fresh count after Back. The `pageshow` reload stays for browsers that keep a no-store page whole in memory; headless Chrome does not use that cache, so that case is untested. S7's test now also checks the page after saving a row, as these test notes say. The spacer is taller than the bar on its own instead of relying on the page's padding. The bar's button lines up with the page content (16px sides, not 12px). The count is taken in the same loop as each person's items, as on the packing page, instead of a second pass. The plan in `../mobile-redesign.md` no longer lists two slices.
- **Escalated, kept (Human-PM, 2026-10-15):** after an add the name field keeps focus, so the bar stays hidden, also on an iPhone that has closed its keyboard (tapping Add closes it, and iOS does not reopen it for the page's own refocus). Kept as is (option A): the bar is one tap away, and stopping the refocus would cost a tap per item when adding several.
- **Declined:** the spoken "1 of 3 packed" differs from the visible "1/3", which voice control (saying the visible text) does not match. Kept, to match the filter pills (10/S2); nobody in the family uses voice control.

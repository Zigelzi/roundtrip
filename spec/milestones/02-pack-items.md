# Milestone 02: Pack items

**Status:** done · **Branch:** feature/pack-items
<!-- Status: draft → red-teamed → in-progress → done -->

## Why
Flow: Pack everything (`../user-flows.md` step 5).
As a parent I want to tick off items as I put them in the bag, so that I can see what is still left to pack and nothing is forgotten.

Planning and packing are separate activities, in the app and in real life: planning is sitting down and deciding what to bring, packing is standing at the bed with the bag. They happen at different times and need different screens — so packing gets its own page, and the trip page stays the planning view. Packing also varies: sometimes the whole family is packed in one go, sometimes one parent packs only their own things and the other members' items are noise.

## Scope
**In:**
1. An item's place in its lifecycle, stored as a status. This milestone implements only `planned` ⇄ `packed`; see [Domain](#domain).
2. Packing is per item, not per unit — "Underwear, 4 pcs" is one tick, not four.
3. A packing page for a trip, reached from the trip page, showing every member's items grouped by member. It opens on the whole family.
4. Marking an item packed, and unpacking it again — things get taken back out.
5. Packed items move out of the way into a collapsed section, so the list of what is left shrinks as packing proceeds. The section stays open while a parent works inside it.
6. Progress: how many of the trip's items are packed.
7. Narrowing the page to a single family member, for packing one person's things. The narrowing holds while packing — every tap keeps it.
8. Packing in place (HTMX): one tap updates the member's list, the packed section and the progress together, and keeps the parent's place in the list. The page must stay usable tapped through ~120 items in a row.
9. The whole row is the tap target, not a small checkbox.

**Out:**
1. The rest of the lifecycle — `prepared`, `needs_buying`, `bought` (see [`../domain-model.md`](../domain-model.md)). "Missing" in this milestone means "not yet packed", nothing more.
2. Partial quantities (2 of 4 packed).
3. Categories or any grouping other than by family member.
4. Adding, editing or removing items on the packing page — that stays planning, on the trip page. Revisit after a real trip has been packed with the app.
5. Recording which items had to be bought, and the post-trip review of which were useful.
6. Live sync between the two parents: a parent sees the other's ticks on their next page load, not as they happen.
7. Unpacking the whole trip at once / resetting the list.
8. Packing progress on the front page's trip list, and on the trip page — the trip page gets a plain link to the packing page.
9. Remembering which member the page was narrowed to between visits.
10. Auth — still single player, same as milestone 01.

## Domain
Entities are defined in [`../domain-model.md`](../domain-model.md), which also holds the item lifecycle and how it is stored. This milestone gives Item its status field and implements two of its values: an item starts `planned` and is packed or unpacked between `planned` and `packed`. Once `prepared` exists, unpacking will return an item there instead — the same transition with the middle of the chain built.

## Acceptance conditions (BDD)
<!-- S4 was folded into S2 (persistence is what S2 stores); the number is retired, not reused. -->

### S1: Parent opens the packing page for a trip
- Given a trip where Parent 1 has "Underwear, 4 pcs" and "Toothbrush, 1 pcs", and Parent 2 has "Raincoat, 1 pcs"
- And none of them are packed
- When the parent opens the trip page and follows the link to pack
- Then they see a section per family member, each listing that member's unpacked items with their quantity
- And they see that 0 of 3 items are packed
- And they see a way to narrow the page to each family member, with the one they are viewing marked as current
- And the packed section shows how many items are in it, and that it can be opened
- And there is a link back to the trip page

### S2: Parent marks an item as packed
- Given the parent is on the packing page and "Toothbrush" is not packed
- When they tap the row for "Toothbrush"
- Then "Toothbrush" is no longer in Parent 1's list of things to pack
- And it is listed in the packed section
- And the progress says 1 of 3 items are packed
- And it is still packed when the packing page is opened again later

Verification: manual on a phone — the whole row is comfortably tappable, ticking ~20 items in a row keeps the parent's place in the list rather than jumping to the top, and the page stays responsive. Automated: the reply contains the item's member list, the packed section and the progress, all updated, and the item is stored as packed.

### S3: Parent unpacks an item
- Given "Toothbrush" is packed
- When the parent opens the packed section and taps "Toothbrush"
- Then "Toothbrush" is back in Parent 1's list of things to pack
- And the progress says 0 of 3 items are packed
- And the packed section is still open, so another item can be unpacked without reopening it
- And its count has gone down by one

Verification: manual on a phone for the section staying open. Automated: the reply contains the updated lists and progress, and the item is stored as not packed.

### S5: Parent packs one person's things only
- Given a trip where Parent 1 and Parent 2 both have items
- When the parent narrows the packing page to Parent 1
- Then they see only Parent 1's items
- And the progress counts only Parent 1's items
- And when they pack one of Parent 1's items, the page is still narrowed to Parent 1 afterwards
- And they can return to seeing every member

### S6: Parent sees when everything is packed
- Given every item on the trip is packed
- When the parent opens the packing page
- Then they see that everything is packed instead of an empty list of things to pack
- And the progress says 3 of 3 items are packed

### S7: Parent opens the packing page for a trip with no items
- Given a trip with no items on it
- When the parent opens its packing page
- Then they see a message that there is nothing to pack yet
- And a link to the trip page to add items

### S8: Parent opens the packing page for a trip that does not exist
- Given there is no trip with the requested id
- When the parent opens that trip's packing page
- Then they see a "trip not found" message with a link back to the front page
- And the response status is 404

### S9: A member with nothing planned reads differently from one who is done
- Given Child 2 has no items on the trip at all
- And every one of Child 1's items is packed
- When the parent opens the packing page
- Then Child 2's section says they have nothing planned
- And Child 1's section says they are fully packed

### S10: Parent taps an item the other parent has already removed
- Given the parent has the packing page open
- And the item they are about to tap has since been removed from the trip page
- When they tap it
- Then nothing is saved
- And the list they are looking at comes back without that item, so they can carry on packing
- And an item belonging to another trip cannot be packed through this trip's page

## Open questions
<!-- For the red-team pass (workflow step 2) and Human-PM to resolve (step 3). -->
- None open.

## Decisions log
Kickoff (Human-PM, 2026-09-20): packed is binary — "missing" means "not yet packed", not "needs buying". Packing lives on its own page, because planning and packing are separate activities in real life. Packed items collapse into a section rather than staying in place.

Red-team pass resolved (Human-PM, 2026-09-20): the whole row is the tap target, not a checkbox — a mis-tap costs one tap to undo, unlike milestone 01's delete, and one-handed packing needs the bigger target. The page opens on the whole family, with narrowing one tap away, rather than asking whose things first. Adding an item while packing stays out; revisit once a real trip has been packed with the app. Decided by the main session as the simpler option: a tap on a since-removed item quietly re-renders instead of erroring (S10); a member with nothing planned reads differently from a finished one (S9); the packed section keeps the same grouping and order as the to-pack list so an item can be found in it; the trip page gets a plain link, not a progress indicator; the narrowing is not remembered between visits.

Storage (Human-PM, 2026-09-20): one status field, no separate packed flag — see the lifecycle in [`../domain-model.md`](../domain-model.md). The draft proposed a flag on the assumption that unpacking was ambiguous; Human-PM corrected the lifecycle (a bought item rejoins the pile rather than going straight into the bag), which gives unpacking a single destination and removes the reason for a second field.

Accepted (Human-PM, 2026-09-20) after testing on a phone. Two fixes went in from that pass, both marking state the parent could not see: the active filter link is marked as current, and the packed section shows its count and that it opens (flexing a <summary> had dropped the browser's disclosure triangle). Code review found and fixed one real bug -- overlapping taps could land out of order and visually un-tick an item, since every reply is a full snapshot; rows now serialise their requests.

Known implementation hazard (red-team): a native `<details>` for the packed section loses its open state on every HTMX swap, which S3 forbids. Resolve it with `hx-preserve`, a swap that leaves the wrapper alone, or Alpine — Alpine only if the lighter options fail.

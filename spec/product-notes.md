# Product notes: why steps feel the way they do, and ideas

Reference for sessions working on Roundtrip. [`user-flows.md`](user-flows.md) is Human-PM's: the flow, the success measures and the current rating of each step. This file holds the detail behind those ratings and the ideas for improving them. Agents keep it up to date; Human-PM reads it when deciding what to build next.

What is built is not restated here: each milestone's Status line in `milestones/` is the record. When a milestone addresses an observation below, add "Addressed by NN" to it, and leave judging whether it worked to Human-PM's next rating.

## Observations per step

Dated, newest last within each step. Step numbers are those of `user-flows.md` (see the renumbering rule in [`README.md`](README.md#flow-steps)).

- **Step 1, Define the basics.** 2026-09-27: the most problematic step. Each person's basics were fixed in the app and could not be reviewed or tuned ([A10](activity-catalogue.md#open-items)). Addressed by 06 (editable basics).
- **Step 2, Plan activities.** Not built yet.
- **Step 3, Plan what to pack.** 2026-09-27: long list; most of the editing is correcting the basics (step 1). Partly addressed by 06 (fix the basics once instead of on every trip) and 07 (only the people going get items).
- **Step 4, Gather items.** 2026-10-10: works, but which items are already in the pile is kept in our heads: an item is either packed or not, with no "prepared" state between. Addressed by 08 (prepared state).
- **Step 5, Review what is missing.** Not built yet.
- **Step 6, Pack.** 2026-10-10: the worst step after testing. Packed items moved to one section at the end of the page instead of staying under each person, so checking a person's bag meant looking in two places. The long list is tiring to review. Addressed by 08 (packed items under each person).
- **After the trip.** Not built yet.

## Learning from use

- 2026-10-10: milestone 02 moved packed items into one collapsed section at the end of the page, so that what is left to pack shrinks as packing goes on. That serves "what is left?" but works against a second job that showed up in real use, "is this person's bag right?", which needs a person's packed items next to their unpacked ones. Addressed by 08.
- Each person packs their own bag, so the packing page's person filter already matches how we pack.

## Ideas for keeping a growing list manageable

The number of items grows with every activity and trip. What limits the app is our attention on a phone screen, not the data. Milestone 05 made long lists easier to move through. Beyond that, the lever is showing fewer things that need attention, not showing more of them.

Ideas are numbered `I1`, `I2`, ... and the numbers are stable (never renumber; a dropped idea keeps its number and says so). Not decided or scheduled unless marked.

- **I1 Editable basics per person.** Review and tune each person's basics once (add, remove, change quantities) instead of correcting the same rows on every trip ([A10](activity-catalogue.md#open-items)). Serves step 1. The post-trip review (I2) changes the same basics. **Built in 06.**
- **I2 Post-trip review by exception.** Assume everything was useful and only ask "what did you not use?" and "what was missing?". This keeps the review to a couple of minutes, and it is what shrinks the list over time instead of only growing it. Its answers adjust the basics (step 1), so it needed I1 first.
- **I3 Kits.** One row that stands for a set of things packed together (toiletry bag, nappy bag), checked when the kit is restocked rather than on every trip. We already keep some kits, but not everything is in one. Changes what an item is, so it is a larger change.
- **I4 Group by where things are gathered.** Order a list by where in the house things are picked up (bathroom, wardrobe, storage room), as grocery apps order by store aisle, rather than by category. Serves step 4. Weigh against categories (I6): categories help check, locations help gather.
- **I5 Essentials check.** A short "don't leave without" list (ID, medicine, chargers) kept apart from the bulk. Low priority, since we rarely forget things.
- **I6 Categories as lists within the list.** Group each person's items into categories (outdoor, hygiene, sports, general), as the old Notion page did, so a whole category can be checked as packed and correct in one go. Fewer things to review at once, which helps steps 1 and 3. Wanted since milestone 04. Open questions from the 2026-10-11 codebase audit: a fixed list or editable by the parents; is the packed part grouped too; can a whole category be packed in one tap; what category existing items and basics get.

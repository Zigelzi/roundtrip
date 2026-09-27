# User flows - What users do in Roundtrip

Roundtrip will be used by our family and should be built fully to our family, without any need for multi-tenancy, minimal configuration, only minimal login ideally with Google account.

Our family includes two adults and two young children. The two adults are the ones preparing and packing together, so this is two player collaboration app.

## High level flow
1. Plan what activites we're doing on the trip.
2. Plan what to pack based on that.
3. Gather the items to one place (e.g on top of bed).
4. Review if something is missing and needs to be bought.
5. Pack everything after confirming everything is available.

After the trip I'd want to review what is useful and which were not to remove unuseful items from the future packing list. This is to optimize the packing in the future. I don't do this because it's too time consuming and hard to maintain with current tools.
## What success looks like per step

We rarely forget anything today. The cost is that preparing is an intensive task, so the success measure for Roundtrip is **how much time and energy we spend preparing for a trip**, not how few things we forget.

Each step has a value moment (when the app has clearly helped) and a rough measure. The measures are a gut-feel way for us to pick which step to improve next, assessed by us after a trip, not something the app tracks.

| Step | Value moment | Rough measure | Current feel (2026-09-27) |
|---|---|---|---|
| 1. Plan activities | Picking what we will do takes a minute, and the list follows from it | Minutes from "new trip" to activities chosen | Not built yet |
| 2. Plan what to pack | The generated list is nearly right for this trip and each person, so we only adjust | How much of the list we had to review and edit by hand; energy spent | **Most problematic.** The list is long and each person's basics are fixed, so we cannot review and tune them per person ([A10](activity-catalogue.md#open-items)) |
| 3. Gather items | We walk the house once, in an order that matches where things are | Trips around the house; time to gather | Works, not measured |
| 4. Review what is missing | What needs buying is obvious without reading the whole list | Time to a shopping list | Not built yet |
| 5. Pack | Packing each person's bag is quick and nothing is packed twice | Time to pack; taps that went to the wrong row | Works; improved by milestone 05 |
| After the trip | Saying what was not used, or was missing, takes a couple of minutes | Whether we actually do it; list shrinking trip by trip | Not built yet |

## Keeping a growing list manageable

The number of items grows with every activity and trip. What limits the app is our attention on a phone screen, not the data. Milestone 05 made long lists easier to move through. Beyond that, the lever is showing fewer things that need attention, not showing more of them. Ideas, none of them decided or scheduled:

1. **Editable basics per person.** Review and tune each person's basics once (add, remove, change quantities) instead of correcting the same rows on every trip ([A10](activity-catalogue.md#open-items)). Serves step 2, the weakest step today.
2. **Post-trip review by exception.** Assume everything was useful and only ask "what did you not use?" and "what was missing?". This keeps the review to a couple of minutes, and it is what shrinks the list over time instead of only growing it.
3. **Kits.** One row that stands for a set of things packed together (toiletry bag, nappy bag), checked when the kit is restocked rather than on every trip. We already keep some kits, but not everything is in one. Changes what an item is, so it is a larger change.
4. **Group by where things are gathered.** Order a list by where in the house things are picked up (bathroom, wardrobe, storage room), as grocery apps order by store aisle, rather than by category. Serves step 3. Weigh against the categories wanted in milestone 04.
5. **Essentials check.** A short "don't leave without" list (ID, medicine, chargers) kept apart from the bulk. Low priority, since we rarely forget things.

Each person packs their own bag, so the packing page's person filter already matches how we pack.

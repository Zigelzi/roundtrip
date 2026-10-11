# Roundtrip: mobile UI redesign brief

I'm attaching phone screenshots of every screen of Roundtrip, a small packing app. Please redesign its look and feel. Everything visual is open for change: colour scheme, typography, spacing, layout, component styles. **Usability on a phone comes before style**: when the two conflict, the easier-to-use option wins.

## What the app is and who uses it

- A packing app for one Finnish family: two parents and two young children. The two parents plan and pack together, each on their own phone. No other users, no login.
- Used on phones about 99% of the time, often one-handed: the other hand holds clothes or a bag. Typical phone width is about 390px.
- The flow is a loop: define each person's **basics** once (what they always bring) → a new trip starts from those basics → plan the trip's list → gather items into one pile → pack the bags. After the trip the basics get adjusted.
- The success measure is how much time and energy preparing takes, not how few things get forgotten. Fewer things to look at, fewer taps and fewer wrong taps all count.

## The screens (screenshots attached)

1. **Trips (front page).** App name, a "New trip" button (primary), a "Basics" button (secondary), and the list of trips, upcoming first. Each trip shows its destination and "17.–18.10.2026 · 2 days".
2. **New trip.** Destination, departure date, return date and number of days (return date and days update each other), and "Who is going": one toggle pill per person, all ticked by default. "Create trip" button. Validation problems show in one red box at the top.
3. **Trip page (plan what to pack).** Sticky header with destination, dates and a "Start packing →" button, plus one jump link per person that scrolls to their section. Then one card per person: their name with an item count ("17 items"), their items (name, quantity such as "3 pcs", and a ✕ to remove), and an add form at the bottom (name field with suggestions, − / quantity / + stepper, Add button). Tapping a quantity opens that row inline for editing (highlighted amber row with − / + stepper, Save and Cancel); while one row is open, every other row's buttons are locked.
4. **Packing page (used standing at the bag).** Sticky header: "Packing · destination" and progress ("12 of 75 packed"), with a green "Everything is packed." line when done. A filter to show Everyone, Family or one person (2-column grid of pills; the active one is solid dark). Then one card per person with their remaining items. **The whole row is the tap target.** Each item moves through three states:
   - **Planned** (empty square box): tap → prepared.
   - **Prepared**, meaning gathered into the pile (circle with a dot, "Prepared" label under the name): tap → packed. A small ↶ button on the row takes it back to planned.
   - **Packed** (green tick, struck through): moves into a collapsible "Packed (n)" part at the bottom of that person's card. Tap → back to unpacked.
   There is deliberately no delete on this page, so packing is never one mis-tap away from losing an item.
5. **Basics.** Same layout as the trip page (sticky header with jump links, one card per person). Each basic has a quantity rule: "1 / day + 1 extra", "6 / day" or a fixed "2". Tapping the quantity opens an inline edit row (name, per day, extra, Save, Cancel). Add form at the bottom of each card.

Please also design these states: an empty state (no trips yet, a person with no items), an inline edit row open on the trip page, a validation error, a person's list fully packed ("All packed."), and the packing page filtered to one person.

## Typical data

Use this, not lorem ipsum. People are shown as **Parent 1, Parent 2, Child 1, Child 2** plus **Family**, a shared bucket for things that belong to nobody in particular (stroller, wet wipes, toothpaste). Item names are a mix of Finnish and English, because the family types whatever comes to mind.

- A typical trip: "Wedding in Tampere", 17.–18.10.2026, 2 days, everyone going: about **75 items in total**, 16 to 19 per person and about 6 for Family. A short weekend trip with only the two parents has about 33 items.
- Parent 1: Kalsarit 3 pcs, Sukat 3 pcs, T-paidat 2 pcs, Kollarit 2 pcs, **Pitkät siistit housut** 1 pc, Vyö, Pitkähihanen, Huppari, Ulkotakki, Hattu, Hanskat, Siistit kengät, Hammasharja, Puhelimen laturi, **Avaimet ja lompakko**, Dödö
- Parent 2: Pikkarit 3 pcs, Socks 4 pcs, T-shirt 2 pcs, Trousers 2 pcs, Long-sleeved shirt, Pyjamas, Outdoor jacket, Hattu, Gloves, Shoes 2 pcs, Toothbrush, Hairbrush, Water bottle, Phone charger, Wallet and keys, Rintaliivit 2 pcs
- Child 1 (basics): Underwear "2 / day + 1 extra", Socks "1 / day + 1 extra", Outfit "1 / day + 1 extra", Long-sleeved shirt 2, Hat, Gloves, Pyjama dress, Pyjama pants, Pyjama socks, Outdoor jacket, **Outdoor overalls**, Shoes 2, Toothbrush, Water bottle, Small towel, Sleep toy 4, Hairbrush, Hair spray, Hair tie "1 / day"
- Child 2 (basics): similar, plus Sleeping bag, Pacifier 2, Nappies "6 / day"
- Family: Kids toothpaste, Adults toothpaste, Wet wipes, Painkillers, **Tuplarattaat**, **Rattaiden sadesuoja** 2 pcs
- Mid-packing, a realistic mix for one person: 9 packed, 4 prepared, 3 still planned.
- Destinations can be long Finnish phrases, around 20 to 25 characters.

## Usability rules to keep

These came from real use on a phone; please keep them or argue for a better alternative.

- Tap targets at least 48px tall; packing rows taller (currently 64px), with the whole row tappable.
- Input text at 16px or more (smaller makes iOS zoom in on focus).
- State is never shown by colour alone: planned, prepared and packed differ in shape as well (box, circle with dot, tick).
- On a phone, each person's list runs edge to edge (no stacked page and card margins wasting width). On a wide screen the content stops at about 672px.
- Sticky headers stay slim (one or two short lines plus the jump links/filter): every pixel they take is a pixel less of a 75-item list, especially with the keyboard open.
- One primary (solid) button per screen; everything else looks secondary.
- Destructive actions (remove ✕) look clearly different from other actions, and the packing page has none.

## Known pain points

- **Packing is the weakest step.** The long list is tiring to review, and the second job, "is this person's bag right?", needs a person's packed items close to their unpacked ones. Packed items currently sit in a collapsed part at the bottom of each person's card.
- Planning a trip is mostly scrolling through and correcting a long list.
- It's not obvious at a glance how far along each person is (only a trip-wide total in the header).
- A later feature will group each person's items into categories (outdoor, hygiene, clothes). No need to design it now, but don't make it hard to add.

## Technical constraints

The app is server-rendered HTML (Go templates) styled with **Tailwind CSS v4 utility classes**, with HTMX swapping parts of the page after a tap. There is no JavaScript framework.

- Design with what plain HTML and Tailwind can do. Native elements (`<details>`, `<dialog>`, checkboxes) are preferred over custom widgets.
- Avoid interactions that need client-side state: swipe gestures, drag and drop, animated reordering.
- System font stack preferred; one web font is fine if it clearly earns its place.
- Light mode only.
- Don't add features that don't exist yet (activities, categories, shopping list, login). If you have ideas, list them separately instead of mocking them.

## What I'd like back

1. Phone-sized artboards (about 390 × 844) for each of the five screens and the states listed above, using the typical data.
2. A short set of design tokens that map to Tailwind: colour palette (with roles such as primary, destructive, prepared, packed), type scale, spacing, corner radius.
3. A ranked list of the changes you made, each with the usability reason behind it, most important first.

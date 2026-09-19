# Domain model

Shared entities across milestones. One definition per entity; milestone specs link here.

- Family member — one of the 4 fixed people in the family: Parent 1 (dad), Parent 2 (mom), Child 1 (kid), Child 2 (kid). Seeded, not editable. Separate from app users (kids own items but never log in).
- Trip — destination, departure date, duration in days. The return date is derived, not stored: departure + duration − 1 (the departure day counts).
- Item — one thing a family member takes on a trip: trip + member + name + quantity. Later it gains a state (Planned → Prepared / Needs to be bought → Bought → Packed); in this milestone every item is implicitly Planned. Reuse across trips is by name.

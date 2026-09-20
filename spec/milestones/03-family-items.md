# Milestone 03: Family items

**Status:** done · **Branch:** feature/family-items
<!-- Status: draft → red-teamed → in-progress → done -->

## Why
Flow: Plan what to pack (`../user-flows.md` step 2) and pack everything (step 5).

Not everything we take belongs to a person. Sunscreen, the first aid kit, the board game and the tent are packed once for the family, not once each. Today the app has no place for them: every item must have an owner, so a shared thing has to be parked on one parent, which makes that parent's list wrong and makes milestone 02's "pack only my things" view misleading.

This came out of drafting [`../activity-catalogue.md`](../activity-catalogue.md): on a worked 4-day trip, 13 of 88 items were shared, more than half of what either parent carries personally. It goes before activity planning because it changes what an item *is*, and designing it underneath activities would mean designing two things at once.

## Scope
**In:**
1. A "Family" bucket that items can belong to, alongside the four people.
2. Adding and removing a family item on the trip page, the same way member items work today.
3. Family items on the packing page: their own section, packed and unpacked like any other.
4. Family items counted in the trip's packing progress.
5. Narrowing the packing page to the Family bucket, alongside narrowing to a person.
6. A filter row that works at phone width. It already wraps with five links (Everyone plus four people) and Family makes six, so this milestone fixes it rather than making it worse. Per `../constitution.md`, tap-target size and reachability count as function, not polish.
7. Updating [`../domain-model.md`](../domain-model.md): after this milestone a family member is no longer "one of the 4 fixed people" and an item no longer belongs to a person. That file is the one place a later milestone reads, so it cannot be left stale.

**Out:**
1. Activities and generated items. That is milestone 04, and it assumes this bucket exists.
2. Per-day quantities (decided for milestone 04, not needed here).
3. Who is responsible for carrying a family item, or which bag it goes in.
4. Any change to how the four people are seeded or named.
5. Moving an existing item between a person and the Family bucket.
6. Auth: still single player.
7. Noticing that "Sunscreen" exists both under a person and under Family. The UNIQUE constraint is per owner, so those are two rows and the parent gets no hint. Correct for two toothbrushes, wrong for one bottle of sunscreen, and out of scope here.

## Domain
Entities are defined in [`../domain-model.md`](../domain-model.md). This milestone adds the idea that an item's owner can be the family rather than a person. The item lifecycle is unchanged: a family item is `planned` or `packed` exactly like a personal one.

### Storage options

The item table today requires an owner: `family_member_id INTEGER NOT NULL REFERENCES family_member (id)`, with `UNIQUE (trip_id, family_member_id, name_key)`. Only two places read the member list (`cmd/web/trip_page.go:50`, `cmd/web/pack_page.go:129`), and both feed "a section per member" straight into the page.

| | 1. Extra row in `family_member` | 2. Nullable `family_member_id` | 3. Separate `shared_item` table |
|---|---|---|---|
| Migration | One INSERT | Drop NOT NULL, rebuild the table | New table |
| `item` table | Untouched | Changed | Untouched |
| Uniqueness | Works as-is | Breaks: SQLite treats NULLs as distinct, so "Sunscreen" could be added twice. Needs a partial unique index | Needs its own |
| Trip page, packing page, filter, progress | Work unchanged; Family appears as an extra section | Every query and grouping needs a NULL branch | Two code paths everywhere |
| Cost of being wrong | `family_member` holds a row that is not a person | Real work to undo | Real work to undo |

Decided: **option 1**. It is one INSERT and the feature largely falls out of code that already exists. The honest cost is conceptual: `family_member` stops being "people" and becomes "things items can belong to".

#### The bucket's id is 100, not 5

The draft said id 5 and claimed `FAMILY_NAMES` was unaffected. That was wrong, and an existing test proves it. `internal/db/config.go:64-79` maps names to ids by position (`id := int64(i+1)`) and errors only when the update matches zero rows. With a row at id 5, `FAMILY_NAMES="Ada,Bo,Cy,Di,Ed"` would match, return no error, and quietly rename the Family bucket to "Ed", producing a fifth person who owns the shared items. `cmd/web/family_names_test.go:72-85` exists to make exactly that typo loud and would start failing.

Seeding the bucket at a deliberately out-of-range id fixes this at the source: positional `FAMILY_NAMES` can never reach it, `ApplyFamilyNames` needs no change, the guard test passes untouched, and `ORDER BY id` still puts the bucket last. Use one named Go constant (`familyBucketID`) rather than a bare 100 scattered through the code, and say in the migration comment why the id is out of range.

#### `ListPeople` alongside `ListFamilyMembers`

`ListFamilyMembers` now returns five rows, which is what the trip page, the packing page and the filter all want. Three places want only the four people:

- `memberNames` in `cmd/web/family_names_test.go:13-24` returns every row, so all four sub-cases of `TestFamilyNamesComeFromConfiguration` compare five names against four.
- `ApplyFamilyNames` conceptually renames people, not buckets.
- **Milestone 04**, which is the real reason. `../activity-catalogue.md` defines `Everyone` as the four people. An expansion step written as "for each member" against `ListFamilyMembers` would silently give the Family bucket a copy of every personal item, roughly 17 extra rows on the worked example, and it would look correct in review.

Adding the query now is a few lines and removes a trap that is invisible later.

#### Migration Down

`family_member_id` has no `ON DELETE` (`sql/schema/20260919130000_add_family_member_and_item.sql:28`) and `internal/db/config.go:26` turns `foreign_keys` on, so a Down that just deletes the bucket row fails once any family item exists. The Down deletes the bucket's items first.

## Acceptance conditions (BDD)
<!-- Drafted from the scope above; sharpen at workflow step 4, after the red-team pass. -->

### S1: Parent adds an item that belongs to the whole family
- Given a trip with no items
- When the parent adds "Sunscreen, 1 pcs" to the Family bucket on the trip page
- Then it appears in the Family section, not under any person
- And it is still there when the trip page is opened again

### S2: Family items show up on the packing page as their own section
- Given a trip where Parent 1 has "Toothbrush" and the family has "Sunscreen" and "First aid kit"
- When the parent opens the packing page
- Then they see a Family section listing the two family items
- And the progress says 0 of 3 items are packed, counting family items alongside personal ones
- And the Family section comes after all four people's sections, on both the trip page and the packing page
- And in the filter row, Family comes second, directly after "Everyone"

### S3: Parent packs a family item
- Given "Sunscreen" is not packed
- When the parent taps its row
- Then it moves into the packed section like any other item
- And the progress goes up by one

### S4: Parent packs only the family things
- Given the packing page is open on the whole family
- When the parent narrows it to Family
- Then they see only the family items, and the progress counts only those
- And the narrowing holds after packing one of them

### S5: Parent removes a family item
- Given the trip has "Board game" in the Family bucket
- When the parent removes it on the trip page
- Then it is gone from the Family section and from the packing page

### S6: A trip with only personal items still reads correctly
- Given a trip where nothing belongs to the family
- When the parent opens the trip page and the packing page
- Then the Family section is still shown, saying there is nothing in it, exactly the way an empty member section does under milestone 02's S9
- And this holds on trips with no family items at all, which is most trips until milestone 04

### S7: The same family item cannot be added twice
- Given the family already has "Sunscreen" on the trip
- When the parent adds "sunscreen" again
- Then it is rejected the same way a duplicate personal item is, and the existing one is untouched

### S8: A parent packing one person is told the trip is not finished
- Given a trip of 61 items where every one of Parent 1's is packed and the family's are not
- When the parent has the packing page narrowed to Parent 1
- Then the page says "Parent 1 is packed." followed by "34 of 61 packed on the trip." (the name is whatever `FAMILY_NAMES` supplies for that member, not the literal placeholder)
- And when the page is narrowed to Family and every family item is packed, it says "The family's things are packed." followed by the same trip line
- And on the whole-family view, with items still unpacked, neither message is shown
- And when everything on the trip is packed, "Everything is packed." is shown, unchanged from milestone 02, on the whole-family view and under a filter alike: the person is done and so is the trip, so the trip-wide message is the honest one and the narrowed wording would only repeat it

The wording is pinned here because the test asserts the literal string. Two things follow from it:

- The current message is wrong, not just imprecise: `Packed`/`Total` in `cmd/web/view/pack.go:15-22` count only what is shown, so `pack_page.templ:32-34` prints "Everything is packed." whenever one filtered member is finished. Family is the bucket nobody owns, so it is the one most likely to be sitting behind that message.
- This is **not** a pure wording change. `cmd/web/pack_page.go:152` computes `TripTotal` and there is no trip-wide packed count, so the second line needs one adding.

### S9: The packing filter is usable at phone width
- Given the packing page with six filter links: Everyone, Family and the four people
- When the parent opens it on a phone
- Then every filter link is reachable and comfortably tappable without the row breaking into a ragged stack
- And the one currently being viewed is still visibly marked as current

Verification: manual on a phone; a Go test cannot observe layout. Automated: S9 asserts its own coverage, namely that all six links are present and exactly one carries `aria-current="page"`. It cannot inherit this from milestone 02's S1: that check loops over `familyMembers` (`cmd/web/trip_page_test.go:13`), a four-element slice, so it proves four links and never six.

The markup is `memberFilter` in `cmd/web/view/pack_page.templ:40`. The shape of the fix (a wrapping grid, a select, tighter pills) is an implementation choice, with one constraint: do not make the row scroll horizontally. Names come from `FAMILY_NAMES` and can be long, so a scrolling row would push Family off-screen on first paint, which is why it sits second rather than last.

## Open questions
<!-- For the red-team pass (workflow step 2) and Human-PM to resolve (step 3). -->

- None open. Q1 to Q5 and the red-team findings are all resolved; see the decisions log.

## Decisions log
Open questions resolved (Human-PM, 2026-09-20):
- **Q1 Storage:** option 1, an extra row in `family_member`. The id was corrected from 5 to an out-of-range 100 by the red-team pass; see the Domain section. The item table and its UNIQUE constraint are untouched.
- **Q2 Sort order:** Family last among the sections, because specific things come before general ones. The filter row is the exception, see the red-team decisions below.
- **Q3 Personal filter:** "pack only my things" means only that person's things. Family is its own filter alongside the four people, not appended to a personal one. A reminder that family items are still unpacked is deliberately deferred until after this milestone; see S8 for the smaller wording fix that goes in now.
- **Q4 Label:** "Family", as drafted.
- **Q5 Name suggestions:** no preference, so leave `ListItemNames` unchanged. Family item names keep being suggested for people and vice versa. Zero work, and a wrong-but-harmless suggestion costs one ignored tap.


Red-team findings resolved (Human-PM, 2026-09-20):
- **Bucket id 5 was unsafe.** Accepted: seed at an out-of-range id behind a named constant. The draft's claim that `FAMILY_NAMES` was unaffected was wrong and an existing guard test proved it.
- **Filter placement.** Family sits second in the filter row, directly after "Everyone", while the sections keep it last. The filter is navigation and the sections are the work order, so they do not have to match: the bucket nobody owns should not be the one you have to hunt for. Horizontal scrolling is ruled out for that row for the same reason.
- **S8 wording.** Pinned to exact copy, with the trip-wide number included: "Parent 1 is packed." plus "34 of 61 packed on the trip." The number no longer matches the chosen filter, which is the point, because it is what stops the family pile being forgotten while the reminder is deferred. The message says one person is ready while things are still missing elsewhere; that is its whole job. Confirmed after testing (Human-PM, 2026-09-20): once the whole trip is packed, a narrowed view falls back to "Everything is packed." rather than repeating the narrowed wording. The draft scenario covered only the case where the old message was a lie, not the case where it became true again; `TestNarrowedViewOnAFullyPackedTrip` pins it.
- **Empty Family section.** Always shown, consistent with milestone 02's S9 for people. Most trips will have an empty one until milestone 04, and that is accepted.
- Also accepted: update `domain-model.md` (scope 7), add `ListPeople` and a `familyBucketID` constant, delete the bucket's items in the migration Down, assert the section order in S2, and list the person-plus-Family duplicate as out of scope.

Accepted (Human-PM, 2026-09-20) after testing both slices on a phone, with no defects found at either check. Code review (medium) found no correctness bug and five low findings, four fixed and one folded in. Its best catch was that nothing pinned `ListPeople`'s literal id to the `familyBucketID` constant, so the two could drift apart silently and the query would start returning the bucket as a person, which is exactly the milestone 04 trap it exists to prevent. A comment cannot catch that; a test can.

Kickoff (Human-PM, 2026-09-20): split out of the milestone 04 activity-planning draft. Activities (milestone 04) are chosen at trip creation and cannot be changed afterwards, so this milestone's Family bucket has to work for hand-added items on its own, independently of anything generated. Nothing in the scope above depends on activities existing. Shared items are real and need their own bucket before activities start generating them. "Family" is the agreed term. Accommodation ("where we sleep") is not an activity and is out of scope for both milestones.

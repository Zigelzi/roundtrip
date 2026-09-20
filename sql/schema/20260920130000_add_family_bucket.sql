-- +goose Up
-- The Family bucket is an item's owner when it belongs to the whole family
-- rather than one person (spec/milestones/03-family-items.md). It is stored
-- as an extra family_member row, not a new table or a nullable
-- family_member_id, so item, trip page, packing page and progress queries
-- need no changes: the bucket is just a fifth "member".
--
-- Seeded at id 100, deliberately out of range. internal/db/config.go's
-- ApplyFamilyNames maps FAMILY_NAMES entries onto ids 1, 2, 3, ... by
-- position and errors only when an entry matches no row. A bucket at id 5
-- would let a five-entry FAMILY_NAMES list (one name too many, a typo)
-- silently rename the bucket to a fifth person instead of failing loudly;
-- see cmd/web/family_names_test.go's TestFamilyNamesRejectsMoreNamesThanMembers.
-- An out-of-range id makes that impossible. Referenced in Go as the
-- familyBucketID constant (cmd/web/app.go).
INSERT INTO family_member (id, name) VALUES (100, 'Family');

-- +goose Down
-- family_member_id has no ON DELETE and foreign_keys is on (internal/db's
-- InitDB), so the bucket's items have to go before the bucket row does.
--
-- WARNING: this destroys data. Every family item on every trip goes, not just
-- the bucket row, and goose runs it when stepping back past this migration on
-- the way to an earlier one. Back the database file up before any goose down
-- on a real install.
DELETE FROM item WHERE family_member_id = 100;
DELETE FROM family_member WHERE id = 100;

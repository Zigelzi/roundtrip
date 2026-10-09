-- +goose Up
-- Basic names are unique per family member, ignoring capitalisation, the same
-- rule as item (spec/milestones/06-edit-basics.md, R1). Without it a double-
-- tapped Add, or two parents adding at once, could save two rows of one name,
-- and then every new trip would fail: creating a trip copies the basics in one
-- transaction and the item table refuses the duplicate.
--
-- name_key is the name lowercased by Go (itemKey in cmd/web/items.go). The
-- seeded rows are filled with SQL lower(), which is safe only because the seed
-- is ASCII. SQLite cannot add a UNIQUE constraint to an existing table, so the
-- rule is a unique index. The empty default is only there so the column can be
-- added; the app always writes the key.
ALTER TABLE basic_item ADD COLUMN name_key TEXT NOT NULL DEFAULT '';
UPDATE basic_item SET name_key = lower(name);
CREATE UNIQUE INDEX basic_item_member_name_key ON basic_item (family_member_id, name_key);

-- +goose Down
DROP INDEX basic_item_member_name_key;
ALTER TABLE basic_item DROP COLUMN name_key;

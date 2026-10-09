-- Keep query files ASCII-only (see sql/queries/item.sql).

-- name: ListBasicItems :many
-- In the order a new trip's items are added: owner, then catalogue position.
-- Includes the Family bucket, so it does not go through ListPeople.
SELECT id, family_member_id, name, per_day, fixed
FROM basic_item
ORDER BY family_member_id, position;

-- name: UpdateBasicItem :execrows
-- The name (with its comparison key) and both quantities change together, as
-- the edit form saves them. Zero rows means the basic is gone (the other
-- parent removed it), which is not an error.
UPDATE basic_item
SET name = ?, name_key = ?, per_day = ?, fixed = ?
WHERE id = ?;

-- name: CreateBasicItem :exec
-- The position is worked out inside the statement (last for that family
-- member), so two adds at once cannot both read the same highest number and
-- clash on UNIQUE (family_member_id, position). Removing leaves gaps, which
-- is harmless: only the order matters.
INSERT INTO basic_item (family_member_id, position, name, name_key, per_day, fixed)
VALUES (
    ?1,
    (SELECT COALESCE(MAX(position), 0) + 1 FROM basic_item WHERE family_member_id = ?1),
    ?2, ?3, ?4, ?5
);

-- name: DeleteBasicItem :one
-- Returning the owner says whose section to go back to; no row means the
-- basic is already gone (R5).
DELETE FROM basic_item
WHERE id = ?
RETURNING family_member_id;

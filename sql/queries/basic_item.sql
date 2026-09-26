-- Keep query files ASCII-only (see sql/queries/item.sql).

-- name: ListBasicItems :many
-- In the order a new trip's items are added: owner, then catalogue position.
-- Includes the Family bucket, so it does not go through ListPeople.
SELECT family_member_id, name, per_day, fixed
FROM basic_item
ORDER BY family_member_id, position;

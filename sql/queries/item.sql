-- Keep query files ASCII-only: a multi-byte character (even in a comment)
-- made sqlc slice the following queries at the wrong offsets.

-- name: ListFamilyMembers :many
SELECT id, name
FROM family_member
ORDER BY id;

-- name: UpdateFamilyMemberName :execrows
UPDATE family_member
SET name = ?
WHERE id = ?;

-- name: ListTripItems :many
SELECT id, family_member_id, name, quantity
FROM item
WHERE trip_id = ?
ORDER BY id;

-- name: ListItemNames :many
-- Names used on any trip, suggested when adding items so names get reused.
-- One spelling per name ignoring case (Sunscreen and sunscreen become one).
SELECT CAST(MIN(name) AS TEXT) AS name
FROM item
GROUP BY name_key
ORDER BY name_key;

-- name: CreateItem :exec
INSERT INTO item (trip_id, family_member_id, name, name_key, quantity)
VALUES (?, ?, ?, ?, ?);

-- name: DeleteItem :one
-- Scoped to the trip so an item can't be removed through another trip's URL.
DELETE FROM item
WHERE id = ? AND trip_id = ?
RETURNING family_member_id;

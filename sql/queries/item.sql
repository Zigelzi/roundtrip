-- Keep query files ASCII-only: a multi-byte character (even in a comment)
-- made sqlc slice the following queries at the wrong offsets.

-- name: ListFamilyMembers :many
SELECT id, name
FROM family_member
ORDER BY id;

-- name: ListPeople :many
-- The four people, excluding the Family bucket (id 100, cmd/web's
-- familyBucketID). For logic that means "each person" and must not hand the
-- bucket a copy of every personal item, such as milestone 05's activity
-- expansion.
SELECT id, name
FROM family_member
WHERE id != 100
ORDER BY id;

-- name: UpdateFamilyMemberName :execrows
UPDATE family_member
SET name = ?
WHERE id = ?;

-- name: ListTripItems :many
SELECT id, family_member_id, name, quantity, status
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

-- name: MoveItemStatus :one
-- One step of the lifecycle (08 Scope 6): the item moves only if it is in the
-- status the step starts from, so a parent whose page is out of date cannot
-- move an item backwards or two steps at once. The check lives in the UPDATE
-- so two phones tapping at once cannot both win. Scoped to the trip for the
-- same reason DeleteItem is: an item id from another trip must not be
-- reachable through this trip's page. No row means nothing moved, either
-- because the item is gone or because it was in another status (see
-- GetItemOwner).
UPDATE item
SET status = sqlc.arg(to_status)
WHERE id = sqlc.arg(id) AND trip_id = sqlc.arg(trip_id) AND status = sqlc.arg(from_status)
RETURNING family_member_id;

-- name: GetItemOwner :one
-- Whose list to re-render when a step moved nothing. No row means the item
-- is not on this trip (removed, or an id from another trip).
SELECT family_member_id
FROM item
WHERE id = ? AND trip_id = ?;

-- name: DeleteItem :one
-- Scoped to the trip so an item can't be removed through another trip's URL.
DELETE FROM item
WHERE id = ? AND trip_id = ?
RETURNING family_member_id;

-- name: UpdateItemQuantity :one
-- Scoped to the trip like DeleteItem. Only the quantity changes: a packed
-- item stays packed (R2). No row means the item is not on this trip.
UPDATE item
SET quantity = ?
WHERE id = ? AND trip_id = ?
RETURNING family_member_id;

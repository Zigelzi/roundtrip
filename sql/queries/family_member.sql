-- Keep query files ASCII-only: a multi-byte character (even in a comment)
-- made sqlc slice the following queries at the wrong offsets.

-- name: ListFamilyMembers :many
SELECT id, name
FROM family_member
ORDER BY id;

-- name: ListPeople :many
-- The four people, excluding the Family bucket (id 100, cmd/web's
-- familyBucketID). For anything that means "each person": the "who is going"
-- choices on a new trip (milestone 07) and the names FAMILY_NAMES renames.
-- The bucket is not a person, so it must never be offered as one.
SELECT id, name
FROM family_member
WHERE id != 100
ORDER BY id;

-- name: UpdateFamilyMemberName :execrows
UPDATE family_member
SET name = ?
WHERE id = ?;

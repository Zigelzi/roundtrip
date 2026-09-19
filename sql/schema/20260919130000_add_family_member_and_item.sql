-- +goose Up
-- The family is fixed (spec/domain-model.md); there is no screen to manage it.
-- Members are separate from app_user: kids own items but never log in.
CREATE TABLE family_member (
    id   INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);
INSERT INTO family_member (id, name) VALUES
    (1, 'Parent 1'),
    (2, 'Parent 2'),
    (3, 'Child 1'),
    (4, 'Child 2');

-- An item is one thing a member takes on a trip.
-- name_key is the name lowercased by Go (see itemKey in cmd/web/items.go).
-- It backs "same item only once per member per trip, ignoring case": SQLite's
-- NOCASE/lower() only fold ASCII, so they'd treat "Ämpäri" and "ämpäri" as
-- different. The UNIQUE constraint also stops a double-tapped Add from
-- saving the item twice.
CREATE TABLE item (
    id               INTEGER PRIMARY KEY,
    trip_id          INTEGER NOT NULL REFERENCES trip (id) ON DELETE CASCADE,
    family_member_id INTEGER NOT NULL REFERENCES family_member (id),
    name             TEXT    NOT NULL,
    name_key         TEXT    NOT NULL,
    quantity         INTEGER NOT NULL CHECK (quantity >= 1),
    UNIQUE (trip_id, family_member_id, name_key)
);
CREATE INDEX item_trip_id ON item (trip_id);

-- +goose Down
DROP TABLE item;
DROP TABLE family_member;

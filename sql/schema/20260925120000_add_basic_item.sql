-- +goose Up
-- The basics are the items every trip starts with (spec/milestones/04-trip-basics.md).
-- Creating a trip copies each row into the trip's items, with quantity
-- per_day x duration_days + fixed. From milestone 04 on this table, not
-- spec/activity-catalogue.md, is the source of truth for the basics; the
-- catalogue is where changes are discussed and may drift from it.
--
-- Rows are keyed by family_member_id, never by name: FAMILY_NAMES renames
-- the people at startup. 1 to 4 are the people, 100 the Family bucket.
-- position is the catalogue order within each owner, which is also the order
-- the items are added to a trip and so the order the trip page lists them.
CREATE TABLE basic_item (
    id               INTEGER PRIMARY KEY,
    family_member_id INTEGER NOT NULL REFERENCES family_member (id),
    position         INTEGER NOT NULL,
    name             TEXT    NOT NULL CHECK (name != ''),
    per_day          INTEGER NOT NULL CHECK (per_day >= 0),
    fixed            INTEGER NOT NULL CHECK (fixed >= 0),
    CHECK (per_day + fixed >= 1),
    UNIQUE (family_member_id, position)
);

INSERT INTO basic_item (family_member_id, position, name, per_day, fixed) VALUES
    -- Parent 1
    (1, 1, 'Underwear', 1, 1),
    (1, 2, 'Socks', 1, 1),
    (1, 3, 'T-shirt', 1, 0),
    (1, 4, 'Sweatpants', 0, 1),
    (1, 5, 'Khakis', 0, 1),
    (1, 6, 'Belt', 0, 1),
    (1, 7, 'Long-sleeved shirt', 0, 1),
    (1, 8, 'Hoodie', 0, 1),
    (1, 9, 'Outdoor jacket', 0, 1),
    (1, 10, 'Hat', 0, 1),
    (1, 11, 'Gloves', 0, 1),
    (1, 12, 'Shoes', 0, 2),
    (1, 13, 'Toothbrush', 0, 1),
    (1, 14, 'Water bottle', 0, 1),
    (1, 15, 'Phone charger', 0, 1),
    (1, 16, 'Wallet and keys', 0, 1),
    (1, 17, 'Deodorant', 0, 1),
    -- Parent 2
    (2, 1, 'Underwear', 1, 1),
    (2, 2, 'Socks', 1, 2),
    (2, 3, 'T-shirt', 1, 0),
    (2, 4, 'Trousers', 0, 2),
    (2, 5, 'Jumper', 0, 1),
    (2, 6, 'Long-sleeved shirt', 0, 1),
    (2, 7, 'Pyjamas', 0, 1),
    (2, 8, 'Outdoor jacket', 0, 1),
    (2, 9, 'Hat', 0, 1),
    (2, 10, 'Gloves', 0, 1),
    (2, 11, 'Shoes', 0, 2),
    (2, 12, 'Toothbrush', 0, 1),
    (2, 13, 'Hairbrush', 0, 1),
    (2, 14, 'Water bottle', 0, 1),
    (2, 15, 'Phone charger', 0, 1),
    (2, 16, 'Wallet and keys', 0, 1),
    -- Child 1
    (3, 1, 'Underwear', 2, 1),
    (3, 2, 'Socks', 1, 1),
    (3, 3, 'Outfit', 1, 1),
    (3, 4, 'Long-sleeved shirt', 0, 2),
    (3, 5, 'Hat', 0, 1),
    (3, 6, 'Gloves', 0, 1),
    (3, 7, 'Pyjama dress', 0, 1),
    (3, 8, 'Pyjama pants', 0, 1),
    (3, 9, 'Pyjama socks', 0, 1),
    (3, 10, 'Outdoor jacket', 0, 1),
    (3, 11, 'Outdoor overalls', 0, 1),
    (3, 12, 'Shoes', 0, 2),
    (3, 13, 'Toothbrush', 0, 1),
    (3, 14, 'Water bottle', 0, 1),
    (3, 15, 'Small towel', 0, 1),
    (3, 16, 'Sleep toy', 0, 4),
    (3, 17, 'Hairbrush', 0, 1),
    (3, 18, 'Hair spray', 0, 1),
    (3, 19, 'Hair tie', 1, 0),
    -- Child 2
    (4, 1, 'Socks', 1, 1),
    (4, 2, 'Outfit', 1, 1),
    (4, 3, 'Long-sleeved shirt', 0, 2),
    (4, 4, 'Hat', 0, 1),
    (4, 5, 'Gloves', 0, 1),
    (4, 6, 'Pyjamas', 0, 1),
    (4, 7, 'Outdoor jacket', 0, 1),
    (4, 8, 'Outdoor overalls', 0, 1),
    (4, 9, 'Shoes', 0, 2),
    (4, 10, 'Toothbrush', 0, 1),
    (4, 11, 'Water bottle', 0, 1),
    (4, 12, 'Sleep toy', 0, 1),
    (4, 13, 'Sleeping bag', 0, 1),
    (4, 14, 'Pacifier', 0, 2),
    (4, 15, 'Small towel', 0, 1),
    (4, 16, 'Nappies', 6, 0),
    (4, 17, 'Hairbrush', 0, 1),
    -- Family
    (100, 1, 'Kids toothpaste', 0, 1),
    (100, 2, 'Adults toothpaste', 0, 1),
    (100, 3, 'Wet wipes', 0, 1),
    (100, 4, 'Painkillers', 0, 1),
    (100, 5, 'Double stroller', 0, 1),
    (100, 6, 'Stroller rain cover', 0, 2);

-- +goose Down
-- Items already generated from the basics stay on their trips, as ordinary
-- items: nothing links an item back to the basics row it came from.
DROP TABLE basic_item;

-- +goose Up
-- departure_date is a calendar date (YYYY-MM-DD) with no time zone: a trip
-- departs on 20.9 wherever you read it. Only "today" is derived from the clock.
CREATE TABLE trip (
    id             INTEGER PRIMARY KEY,
    destination    TEXT    NOT NULL,
    departure_date TEXT    NOT NULL,
    duration_days  INTEGER NOT NULL CHECK (duration_days >= 1)
);

-- +goose Down
DROP TABLE trip;

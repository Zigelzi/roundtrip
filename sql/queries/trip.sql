-- name: ListTrips :many
SELECT id, destination, departure_date, duration_days
FROM trip
ORDER BY departure_date, id;

-- name: CreateTrip :one
INSERT INTO trip (destination, departure_date, duration_days)
VALUES (?, ?, ?)
RETURNING id;

-- name: GetTrip :one
SELECT id, destination, departure_date, duration_days
FROM trip
WHERE id = ?;

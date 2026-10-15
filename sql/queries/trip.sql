-- name: ListTrips :many
-- Each trip with its packing counts, for the trips list (14/S7): every item
-- on the trip, the same items the trip page's packing bar counts (14/R4).
-- One query for the whole list rather than one per trip (14/R6).
SELECT
    sqlc.embed(trip),
    COUNT(item.id) AS total,
    CAST(COALESCE(SUM(item.status = 'packed'), 0) AS INTEGER) AS packed,
    CAST(COALESCE(SUM(item.status = 'prepared'), 0) AS INTEGER) AS prepared
FROM trip
LEFT JOIN item ON item.trip_id = trip.id
GROUP BY trip.id
ORDER BY trip.departure_date, trip.id;

-- name: CreateTrip :one
INSERT INTO trip (destination, departure_date, duration_days)
VALUES (?, ?, ?)
RETURNING id;

-- name: GetTrip :one
SELECT id, destination, departure_date, duration_days
FROM trip
WHERE id = ?;

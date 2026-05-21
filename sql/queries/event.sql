-- name: CreateEvent :one
INSERT INTO events (
    id, name, description, venue, image, banner_image, event_date, status
)
VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;


-- name: GetEventByID :one
SELECT *
FROM events
WHERE id = $1;


-- name: ListEvents :many
SELECT *
FROM events
ORDER BY created_at DESC;


-- name: ListUpcomingEvents :many
SELECT *
FROM events
WHERE event_date >= CURRENT_DATE
ORDER BY event_date ASC;


-- name: ListEventsByMode :many
SELECT *
FROM events
WHERE status = $1
ORDER BY event_date ASC;


-- name: UpdateEvent :one
UPDATE events
SET
    name = $2,
    description = $3,
    venue = $4,
    image = $5,
    banner_image = $6,
    event_date = $7,
    status = $8,
    updated_at = NOW()
WHERE id = $1
RETURNING *;


-- name: DeleteEvent :one
DELETE FROM events
WHERE id = $1
RETURNING id;

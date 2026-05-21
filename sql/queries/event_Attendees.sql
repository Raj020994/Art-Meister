-- name: EnrollUserToEvent :one
INSERT INTO event_attendees (
    id, event_id, user_id
)
VALUES (
    $1, $2, $3
)
RETURNING *;


-- name: RemoveUserFromEvent :one
DELETE FROM event_attendees
WHERE event_id = $1 AND user_id = $2
RETURNING event_id;


-- name: ListEventAttendees :many
SELECT u.*
FROM users u
JOIN event_attendees ea ON ea.user_id = u.id
WHERE ea.event_id = $1
ORDER BY ea.joined_at ASC;


-- name: CountEventAttendees :one
SELECT COUNT(*)::int
FROM event_attendees
WHERE event_id = $1;


-- name: ListMyEvents :many
SELECT e.*
FROM events e
JOIN event_attendees ea ON ea.event_id = e.id
WHERE ea.user_id = $1
ORDER BY e.event_date ASC;

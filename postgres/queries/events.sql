-- name: CreateEvent :exec
INSERT INTO events (
    event_id, calendar_id, name, location, description, notification, frequency, priority, start_time, end_time
)
VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
);

-- name: GetEvent :one
SELECT *
FROM events
WHERE event_id = $1;

-- name: UpdateEvent :exec
UPDATE events 
SET calendar_id = $1, name = $2, location = $3, description = $4, notification = $5, frequency = $6, priority = $7, start_time = $8, end_time = $9
WHERE event_id = $10;

-- name: DeleteEvent :exec
DELETE FROM events
WHERE event_id = $1;

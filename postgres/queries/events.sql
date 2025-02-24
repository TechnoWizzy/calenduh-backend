-- name: CreateEvent :one
insert into events (event_id, calendar_id, name, location, description, notification, frequency, priority, start_time, end_time)
values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
returning *;

-- name: GetEventById :one
select *
from events
where event_id = $1;

-- name: GetEventByCalendarId :many
select *
from events
where calendar_id = $1;

-- name: UpdateEvent :one
update events
set calendar_id = $1, name = $2, location = $3, description = $4, notification = $5, frequency = $6, priority = $7, start_time = $8, end_time = $9
where event_id = $10 and calendar_id = $1
returning *;

-- name: DeleteEvent :exec
delete from events
where event_id = $1 and calendar_id = $2;

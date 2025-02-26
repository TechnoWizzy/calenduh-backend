-- name: GetAllEvents :many
select *
from events
where start_time between $1 and sqlc.arg(end_time) or end_time between $1 and sqlc.arg(end_time);

-- name: GetEventById :one
select *
from events
where event_id = $1;

-- name: GetEventsByUserId :many
select e.*
from users u
inner join calendars c on u.user_id = c.user_id
inner join events e on c.calendar_id = e.calendar_id
where u.user_id = $1 and (start_time between $2 and sqlc.arg(end_time) or end_time between $2 and sqlc.arg(end_time));

-- name: GetEventByCalendarId :many
select *
from events
where calendar_id = $1 and (start_time between $2 and sqlc.arg(end_time) or end_time between $2 and sqlc.arg(end_time));

-- name: CreateEvent :one
insert into events (event_id, calendar_id, name, location, description, notification, frequency, priority, start_time, end_time)
values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
returning *;

-- name: UpdateEvent :one
update events
set calendar_id = $1, name = $2, location = $3, description = $4, notification = $5, frequency = $6, priority = $7, start_time = $8, end_time = $9
where event_id = $10 and calendar_id = $1
returning *;

-- name: DeleteEvent :exec
delete from events
where event_id = $1 and calendar_id = $2;

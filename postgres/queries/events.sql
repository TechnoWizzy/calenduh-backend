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
insert into events (event_id, calendar_id, name, location, description, notification, frequency, priority, start_time, end_time, all_day, first_notification, second_notification)
values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
returning *;

-- name: UpdateEvent :one
update events
set calendar_id = $3, name = $4, location = $5, description = $6, notification = $7, frequency = $8, priority = $9, start_time = $10, end_time = $11, all_day = $12, first_notification = $13, second_notification = $14, last_edited = now()
where event_id = $1 and calendar_id = $2
returning *;

-- name: DeleteEvent :exec
delete from events
where event_id = $1 and calendar_id = $2;

-- name: DeleteAllEvents :exec
delete from events
where true;
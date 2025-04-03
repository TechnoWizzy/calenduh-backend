-- name: GetAllEvents :many
select *
from events
where start_time < sqlc.arg(end_time);

-- name: GetEventById :one
select *
from events
where event_id = $1;

-- name: GetEventsByUserId :many
select e.*
from users u
inner join calendars c on u.user_id = c.user_id
inner join events e on c.calendar_id = e.calendar_id
where u.user_id = $1  and start_time < sqlc.arg(end_time);

-- name: GetEventsByGroupId :many
select e.*
from groups g
    inner join calendars c on g.group_id = c.group_id
    inner join events e on c.calendar_id = e.calendar_id
where g.group_id = $1  and start_time < sqlc.arg(end_time);

-- name: GetEventByCalendarId :many
select *
from events
where calendar_id = $1 and start_time < sqlc.arg(end_time);

-- name: CreateEvent :one
insert into events (event_id, calendar_id, name, location, description, notification, frequency, priority, start_time, end_time, all_day, first_notification, second_notification)
values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
returning *;

-- name: UpdateEvent :one
update events
set name = $3, location = $4, description = $5, notification = $6, frequency = $7, priority = $8, start_time = $9, end_time = $10, all_day = $11, first_notification = $12, second_notification = $13, last_edited = now()
where event_id = $1 and calendar_id = $2
returning *;

-- name: DeleteEvent :exec
delete from events
where event_id = $1 and calendar_id = $2;

-- name: DeleteAllEvents :exec
delete from events
where true;
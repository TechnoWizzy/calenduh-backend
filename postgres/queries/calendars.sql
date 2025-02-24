-- name: GetAllCalendars :many
select * from calendars;

-- name: GetCalendarById :one
select * from calendars
where calendar_id = $1;

-- name: GetCalendarsByUserId :many
select * from calendars
where user_id = $1;

-- name: GetCalendarsByGroupId :many
select * from calendars
where group_id = $1;

-- name: GetSubscribedCalendars :many
select distinct c.* from users u
inner join subscriptions s on u.user_id = s.user_id
inner join calendars c on s.calendar_id = c.calendar_id
where u.user_id = $1;

-- name: CreateCalendar :one
insert into calendars (calendar_id, user_id, group_id, title, is_public)
values ($1, $2, $3, $4, $5)
returning *;

-- name: DeleteCalendar :exec
DELETE FROM calendars
WHERE calendar_id = $1;

-- name: DeleteAllUserCalendars :exec
DELETE FROM calendars
WHERE user_id = $1;

-- name: UpdateCalendar :one
update calendars
set title = $1, is_public = $2, user_id = $3, group_id = $4
where calendar_id = $5
returning *;
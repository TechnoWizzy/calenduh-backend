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
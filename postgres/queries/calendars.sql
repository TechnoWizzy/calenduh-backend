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

-- name: CreateCalendar :exec
INSERT INTO calendars (
    calendar_id, user_id, group_id, title, is_public
)
VALUES (
    $1, $2, $3, $4, $5
);

-- name: DeleteCalendar :exec
DELETE FROM calendars
WHERE calendar_id = $1;

-- name: DeleteAllUserCalendars :exec
DELETE FROM calendars
WHERE user_id = $1;

-- name: UpdateCalendar :exec
UPDATE calendars
SET title = $1, is_public = $2, user_id = $3, group_id = $4
WHERE calendar_id = $5;
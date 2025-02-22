-- name: GetAllSessions :many
select * from sessions;

-- name: GetSessionById :one
select * from sessions
where session_id = $1;

-- name: InsertSession :one
insert into sessions (session_id, user_id, type, access_token, refresh_token, expires_on) values (
$1,
$2,
$3,
$4,
$5,
$6
) returning *;

-- name: DeleteSession :exec
delete from sessions session
where session_id = $1;
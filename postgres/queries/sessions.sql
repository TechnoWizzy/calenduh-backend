-- name: GetSessionById :one
select * from sessions
where session_id = @session_id;

-- name: InsertSession :one
insert into sessions (session_id, user_id, type, access_token, refresh_token, expires_on) values (
@session_id::text,
@user_id::text,
@type::session_type,
@access_token::text,
@refresh_token::text,
@expires_on::date
) returning *;

-- name: DeleteSession :exec
delete from sessions session
where session_id = @session_id;
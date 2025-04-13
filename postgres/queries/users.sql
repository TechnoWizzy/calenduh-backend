-- name: GetAllUsers :many
select * from users;

-- name: GetUserById :one
select * from users
where user_id = $1;

-- name: CreateUser :one
insert into users (user_id, email, username)
values ($1,$2,$3)
returning *;

-- name: UpdateUser :one
update users
set email = $2, username = $3, birthday = $4, name = $5, default_calendar_id = $6, is_24_hour = $7
where user_id = $1
returning *;

-- name: DeleteUserProfilePicture :exec
update users
set profile_picture = null
where profile_picture = $1;

-- name: DeleteUser :exec
delete from users
where user_id = $1;

-- name: DeleteAllUsers :exec
delete from users
where true;
-- name: GetAllUsers :many
select * from users;

-- name: GetUserById :one
select * from users
where user_id = $1;

-- name: GetUserByEmail :one
select * from users
where email = $1;

-- name: CreateUser :one
insert into users (user_id, email, username)
values ($1,$2,$3)
returning *;

-- name: UpdateUser :one
update users
set email = $1, username = $2
where user_id = $3
returning *;

-- name: DeleteUser :exec
delete from users
where user_id = $1;

-- name: DeleteAllUsers :exec
delete from users
where true;
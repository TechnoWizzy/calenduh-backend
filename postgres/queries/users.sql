-- name: GetAllUsers :many
select * from users;

-- name: GetUserById :one
select * from users
where user_id = $1;

-- name: GetUserByEmail :one
select * from users
where email = $1;

-- name: InsertUser :one
insert into users (user_id, email, username) values (
    $1,
    $2,
    $3
) returning *;
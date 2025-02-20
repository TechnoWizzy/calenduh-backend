-- name: GetAllUsers :many
select * from users;

-- name: GetUserById :one
select * from users
where user_id = @user_id;

-- name: GetUserByEmail :one
select * from users
where email = @email::text;

-- name: InsertUser :one
insert into users (user_id, email, username) values (
    @user_id::text,
    @email::text,
    @username::text
) returning *;
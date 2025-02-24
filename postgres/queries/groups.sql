-- name: GetAllGroups :many
select * from groups;

-- name: GetGroupById :one
select * from groups
where group_id = $1;

-- name: GetGroupsByUserId :many
select g.* from users u
inner join group_members gm on u.user_id = gm.user_id
inner join groups g on gm.group_id = g.group_id
where u.user_id = $1;

-- name: CreateGroup :one
insert into groups (group_id, name)
values ($1,$2)
returning *;

-- name: UpdateGroup :one
update groups
set name = $1
where group_id = $2
returning *;

-- name: DeleteGroup :exec
delete from groups
where group_id = $1;
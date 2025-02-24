-- name: CreateGroupMember :exec
insert into group_members (user_id, group_id)
values ($1, $2);

-- name: DeleteGroupMember :exec
delete from group_members
where user_id = $1 and group_id = $2;
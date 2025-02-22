-- name: FetchGroupsByUserId :many
select g.* from users u
inner join group_members gm on u.user_id = gm.user_id
inner join groups g on gm.group_id = g.group_id
where u.user_id = $1;
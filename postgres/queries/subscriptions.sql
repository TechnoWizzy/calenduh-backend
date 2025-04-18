-- name: GetAllSubscriptions :many
select * from subscriptions;

-- name: CreateSubscription :exec
insert into subscriptions (user_id, calendar_id, invite_code)
values ($1, $2, $3);

-- name: DeleteSubscription :exec
delete from subscriptions
where user_id = $1 and calendar_id = $2;
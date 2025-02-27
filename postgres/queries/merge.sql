-- name: TransferCalendarOwnership :exec
update calendars
set user_id = $1
where calendar_id = $2;
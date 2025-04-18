begin;

alter table subscriptions
add column invite_code text;

commit;
begin;

alter table events
    add column first_notification integer,
    add column second_notification integer;

commit;
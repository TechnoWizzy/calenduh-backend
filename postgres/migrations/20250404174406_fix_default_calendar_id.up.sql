begin;

alter table users
    drop column default_calendar_id;

alter table users
    add column default_calendar_id text references calendars (calendar_id) on delete set null on update cascade;

commit;
begin;

alter table users
    add column default_calendar_enabled boolean not null default true;

alter table users
    drop column default_calendar_id;

alter table users
    add column default_calendar_id text references calendars (calendar_id) on delete cascade on update cascade;

commit;
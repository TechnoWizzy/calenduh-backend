begin;

alter table users
    add column default_calendar_id text;

alter table users
    drop column profile_picture;

alter table users
    add column profile_picture text;

commit;
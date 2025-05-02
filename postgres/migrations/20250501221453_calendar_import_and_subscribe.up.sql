begin;

alter table calendars
add column is_imported boolean not null default false,
add column is_web_based boolean not null default false,
add column url text;

commit;
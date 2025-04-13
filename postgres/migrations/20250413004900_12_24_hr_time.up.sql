begin;

alter table users
    add column is_24_hour boolean not null default false;

commit;
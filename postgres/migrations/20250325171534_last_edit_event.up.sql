begin;

alter table events
    add column last_edited timestamp(3) not null default now();

commit;
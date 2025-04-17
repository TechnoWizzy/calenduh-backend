begin;

alter table calendars
    add column invite_code text unique not null default substring(md5(random()::text), 1, 6);

-- Commit migration
commit;
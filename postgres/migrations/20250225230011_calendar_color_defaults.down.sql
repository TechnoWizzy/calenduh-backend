begin;

-- Remove NOT NULL constraint and default value
alter table calendars
    alter column color drop not null,
    alter column color drop default;

-- Reset color to NULL where it was set to '#000000'
update calendars
set color = null
where color = '#000000';

commit;

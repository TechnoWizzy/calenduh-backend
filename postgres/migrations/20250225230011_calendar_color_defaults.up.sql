begin;

-- Fix null columns
update calendars
set color = '#000000'
where color is null;

-- Add default and constraint
alter table calendars
    alter column color set not null,
    alter column color set default '#000000';

commit;
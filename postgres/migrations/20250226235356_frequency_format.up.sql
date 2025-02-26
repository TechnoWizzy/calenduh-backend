begin;

-- Add default and constraint
alter table events
    alter column frequency type text;

commit;
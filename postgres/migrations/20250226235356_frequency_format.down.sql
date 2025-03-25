begin;

-- Add default and constraint
alter table events
    drop column frequency;

alter table events
    add column frequency int;

commit;
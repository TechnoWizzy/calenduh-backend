begin;

update events set all_day = false
where all_day is null;

alter table events
alter column all_day set not null;

commit;
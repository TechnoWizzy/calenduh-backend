begin;

alter table events
    drop column last_edited;

commit;
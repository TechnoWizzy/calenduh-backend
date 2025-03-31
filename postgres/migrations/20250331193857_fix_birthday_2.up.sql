begin;

alter table users
    drop column birthday;

alter table users
    add column birthday date;

commit;
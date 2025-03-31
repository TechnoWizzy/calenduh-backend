begin;

alter table users
    drop column birthday;

alter table users
    add column birthday timestamp(3);

commit;
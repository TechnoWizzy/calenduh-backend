create type session_type as enum ('APPLE', 'GOOGLE', 'DISCORD');

create table users (
    user_id text primary key,
    email text unique not null,
    username text not null
);

create table sessions (
    session_id text primary key,
    user_id text not null references users(user_id) on delete cascade on update cascade,
    type session_type not null,
    access_token text,
    refresh_token text,
    expires_on timestamp(3) not null
);

create table groups (
    group_id text primary key,
    name text not null
);

create table group_members (
    group_id text not null references groups(group_id) on delete cascade on update cascade,
    user_id text not null references users(user_id) on delete cascade on update cascade,
    primary key (group_id, user_id)
);

create table calendars (
    calendar_id text primary key,
    user_id text references users(user_id) on delete cascade on update cascade,
    group_id text references groups(group_id) on delete cascade on update cascade,
    title text not null,
    is_public boolean not null default false
);

create table events (
                        event_id text primary key,
                        calendar_id text not null references calendars(calendar_id) on delete cascade on update cascade,
                        name text not null,
                        location text,
                        description text,
                        notification text,
                        frequency int,
                        priority int,
                        start_time timestamp(3),
                        end_time timestamp(3)
);

create table subscriptions (
    user_id text not null references users(user_id) on delete cascade on update cascade,
    calendar_id text not null references calendars(calendar_id) on delete cascade on update cascade,
    primary key (user_id, calendar_id)
);
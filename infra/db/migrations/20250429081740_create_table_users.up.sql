create table if not exists users (
    id uuid not null primary key,
    fullname varchar not null,
    dob date not null,
    gender varchar not null,
    role varchar not null,
    is_confirmed boolean not null,
    created_at timestamp default current_timestamp,
    updated_at timestamp default current_timestamp
);

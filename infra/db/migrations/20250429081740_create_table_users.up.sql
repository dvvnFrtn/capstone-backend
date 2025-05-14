create table if not exists users (
    id uuid not null primary key,
    fullname varchar not null,
    email varchar unique,
    phone varchar unique,
    address varchar,
    role varchar not null,
    created_at timestamp default current_timestamp,
    updated_at timestamp default current_timestamp
);

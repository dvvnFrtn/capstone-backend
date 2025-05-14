create table if not exists communities (
    id uuid not null primary key,
    rt_number int not null,
    rw_number int not null,
    subdistrict varchar not null,
    district varchar not null,
    city varchar not null,
    province varchar not null,
    created_at timestamp default current_timestamp,
    updated_at timestamp default current_timestamp
);

alter table users
    add column community_id uuid not null;

alter table users
    add constraint fk_community
        foreign key(community_id) references communities(id);

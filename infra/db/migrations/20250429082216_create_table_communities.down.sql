alter table users
    drop constraint fk_community;

alter table users
    drop column community_id;

drop table if exists communities;

-- name: InsertUser :one
insert into users (
    id,
    community_id,
    fullname,
    role,
    is_confirmed
) values ($1, $2, $3, $4, $5)
returning id;

-- name: InsertCommunity :one
insert into communities (
    id,
    rt_number,
    rw_number,
    subdistrict,
    district,
    city,
    province,
    is_confirmed
) values ($1, $2, $3, $4, $5, $6, $7, $8)
returning id;

-- name: FindUserByID :one
select
  u.*,
  c.*
from users u
inner join communities c on c.id = u.community_id
where u.id = $1;

-- name: UpdateUserStatus :exec
update users
set is_confirmed = $1
where id = $2;

-- name: UpdateCommunityStatus :exec
update communities
set is_confirmed = $1
where id = (
  select u.community_id
  from users u
  where u.id = $2
);

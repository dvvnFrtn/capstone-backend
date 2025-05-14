-- name: InsertUser :one
insert into users (
    id,
    community_id,
    fullname,
    email,
    phone,
    address,
    role
) values ($1, $2, $3, $4, $5, $6, $7)
returning id;

-- name: UpdateUser :one
update users
set
  fullname = coalesce(sqlc.narg('fullname')::text, fullname),
  email = coalesce(sqlc.narg('email')::text, email),
  phone = coalesce(sqlc.narg('phone')::text, phone),
  address = coalesce(sqlc.narg('address'), address),
  role = coalesce(sqlc.narg('role')::text, role)
where
  id = sqlc.arg('id')::uuid
returning id;

-- name: DeleteUser :exec
delete from users
where id = $1;

-- name: InsertCommunity :one
insert into communities (
    id,
    rt_number,
    rw_number,
    subdistrict,
    district,
    city,
    province
) values ($1, $2, $3, $4, $5, $6, $7)
returning id;

-- name: FindUserByID :one
select
  u.*,
  c.*
from users u
inner join communities c on c.id = u.community_id
where
  (
    sqlc.narg('id')::uuid is null or 
    u.id = sqlc.narg('id')::uuid
  )
  and
  (
    sqlc.narg('email')::text is null or 
    u.email = sqlc.narg('email')::text
  )
  and
  (
    sqlc.narg('phone')::text is null or 
    u.phone = sqlc.narg('phone')::text
  )
  and
  (
    sqlc.narg('community_id')::uuid is null or
    u.community_id = sqlc.narg('community_id')::uuid
  );

-- name: FindUserByCommunityID :many
select
  u.*,
  c.*
from users u
inner join communities c on c.id = u.community_id
where
  u.community_id = $1;

-- name: IsEmailExists :one
select exists(
  select 1 from users where email = $1
);

-- name: IsPhoneExists :one
select exists(
  select 1 from users where phone = $1
);

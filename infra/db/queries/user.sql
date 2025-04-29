-- name: InsertUser :one
insert into users (
    id,
    community_id,
    fullname,
    dob,
    gender,
    role,
    is_confirmed
) values ($1, $2, $3, $4, $5, $6, $7)
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
select u.* from users u where u.id = $1;

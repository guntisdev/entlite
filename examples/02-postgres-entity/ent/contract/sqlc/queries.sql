-- Generate queries.sql
-- This file contains SQLC-compatible queries definitions

-- User CRUD operations

-- name: CreateUser :one
INSERT INTO "user" (
  email,
  name,
  age,
  password,
  score,
  uuid,
  is_admin,
  api_key,
  last_login_ms,
  created_at,
  updated_at
) VALUES (
  $1,
  $2,
  $3,
  $4,
  $5,
  $6,
  $7,
  $8,
  $9,
  $10,
  $11
) RETURNING ID;

-- name: GetUserByID :one
SELECT * FROM "user" WHERE ID = $1;

-- name: GetUserByEmail :one
SELECT * FROM "user" WHERE email = $1;

-- name: GetUserByNameAge :one
SELECT * FROM "user" WHERE name = $1 AND age = $2;

-- name: ListUserByName :many
SELECT * FROM "user" WHERE name = :name;

-- name: UpdateUser :one
UPDATE "user" SET
  email = :email,
  name = :name,
  age = :age,
  password = COALESCE(sqlc.narg('password'), password),
  score = COALESCE(sqlc.narg('score'), score),
  is_admin = :is_admin,
  api_key = COALESCE(sqlc.narg('api_key'), api_key),
  last_login_ms = :last_login_ms,
  updated_at = :updated_at
WHERE ID = :ID
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM "user" WHERE ID = $1;


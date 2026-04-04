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

-- name: GetUser :one
SELECT * FROM "user" WHERE ID = $1;

-- name: ListUser :many
SELECT * FROM "user" ORDER BY ID;

-- name: UpdateUser :one
UPDATE "user" SET
  email = sqlc.arg('email'),
  name = sqlc.arg('name'),
  age = sqlc.arg('age'),
  password = COALESCE(sqlc.narg('password'), password),
  score = COALESCE(sqlc.narg('score'), score),
  is_admin = sqlc.arg('is_admin'),
  api_key = COALESCE(sqlc.narg('api_key'), api_key),
  last_login_ms = sqlc.arg('last_login_ms'),
  updated_at = sqlc.arg('updated_at')
WHERE ID = sqlc.arg('ID')
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM "user" WHERE ID = $1;


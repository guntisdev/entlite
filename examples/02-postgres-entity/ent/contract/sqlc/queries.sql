-- Generate queries.sql
-- This file contains SQLC-compatible queries definitions

-- User CRUD operations

-- name: CreateUser :one
INSERT INTO "user" (
  email,
  name,
  age,
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
  $10
) RETURNING ID;

-- name: GetUser :one
SELECT * FROM "user" WHERE ID = $1;

-- name: ListUser :many
SELECT * FROM "user" ORDER BY ID;

-- name: UpdateUser :one
UPDATE "user" SET
  email = $1,
  name = $2,
  age = $3,
  score = $4,
  is_admin = $5,
  last_login_ms = $6,
  updated_at = $7
WHERE ID = $8
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM "user" WHERE ID = $1;


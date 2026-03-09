-- Generate queries.sql
-- This file contains SQLC-compatible queries definitions

-- User CRUD operations

-- name: CreateUser :one
INSERT INTO "user" (
  email,
  name,
  age,
  uuid,
  is_admin,
  created_at,
  updated_at
) VALUES (
  $1,
  $2,
  $3,
  $4,
  $5,
  $6,
  $7
) RETURNING id;

-- name: GetUser :one
SELECT * FROM "user" WHERE id = $1;

-- name: ListUser :many
SELECT * FROM "user" ORDER BY id;

-- name: UpdateUser :one
UPDATE "user" SET
  email = $1,
  name = $2,
  age = $3,
  is_admin = $4,
  updated_at = $5
WHERE id = $6
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM "user" WHERE id = $1;


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
  ?,
  ?,
  ?,
  ?,
  ?,
  ?,
  ?,
  ?,
  ?,
  ?
) RETURNING id;

-- name: GetUser :one
SELECT * FROM "user" WHERE id = ?;

-- name: ListUser :many
SELECT * FROM "user" ORDER BY id;

-- name: UpdateUser :one
UPDATE "user" SET
  email = ?,
  name = ?,
  age = ?,
  score = ?,
  is_admin = ?,
  last_login_ms = ?,
  updated_at = ?
WHERE id = ?
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM "user" WHERE id = ?;


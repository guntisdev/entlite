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
  ?,
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
) RETURNING ID;

-- name: GetUser :one
SELECT * FROM "user" WHERE ID = ?;

-- name: ListUser :many
SELECT * FROM "user" ORDER BY ID;

-- name: UpdateUser :one
UPDATE "user" SET
  email = ?,
  name = ?,
  age = ?,
  password = ?,
  score = ?,
  is_admin = ?,
  last_login_ms = ?,
  updated_at = ?
WHERE ID = ?
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM "user" WHERE ID = ?;


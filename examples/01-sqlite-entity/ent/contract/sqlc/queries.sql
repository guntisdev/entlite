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

-- name: GetUserByID :one
SELECT * FROM "user" WHERE ID = ?;

-- name: GetUserByEmail :one
SELECT * FROM "user" WHERE email = ?;

-- name: GetUserByNameAge :one
SELECT * FROM "user" WHERE name = ? AND age = ?;

-- name: ListUserByAge :many
SELECT * FROM "user" WHERE age = ?;

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
DELETE FROM "user" WHERE ID = ?;


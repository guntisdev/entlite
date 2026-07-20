-- Generate queries.sql
-- This file contains SQLC-compatible queries definitions

-- User CRUD operations

-- name: CreateUser :one
INSERT INTO "user" (
  email,
  name,
  age,
  password,
  api_key,
  is_active,
  login_count,
  rating,
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
) RETURNING ID;

-- name: CreateBulkUser :one
INSERT INTO "user" (
  email,
  name,
  age,
  password,
  api_key,
  is_active,
  login_count,
  rating,
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
) RETURNING ID;

-- name: GetUserByID :one
SELECT * FROM "user" WHERE ID = ?;

-- name: GetUserByEmail :one
SELECT * FROM "user" WHERE email = ?;

-- name: ListAllUser :many
SELECT * FROM "user";

-- name: ListUserByIsActive :many
SELECT * FROM "user" WHERE is_active = :is_active;

-- name: ListUserFilterByAgeName :many
SELECT * FROM "user" WHERE age BETWEEN :min_age AND :max_age AND name LIKE :name;

-- name: UpdateUser :one
UPDATE "user" SET
  email = :email,
  name = :name,
  age = :age,
  password = COALESCE(sqlc.narg('password'), password),
  is_active = COALESCE(sqlc.narg('is_active'), is_active),
  login_count = COALESCE(sqlc.narg('login_count'), login_count),
  rating = COALESCE(sqlc.narg('rating'), rating),
  updated_at = :updated_at
WHERE ID = :ID
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM "user" WHERE ID = ?;

-- name: DeleteAllUser :exec
DELETE FROM "user";


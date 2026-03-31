-- Generate queries.sql
-- This file contains SQLC-compatible queries definitions

-- User CRUD operations

-- name: CreateUser :exec
INSERT INTO `user` (
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
);
-- name: GetUser :one
SELECT * FROM `user` WHERE ID = ?;

-- name: ListUser :many
SELECT * FROM `user` ORDER BY ID;

-- name: UpdateUser :exec
UPDATE `user` SET
  email = ?,
  name = ?,
  age = ?,
  password = COALESCE(sqlc.narg('password'), password),
  score = ?,
  is_admin = ?,
  api_key = ?,
  last_login_ms = ?,
  updated_at = ?
WHERE ID = ?;

-- name: DeleteUser :exec
DELETE FROM `user` WHERE ID = ?;


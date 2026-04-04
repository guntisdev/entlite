-- Generate queries.sql
-- This file contains SQLC-compatible queries definitions

-- User CRUD operations

-- name: CreateUser :execlastid
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
  email = sqlc.arg('email'),
  name = sqlc.arg('name'),
  age = sqlc.arg('age'),
  password = COALESCE(sqlc.narg('password'), password),
  score = COALESCE(sqlc.narg('score'), score),
  is_admin = sqlc.arg('is_admin'),
  api_key = COALESCE(sqlc.narg('api_key'), api_key),
  last_login_ms = sqlc.arg('last_login_ms'),
  updated_at = sqlc.arg('updated_at')
WHERE ID = sqlc.arg('ID');

-- name: DeleteUser :exec
DELETE FROM `user` WHERE ID = ?;


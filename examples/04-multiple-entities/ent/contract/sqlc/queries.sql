-- Generate queries.sql
-- This file contains SQLC-compatible queries definitions

-- Post CRUD operations

-- name: CreatePost :one
INSERT INTO "post" (
  title,
  content,
  published
) VALUES (
  $1,
  $2,
  $3
) RETURNING ID;

-- name: GetPost :one
SELECT * FROM "post" WHERE ID = $1;

-- name: ListPost :many
SELECT * FROM "post" ORDER BY ID;

-- name: UpdatePost :one
UPDATE "post" SET
  title = $1,
  content = $2,
  published = $3
WHERE ID = $4
RETURNING *;

-- name: DeletePost :exec
DELETE FROM "post" WHERE ID = $1;

-- User CRUD operations

-- name: CreateUser :one
INSERT INTO "user" (
  email,
  name
) VALUES (
  $1,
  $2
) RETURNING ID;

-- name: GetUser :one
SELECT * FROM "user" WHERE ID = $1;

-- name: ListUser :many
SELECT * FROM "user" ORDER BY ID;

-- name: UpdateUser :one
UPDATE "user" SET
  email = $1,
  name = $2
WHERE ID = $3
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM "user" WHERE ID = $1;


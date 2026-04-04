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
  title = sqlc.arg('title'),
  content = sqlc.arg('content'),
  published = COALESCE(sqlc.narg('published'), published)
WHERE ID = sqlc.arg('ID')
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
  email = sqlc.arg('email'),
  name = sqlc.arg('name')
WHERE ID = sqlc.arg('ID')
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM "user" WHERE ID = $1;


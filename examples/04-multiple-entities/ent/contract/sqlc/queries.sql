-- Generate queries.sql
-- This file contains SQLC-compatible queries definitions

-- Post CRUD operations

-- name: GetPostByID :one
SELECT * FROM "post" WHERE ID = $1;

-- User CRUD operations


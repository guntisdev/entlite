-- Generate queries.sql
-- This file contains SQLC-compatible queries definitions

-- Article CRUD operations

-- name: CreateArticle :one
INSERT INTO "article" (
  id,
  slug,
  title,
  author,
  subtitle,
  reading_minutes,
  last_viewed_ms,
  rating,
  cover_image,
  published_at,
  is_featured,
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
  ?,
  ?,
  ?
) RETURNING id;

-- name: GetArticleByID :one
SELECT * FROM "article" WHERE ID = ?;

-- name: GetArticleBySlug :one
SELECT * FROM "article" WHERE slug = ?;

-- name: ListArticleByAuthor :many
SELECT * FROM "article" WHERE author = :author;

-- name: ListArticleFilterByAuthorIsFeaturedPublishedAtTitle :many
SELECT * FROM "article" WHERE author = :author AND is_featured = :is_featured AND published_at BETWEEN :min_published_at AND :max_published_at AND title LIKE :title;

-- name: UpdateArticle :one
UPDATE "article" SET
  slug = :slug,
  title = :title,
  author = :author,
  subtitle = :subtitle,
  reading_minutes = :reading_minutes,
  last_viewed_ms = :last_viewed_ms,
  rating = :rating,
  cover_image = :cover_image,
  published_at = :published_at,
  is_featured = COALESCE(sqlc.narg('is_featured'), is_featured),
  updated_at = :updated_at
WHERE id = :id
RETURNING *;

-- name: DeleteArticle :exec
DELETE FROM "article" WHERE id = ?;


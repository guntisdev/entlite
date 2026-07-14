-- Generated schema.sql
-- This file contains table definitions for all entities

-- article table
CREATE TABLE "article"(
  id TEXT PRIMARY KEY,
  slug TEXT UNIQUE NOT NULL,
  title TEXT NOT NULL,
  author TEXT NOT NULL,
  subtitle TEXT,
  reading_minutes INTEGER,
  last_viewed_ms INTEGER,
  rating REAL,
  cover_image BLOB,
  published_at DATETIME,
  is_featured INTEGER DEFAULT false NOT NULL,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL
);


-- Generated schema.sql
-- This file contains table definitions for all entities

-- user table
CREATE TABLE "user"(
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  email TEXT UNIQUE NOT NULL,
  name TEXT NOT NULL,
  age INTEGER,
  score REAL DEFAULT 0 NOT NULL,
  uuid TEXT NOT NULL,
  is_admin INTEGER NOT NULL,
  api_key BLOB NOT NULL,
  last_login_ms INTEGER NOT NULL,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL
);


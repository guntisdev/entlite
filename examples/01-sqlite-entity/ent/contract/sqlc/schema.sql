-- Generated schema.sql
-- This file contains table definitions for all entities

-- user table
CREATE TABLE "user"(
  ID INTEGER PRIMARY KEY AUTOINCREMENT,
  email TEXT UNIQUE NOT NULL,
  name TEXT NOT NULL,
  age INTEGER,
  password TEXT NOT NULL,
  api_key BLOB NOT NULL,
  is_active INTEGER DEFAULT true NOT NULL,
  login_count INTEGER DEFAULT 0 NOT NULL,
  rating REAL DEFAULT 0 NOT NULL,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL
);


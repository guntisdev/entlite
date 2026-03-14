-- Generated schema.sql
-- This file contains table definitions for all entities

-- user table
CREATE TABLE "user"(
  id BIGSERIAL PRIMARY KEY,
  email TEXT UNIQUE NOT NULL,
  name TEXT NOT NULL,
  age BIGINT,
  uuid TEXT NOT NULL,
  is_admin BOOLEAN NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL
);


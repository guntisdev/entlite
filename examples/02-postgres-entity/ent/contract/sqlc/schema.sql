-- Generated schema.sql
-- This file contains table definitions for all entities

-- user table
CREATE TABLE "user"(
  ID SERIAL PRIMARY KEY,
  email TEXT UNIQUE NOT NULL,
  name TEXT NOT NULL,
  age INT,
  score DOUBLE PRECISION DEFAULT 0 NOT NULL,
  uuid TEXT NOT NULL,
  is_admin BOOLEAN NOT NULL,
  api_key BYTEA NOT NULL,
  last_login_ms BIGINT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL
);


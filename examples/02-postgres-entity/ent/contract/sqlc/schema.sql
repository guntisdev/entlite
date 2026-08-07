-- Generated schema.sql
-- This file contains table definitions for all entities

-- user table
CREATE TABLE "user"(
  ID SERIAL PRIMARY KEY,
  email TEXT UNIQUE NOT NULL,
  -- Full name, e.g. "Jane Doe"
  name TEXT NOT NULL,
  age INT,
  password TEXT NOT NULL,
  api_key BYTEA NOT NULL,
  is_active BOOLEAN DEFAULT true NOT NULL,
  login_count BIGINT DEFAULT 0 NOT NULL,
  rating DOUBLE PRECISION DEFAULT 0 NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX "idx_user_age_is_active" ON "user" (age, is_active);
CREATE INDEX "idx_user_is_active_created_at" ON "user" (is_active, created_at DESC);
CREATE UNIQUE INDEX "idx_user_name_email" ON "user" (name, email);
CREATE INDEX "idx_users_stats" ON "user" (login_count, rating);


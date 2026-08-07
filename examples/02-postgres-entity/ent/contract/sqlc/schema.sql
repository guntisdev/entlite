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
CREATE INDEX "idx_user_env_is_active" ON "user" (env, is_active);
CREATE INDEX "idx_user_country_env_created_at" ON "user" (country, env, created_at DESC);
CREATE UNIQUE INDEX "idx_user_tenant_id_email" ON "user" (tenant_id, email);
CREATE INDEX "idx_users_stats" ON "user" (login_count, rating);


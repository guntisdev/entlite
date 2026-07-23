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
  updated_at DATETIME NOT NULL,
  PRIMARY KEY (country, timestamp, env)
);
CREATE INDEX "idx_user_env_is_active" ON "user" (env, is_active);
CREATE INDEX "idx_user_country_env_created_at" ON "user" (country, env, created_at DESC);
CREATE UNIQUE INDEX "idx_user_tenant_id_email" ON "user" (tenant_id, email);
CREATE INDEX "idx_users_stats" ON "user" (login_count, rating);


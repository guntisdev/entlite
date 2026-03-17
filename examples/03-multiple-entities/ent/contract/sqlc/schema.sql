-- Generated schema.sql
-- This file contains table definitions for all entities

-- post table
CREATE TABLE "post"(
  id SERIAL PRIMARY KEY,
  title TEXT NOT NULL,
  content TEXT NOT NULL,
  published BOOLEAN DEFAULT false NOT NULL
);

-- user table
CREATE TABLE "user"(
  id SERIAL PRIMARY KEY,
  email TEXT UNIQUE NOT NULL,
  name TEXT NOT NULL
);


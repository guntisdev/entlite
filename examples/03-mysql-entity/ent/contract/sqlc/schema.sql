-- Generated schema.sql
-- This file contains table definitions for all entities

-- user table
CREATE TABLE `user`(
  ID INT AUTO_INCREMENT PRIMARY KEY,
  email TEXT UNIQUE NOT NULL,
  name TEXT NOT NULL,
  age INT,
  password TEXT NOT NULL,
  score DOUBLE DEFAULT 4.2 NOT NULL,
  uuid TEXT NOT NULL,
  is_admin TINYINT(1) NOT NULL,
  api_key BLOB NOT NULL,
  last_login_ms BIGINT NOT NULL,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL
);


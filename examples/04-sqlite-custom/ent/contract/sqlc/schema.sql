-- Generated schema.sql
-- This file contains table definitions for all entities

-- reading table
CREATE TABLE "reading"(
  ID INTEGER PRIMARY KEY AUTOINCREMENT,
  sensor_id INTEGER NOT NULL,
  value REAL NOT NULL,
  quality INTEGER,
  flagged INTEGER DEFAULT false NOT NULL,
  recorded_at DATETIME NOT NULL,
  created_at DATETIME NOT NULL
);

-- sensor table
CREATE TABLE "sensor"(
  ID INTEGER PRIMARY KEY AUTOINCREMENT,
  code TEXT UNIQUE NOT NULL,
  label TEXT NOT NULL,
  kind TEXT NOT NULL,
  unit TEXT NOT NULL,
  location TEXT,
  active INTEGER DEFAULT true NOT NULL,
  firmware TEXT DEFAULT '1.0.0' NOT NULL,
  sample_rate_ms INTEGER DEFAULT 1000 NOT NULL,
  installed_at DATETIME NOT NULL,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL
);


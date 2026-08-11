-- Generated schema.sql
-- This file contains table definitions for all entities

-- reading table
CREATE TABLE "reading"(
  ID INTEGER PRIMARY KEY AUTOINCREMENT,
  -- References sensor.ID
  sensor_id INTEGER NOT NULL,
  value REAL NOT NULL,
  -- Signal quality 0-100
  quality INTEGER NOT NULL,
  -- Marked as anomalous by ingestion
  flagged INTEGER DEFAULT false NOT NULL,
  -- Device measurement time (client-supplied)
  recorded_at DATETIME NOT NULL,
  created_at DATETIME NOT NULL
);

-- sensor table
CREATE TABLE "sensor"(
  ID INTEGER PRIMARY KEY AUTOINCREMENT,
  -- External hardware identifier, e.g. TEMP-A1
  code TEXT UNIQUE NOT NULL,
  -- Human friendly name
  label TEXT NOT NULL,
  -- temperature | humidity | pressure | motion
  kind TEXT NOT NULL,
  -- Measurement unit, e.g. celsius
  unit TEXT NOT NULL,
  location TEXT,
  active INTEGER DEFAULT true NOT NULL,
  firmware TEXT DEFAULT '1.0.0' NOT NULL,
  -- Sampling interval in milliseconds
  sample_rate_ms INTEGER DEFAULT 1000 NOT NULL,
  -- When the device was physically installed (client-supplied)
  installed_at DATETIME NOT NULL,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL
);


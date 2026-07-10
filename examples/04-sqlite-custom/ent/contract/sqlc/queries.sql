-- Generate queries.sql
-- This file contains SQLC-compatible queries definitions

-- Reading CRUD operations

-- name: CreateReading :one
INSERT INTO "reading" (
  sensor_id,
  value,
  quality,
  flagged,
  recorded_at,
  created_at
) VALUES (
  ?,
  ?,
  ?,
  ?,
  ?,
  ?
) RETURNING ID;

-- name: GetReadingByID :one
SELECT * FROM "reading" WHERE ID = ?;

-- name: ListReadingBySensorId :many
SELECT * FROM "reading" WHERE sensor_id = :sensor_id;

-- name: ListReadingFilterBySensorIdRecordedAtFlagged :many
SELECT * FROM "reading" WHERE sensor_id = :sensor_id AND recorded_at BETWEEN :min_recorded_at AND :max_recorded_at AND flagged = :flagged;

-- name: UpdateReading :one
UPDATE "reading" SET
  sensor_id = :sensor_id,
  value = :value,
  quality = :quality,
  flagged = COALESCE(sqlc.narg('flagged'), flagged),
  recorded_at = COALESCE(sqlc.narg('recorded_at'), recorded_at)
WHERE ID = :ID
RETURNING *;

-- name: DeleteReading :exec
DELETE FROM "reading" WHERE ID = ?;

-- Sensor CRUD operations

-- name: CreateSensor :one
INSERT INTO "sensor" (
  code,
  label,
  kind,
  unit,
  location,
  active,
  firmware,
  sample_rate_ms,
  installed_at,
  created_at,
  updated_at
) VALUES (
  ?,
  ?,
  ?,
  ?,
  ?,
  ?,
  ?,
  ?,
  ?,
  ?,
  ?
) RETURNING ID;

-- name: GetSensorByID :one
SELECT * FROM "sensor" WHERE ID = ?;

-- name: GetSensorByCode :one
SELECT * FROM "sensor" WHERE code = ?;

-- name: ListSensorFilterByLabelKindActive :many
SELECT * FROM "sensor" WHERE label LIKE :label AND kind = :kind AND active = :active;

-- name: UpdateSensor :one
UPDATE "sensor" SET
  code = :code,
  label = :label,
  kind = :kind,
  unit = :unit,
  location = :location,
  active = COALESCE(sqlc.narg('active'), active),
  firmware = COALESCE(sqlc.narg('firmware'), firmware),
  sample_rate_ms = COALESCE(sqlc.narg('sample_rate_ms'), sample_rate_ms),
  updated_at = :updated_at
WHERE ID = :ID
RETURNING *;

-- name: DeleteSensor :exec
DELETE FROM "sensor" WHERE ID = ?;


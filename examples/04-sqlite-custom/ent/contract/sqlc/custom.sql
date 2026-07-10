-- custom.sql
--
-- Hand-written queries that live ALONGSIDE the DSL-generated queries.sql.
-- entlite never touches this file, so anything here survives regeneration.
-- These are queries the DSL query builder cannot express: cross-table JOINs,
-- aggregates, and bulk maintenance operations.
--
-- The tables ("sensor", "reading") are defined in the generated schema.sql.

-- Aggregate value statistics for a single sensor over a time window.
-- name: GetSensorReadingStats :one
SELECT
  COUNT(*)   AS reading_count,
  AVG(value) AS avg_value,
  MIN(value) AS min_value,
  MAX(value) AS max_value
FROM "reading"
WHERE sensor_id = :sensor_id
  AND recorded_at BETWEEN :from_ts AND :to_ts;

-- Every active sensor together with its most recent reading (LEFT JOIN, so
-- sensors that have never reported still show up with NULL latest values).
-- sqlc.embed(s) maps the joined sensor columns back to the full Sensor struct.
-- name: ListSensorsWithLatestReading :many
SELECT
  sqlc.embed(s),
  r.value       AS latest_value,
  r.recorded_at AS latest_recorded_at
FROM "sensor" s
LEFT JOIN "reading" r
  ON r.ID = (
    SELECT r2.ID
    FROM "reading" r2
    WHERE r2.sensor_id = s.ID
    ORDER BY r2.recorded_at DESC
    LIMIT 1
  )
WHERE s.active = 1
ORDER BY s.code
LIMIT :limit OFFSET :offset;

-- Retention: drop readings older than a cutoff. Returns nothing.
-- name: PruneReadingsOlderThan :execrows
DELETE FROM "reading" WHERE recorded_at < :cutoff;

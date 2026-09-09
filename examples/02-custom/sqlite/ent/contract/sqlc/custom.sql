-- Hand-written queries the DSL cannot express. entlite never writes this file.
-- The tables are defined in the generated schema.sql.

-- Aggregate value statistics for a single sensor over a time window.
-- name: GetSensorReadingStats :one
SELECT
  COUNT(*)   AS reading_count,
  AVG(value) AS avg_value,
  MIN(value) AS min_value,
  MAX(value) AS max_value
FROM "sensor_reading"
-- >= / <= instead of BETWEEN: sqlc drops the bounds from the params struct
-- unless the range is the first filter.
WHERE sensor_id = :sensor_id
  AND recorded_at >= :from_ts AND recorded_at <= :to_ts;

-- Every active sensor with its most recent reading. LEFT JOIN, so a sensor that
-- never reported comes back with NULL latest values.
-- sqlc.embed(s) maps the joined columns back to the Sensor struct.
-- name: ListSensorsWithLatestReading :many
SELECT
  sqlc.embed(s),
  r.value       AS latest_value,
  r.recorded_at AS latest_recorded_at
FROM "sensor" s
LEFT JOIN "sensor_reading" r
  ON r.id = (
    SELECT r2.id
    FROM "sensor_reading" r2
    WHERE r2.sensor_id = s.id
    ORDER BY r2.recorded_at DESC
    LIMIT 1
  )
WHERE s.active = 1
ORDER BY s.code
LIMIT :limit OFFSET :offset;

-- Drops readings older than a cutoff.
-- name: PruneReadingsOlderThan :execrows
DELETE FROM "sensor_reading" WHERE recorded_at < :cutoff;

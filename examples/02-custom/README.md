# 02-custom 

Demonstrates that a developer can add functionality **outside** the entlite DSL
and still have it compile into typed Go/TS alongside the generated code.

## Entities (DSL)

Defined in [`ent/schema`](ent/schema):

- **Sensor** — a physical device in the field (`code`, `label`, `kind`, `unit`,
  `active`, `firmware`, `sample_rate_ms`, timestamps), plus `latest_value` with
  `permissions.Virtual`: a proto-only field with no column, filled at the API
  layer from the custom `ListSensorsWithLatestReading` query below.
- **Reading** — a measurement emitted by a sensor (`sensor_id`, `value`,
  `quality`, `flagged`, `recorded_at`).

## Custom additions (hand-written)

These files live next to the generated ones. entlite only ever writes the fixed
filenames above, so these survive regeneration:

- [`contract/sqlc/custom.sql`](ent/contract/sqlc/custom.sql) — queries the DSL
  cannot express: a cross-table `LEFT JOIN` with a correlated subquery
  (`ListSensorsWithLatestReading`, using `sqlc.embed`), an aggregate
  (`GetSensorReadingStats`), and a bulk retention delete
  (`PruneReadingsOlderThan`).
- [`contract/proto/custom.proto`](ent/contract/proto/custom.proto) — a
  hand-written `SensorAnalyticsService` in the same `entlite` package that
  imports `schema.proto` and reuses the generated `Sensor` message.

## Server

[`server/server.go`](server/server.go) implements all three services against the
same generated types: `SensorService` and `ReadingService` from the DSL, and the
hand-written `SensorAnalyticsService` on top of `custom.sql`. `ListWithLatestReading`
is where both halves meet — the custom `LEFT JOIN` returns an embedded sensor row,
the generated converter turns it into the same `pb.Sensor` the CRUD service returns,
and the `permissions.Virtual` `latest_value` field is filled from the joined reading.

## Run

```bash
make run   # generates, bundles the frontend, then serves on :8080
```

Known gap: `ReadingService.FilterBySensorIdRecordedAtFlagged` fails at runtime.
The DSL's `filter.Range("recorded_at")` emits `recorded_at BETWEEN @min AND @max`,
and sqlc cannot infer the type of a DATETIME placeholder inside `BETWEEN`, so it
drops both bounds from the params struct while the query still binds them.
`custom.sql` works around this by spelling the same range as `>= AND <=`.

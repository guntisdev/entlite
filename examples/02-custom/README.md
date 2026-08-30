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
  `quality`, `flagged`, `recorded_at`). Declares its own key as
  `field.Int64("ID")` instead of letting entlite add the default int32 one:
  readings are high volume and int32 runs out at 2.1B rows. `sensor_id` stays
  `field.Int` because it points at Sensor's int32 `ID` — the key type is chosen
  per entity, and a foreign key follows the entity it references.

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

## int64 ID, end to end

The declared key type travels the whole stack. Proto gets `int64 ID`, and since
SQLite columns are already 64-bit the generated wrapper drops the narrowing
`IntConvert[int32, int64]` it emits for int32 keys — compare `GetReadingByID`
with `GetSensorByID` in [`gen/db/queries.sql.go`](ent/gen/db/queries.sql.go).

On the wire an int64 is JSON-encoded as a string, so a reading comes back as
`{"ID":"2", ...}` while a sensor is `{"ID":2, ...}`. In TypeScript protobuf-es
maps it to `bigint`, which is why reading IDs in the frontend go through
`bigIntInput()` rather than `numberInput()`.

## Run

```bash
make run   # generates, bundles the frontend, then serves on :8080
```

Known gap: `ReadingService.FilterBySensorIdRecordedAtFlagged` fails at runtime.
The DSL's `filter.Range("recorded_at")` emits `recorded_at BETWEEN @min AND @max`,
and sqlc cannot infer the type of a DATETIME placeholder inside `BETWEEN`, so it
drops both bounds from the params struct while the query still binds them.
`custom.sql` works around this by spelling the same range as `>= AND <=`.

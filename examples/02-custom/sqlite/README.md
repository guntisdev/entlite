# 02-custom / sqlite

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

From these, entlite generates the contract:
`contract/proto/schema.proto`, `contract/sqlc/schema.sql`, `contract/sqlc/queries.sql`.

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

Wiring:

- `sqlc.yaml` lists both `queries.sql` and `custom.sql`, so `sqlc generate`
  compiles them together; `entlite sqlc-wrap` then wraps every generated query
  (custom included) into typed helpers.
- `buf.yaml` compiles the whole `contract/proto` directory, so `custom.proto`
  builds alongside `schema.proto` into `gen/pb` and `gen/ts`.

The goal: prove the generated and custom layers compile into one coherent set of
types, so extending beyond the DSL requires no escape hatch — just drop files in.

## Server

[`server/server.go`](server/server.go) implements all three services against the
same generated types: `SensorService` and `ReadingService` from the DSL, and the
hand-written `SensorAnalyticsService` on top of `custom.sql`. `ListWithLatestReading`
is where both halves meet — the custom `LEFT JOIN` returns an embedded sensor row,
the generated converter turns it into the same `pb.Sensor` the CRUD service returns,
and the `permissions.Virtual` `latest_value` field is filled from the joined reading.

[`main.go`](main.go) mounts all three on one Connect mux with both validation
layers (protovalidate + the generated `Validate()` interceptor).

```bash
make run   # generates, then serves on :8080
```

```bash
curl -X POST localhost:8080/entlite.SensorService/Create \
  -H "Content-Type: application/json" \
  -d '{"code":"TEMP-A1","label":"Lab temp","kind":"temperature","unit":"celsius","installed_at":"2026-01-01T00:00:00Z"}'

curl -X POST localhost:8080/entlite.SensorAnalyticsService/ListWithLatestReading \
  -H "Content-Type: application/json" -d '{"limit":10,"offset":0}'
```

Unlike `01-basic-entity`, this example has no web UI — it is exercised over HTTP
directly.

Known gap: `ReadingService.FilterBySensorIdRecordedAtFlagged` fails at runtime.
The DSL's `filter.Range("recorded_at")` emits `recorded_at BETWEEN @min AND @max`,
and sqlc cannot infer the type of a DATETIME placeholder inside `BETWEEN`, so it
drops both bounds from the params struct while the query still binds them.
`custom.sql` works around this by spelling the same range as `>= AND <=`.

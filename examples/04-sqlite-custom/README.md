# 04-sqlite-custom

Demonstrates that a developer can add functionality **outside** the entlite DSL
and still have it compile into typed Go/TS alongside the generated code.

## Entities (DSL)

Defined in [`ent/schema`](ent/schema):

- **Sensor** — a physical device in the field (`code`, `label`, `kind`, `unit`,
  `active`, `firmware`, `sample_rate_ms`, timestamps).
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

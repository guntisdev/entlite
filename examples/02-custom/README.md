# 02-custom

The DSL cannot express every query. This example adds hand-written SQL and proto
next to the generated files, and both halves compile into one typed API.

<!-- teaches:start -->
- Hand-written `custom.sql` and `custom.proto` live beside the generated files and survive regeneration
- A hand-written service that reuses a generated proto message
- Virtual fields: `Contracts(entlite.PROTO())` gives a field with no column, filled in by the server
- Choosing the key type per entity: `field.Int64("ID")` on a high volume table
- A foreign key follows the type of the entity it points at
- Query level `Contracts()`: a query that stays in the database layer and gets no rpc
- Two entities in one schema
<!-- teaches:end -->

## Entities

[`sqlite/ent/schema`](sqlite/ent/schema) has two:

- **Sensor** — a device in the field. Also declares `latest_value` as
  `Contracts(entlite.PROTO())`. That is a virtual field: no column, no place in
  any generated SQL, but it is in the proto message. The server fills it in.
- **Reading** — a measurement from a sensor. Declares its own key as
  `field.Int64("ID")` instead of taking the default int32. Readings are high
  volume and int32 stops at 2.1B rows. `sensor_id` stays `field.Int` because it
  points at Sensor's int32 key.

Reading's `Update()` is `Contracts(entlite.SQLC())`. A reading is a recorded
fact, so clients never edit it. The database query exists, the rpc does not.

## Hand-written files

entlite only writes fixed filenames, so anything else in these folders stays.

- [`sqlite/ent/contract/sqlc/custom.sql`](sqlite/ent/contract/sqlc/custom.sql) —
  a `LEFT JOIN` with a subquery using `sqlc.embed`, an aggregate, and a bulk
  delete.
- [`sqlite/ent/contract/proto/custom.proto`](sqlite/ent/contract/proto/custom.proto) —
  a `SensorAnalyticsService` in the same package. It imports `schema.proto` and
  reuses the generated `Sensor` message.

## Server

[`sqlite/server/server.go`](sqlite/server/server.go) serves all three services
from the same generated types. `ListWithLatestReading` is where the two halves
meet: the custom `LEFT JOIN` returns an embedded sensor row, the generated
converter turns it into the same `pb.Sensor` the CRUD service returns, and
`latest_value` is filled from the joined reading.

## int64 key, end to end

The key type travels the whole stack. Proto gets `int64 ID`. SQLite columns are
already 64-bit, so the wrapper drops the narrowing convert it emits for int32
keys — compare `GetReadingByID` with `GetSensorByID` in
[`sqlite/ent/gen/db/queries.sql.go`](sqlite/ent/gen/db/queries.sql.go).

On the wire an int64 is JSON encoded as a string, so a reading is
`{"ID":"2", ...}` and a sensor is `{"ID":2, ...}`. In TypeScript it is a
`bigint`, which is why reading IDs in the frontend use `bigIntInput()` and not
`numberInput()`.

## Run

```bash
cd sqlite
make run     # serves on :8080
```

Known gap: `ReadingService.FilterBySensorIdRecordedAtFlagged` fails at runtime.
`filter.Range("recorded_at")` emits `recorded_at BETWEEN @min AND @max`, and
sqlc cannot infer the type of a DATETIME placeholder inside `BETWEEN`. It drops
both bounds from the params struct while the query still binds them.
`custom.sql` works around this by writing the range as `>=` and `<=`.

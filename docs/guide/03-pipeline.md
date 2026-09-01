# Pipeline

One schema in, four outputs out. `go generate ./...` inside `ent/` runs six
steps in a fixed order.

```
ent/schema/*.go                     you write this
      │
      │  1. entlite gen .
      ▼
ent/contract/sqlc/schema.sql        table definitions
ent/contract/sqlc/queries.sql       one query per dsl query
ent/contract/proto/schema.proto     messages and services
      │
      │  2. sqlc generate      3. buf generate
      ▼                             ▼
ent/gen/db/internal/            ent/gen/pb/    ent/gen/ts/
      │                             │
      │  4. entlite sqlc-wrap  5. entlite proto-validate
      ▼                             ▼
ent/gen/db/                     ent/gen/pb/proto_validate.go
```

## The steps

| Step | Command | Writes |
| --- | --- | --- |
| 1 | `entlite gen .` | the sql and proto contracts |
| 2 | `sqlc generate` | raw database code in `gen/db/internal` |
| 3 | `buf generate` | proto messages, connect service, TypeScript |
| 4 | `entlite sqlc-wrap` | the typed wrapper in `gen/db` |
| 5 | `entlite proto-validate` | `Validate()` methods for the proto types |

Steps 1, 4 and 5 are entlite. Steps 2 and 3 are the outside tools, driven by
`sqlc.yaml` and `buf.gen.yaml`.

## Why wrap sqlc

sqlc returns its own row types, with `sql.NullString` for nullable columns.
Step 4 wraps them: pointers instead of null types, `time.Time` instead of the
driver type, and converters between the database row and the proto message. The
raw output stays in `gen/db/internal`, so you never import it.

## Why a separate validate step

`Validate()` on a field is a Go func, and proto has no way to carry it. Step 5
writes a `Validate()` method per request message plus a connect interceptor that
calls it. Rules protovalidate can express, such as `required`, land in the proto
file in step 1 instead, so a request passes two independent checks.

## Regenerating

The steps are safe to rerun. Hand written files in `contract/` and everything
outside `ent/` are never touched. If the generated output changes, it is because
the schema changed.

# 03-optional

What `Optional()` does to each type, and how an optional field differs from an
optional filter.

<!-- teaches:start -->
- `Optional()` on every type that supports it, one field each
- `Bool` cannot be optional — use `Default(false)` instead
- An optional field becomes `optional` in proto and a nullable column in SQL
- `filter.Eq("is_featured").Optional()` — a required field used as an optional filter
- A string uuid primary key: `Immutable()`, `DefaultFunc()` and `PROTO().ReadOnly()`
- Server generated fields drop out of create and update requests
<!-- teaches:end -->

## Entity

[`sqlite/ent/schema/article.go`](sqlite/ent/schema/article.go) is one `Article`,
grouped into four blocks.

**The key.** `field.String("id")` with `DefaultFunc(logic.NewUUID)`. It is
`Immutable()` and `PROTO().ReadOnly()`, so the server mints it and clients never
send it. Look at `CreateArticleRequest` in the generated
`schema.proto` — the field is not there.

The column is `TEXT PRIMARY KEY` on sqlite and postgresql, and
`VARCHAR(36) PRIMARY KEY` on mysql. The uuid is generated in Go, not by the
database.

**Optional fields.** One per type, so you can compare them side by side:

| Field | Type | Proto | SQL (sqlite) |
|---|---|---|---|
| `subtitle` | String | `optional string` | `TEXT` |
| `reading_minutes` | Int | `optional int32` | `INTEGER` |
| `last_viewed_ms` | Int64 | `optional int64` | `INTEGER` |
| `rating` | Float | `optional double` | `REAL` |
| `cover_image` | Byte | `optional bytes` | `BLOB` |
| `published_at` | Time | `optional Timestamp` | `DATETIME` |
| `metadata` | JSON | `optional string` | `TEXT` |

A required field gets `NOT NULL` and a `required` validation rule. An optional
one gets neither.

**Bool.** `is_featured` is `Default(false)`. Bool has no `Optional()`. A missing
proto bool arrives as `false`, so there is no way to tell "not sent" from
"sent as false". A default says what happens instead.

## Optional field vs optional filter

Two different things:

- `field.String("subtitle").Optional()` — the value may be missing.
- `filter.Eq("is_featured").Optional()` — the filter may be missing.

`is_featured` is required on the entity, but the last query in `Queries()` uses
it as an optional filter. Send it and the rows are narrowed. Leave it out and
the filter is skipped. The same query does this with `Range` and `Search` too.

## Run

```bash
cd sqlite
make run     # serves on :8080
```

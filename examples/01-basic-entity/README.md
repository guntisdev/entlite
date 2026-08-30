# 01-basic-entity

One `User` entity, written out in every field type and every query type. Start here.

<!-- teaches:start -->
- Every field type: `String`, `Int`, `Int64`, `Float`, `Bool`, `Byte`, `Time`, `JSON`
- Field options: `Unique()`, `Optional()`, `Immutable()`, `Default()`, `DefaultFunc()`
- `Validate()` with your own Go function from the `logic` package
- Field level `Contracts()`: a password that clients write but never read, timestamps they read but never write
- Ready made queries: `DefaultCRUD()`, `CreateBulk()`, `ListAll()`, `DeleteAll()`
- Queries by field: `GetBy("email")`, `ListBy("is_active")`, and `Name()` to rename one
- Filters: `filter.Range()` and `filter.Search()`
- Indexes: multi column, `Desc()` sort order, `Unique()`, `Name()`, and a commented out `index.Primary()`
- The same schema on three dialects: sqlite, postgres, mysql
<!-- teaches:end -->

## Entity

[`ent/schema/user.go`](sqlite/ent/schema/user.go) is the same file in all three
dialect folders. Only `sqlc.yaml` and the generated SQL differ.

`Fields()` lists one field per type, so you can see what each one becomes in
`schema.sql` and in `schema.proto`. Two fields are worth a look:

- `password` is `SQLC()` plus `PROTO().WriteOnly()`. It is a column, and clients
  can send it, but it is not in any response message.
- `created_at` and `updated_at` are `PROTO().ReadOnly()`. The server sets them
  with `DefaultFunc(time.Now)`, so they are not in create or update requests.

`api_key` shows `DefaultFunc` on bytes. The value is generated in Go, not by the
database.

## Dialects

| Folder | Database | Notes |
|---|---|---|
| [`sqlite`](sqlite) | file on disk | nothing to start, fastest way to try it |
| [`postgres`](postgres) | docker | `make run` starts it and waits |
| [`mysql`](mysql) | docker | `make run` starts it and waits |

Compare the three `ent/contract/sqlc/schema.sql` files to see how types map. For
example `field.Float` becomes `REAL` on sqlite, `DOUBLE PRECISION` on postgres
and `DOUBLE` on mysql.

## Run

```bash
cd sqlite
make run     # generates types, bundles the frontend, serves on :8080
```

For postgres and mysql, `make down` stops the database and drops its data.

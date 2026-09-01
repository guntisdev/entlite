# Getting started

Create an entity, generate the code, run the server. About five minutes.

## What you need

| Tool | Why |
| --- | --- |
| Go 1.26 | the dsl is Go, the generator is Go |
| sqlc | turns the sql contract into typed Go |
| buf | turns the proto contract into Go and TypeScript |

`sqlc` and `buf` come from the `tool` block in `go.mod`, so `go tool sqlc` and
`go tool buf` work without a separate install.

## 1. Scaffold

```bash
go run github.com/guntisdev/entlite/cmd/entlite new -dialect sqlite User
```

This writes an `ent/` directory: a schema file per entity, `sqlc.yaml`,
`buf.yaml`, `buf.gen.yaml` and two `generate.go` files. Pick the dialect now,
it decides the column types. See [cli](../reference/cli.md) for the flags.

## 2. Write the schema

Open `ent/schema/user.go` and list the fields. The scaffold starts with one:

```go
func (User) Fields() []entlite.Field {
	return []entlite.Field{
		field.String("name").ProtoField(2),
	}
}
```

[Writing a schema](04-writing-a-schema.md) covers fields, queries and indexes.

## 3. Generate

```bash
cd ent && go generate ./...
```

Six steps run in order: the contracts, sqlc, buf, and two entlite steps that
wrap the output. [Pipeline](03-pipeline.md) explains each one.

## 4. Use it

The typed database calls are in `ent/gen/db`, the proto messages and the connect
service in `ent/gen/pb`, the TypeScript client in `ent/gen/ts`.

```go
store := db.New(sqlDB)
user, err := store.CreateUser(ctx, db.CreateUserParams{Name: "Jane"})
```

## Next

- [Project layout](02-project-layout.md), what each folder holds
- [Examples](../examples/01-basic-entity.md), a full entity with every field type

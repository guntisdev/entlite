# Writing a schema

An entity is a Go struct with methods. The generator reads the source, it never
runs it, so everything must be a literal call the parser can see.

The snippets below come from [examples/01-basic-entity/sqlite/ent/schema/user.go](../../examples/01-basic-entity/sqlite/ent/schema/user.go), the
[01-basic-entity](../examples/01-basic-entity.md) example.

## The struct

Embed `entlite.Schema` to mark a struct as an entity. The struct name becomes
the table name and the proto message name.

```go
type User struct {
	entlite.Schema
}
```

## Contracts

`Contracts()` says which layers the entity is generated for. It is the only
required method.

<!-- snippet:examples/01-basic-entity/sqlite/ent/schema/user.go:Contracts:start -->
```go
func (User) Contracts() []entlite.Contract {
	return []entlite.Contract{
		entlite.SQLC(),
		entlite.PROTO(),
	}
}
```
<!-- snippet:examples/01-basic-entity/sqlite/ent/schema/user.go:Contracts:end -->

`SQLC()` gives a table, `PROTO()` gives a message and a service. See
[contracts](../reference/contracts.md).

## Fields

One entry per column, the builder chain sets the options.

<!-- snippet:examples/01-basic-entity/sqlite/ent/schema/user.go:Fields:start -->
```go
func (User) Fields() []entlite.Field {
	return []entlite.Field{
		field.String("email").Unique(),
		// Full name, e.g. "Jane Doe"
		field.String("name").Validate(logic.StartsWithCapital),
		field.Int("age").Optional(),
		field.String("password").Contracts(entlite.SQLC(), entlite.PROTO().WriteOnly()),
		field.Byte("api_key").Immutable().DefaultFunc(logic.GenerateAPIKey),
		field.Bool("is_active").Default(true),
		field.Int64("login_count").Default(0),
		field.Float("rating").Default(0),
		// UI preferences, e.g. {"theme":"dark"}
		field.JSON("preferences").Default("{}"),
		field.Time("created_at").DefaultFunc(time.Now).Immutable().Contracts(entlite.SQLC(), entlite.PROTO().ReadOnly()),
		field.Time("updated_at").DefaultFunc(time.Now).Contracts(entlite.SQLC(), entlite.PROTO().ReadOnly()),
	}
}
```
<!-- snippet:examples/01-basic-entity/sqlite/ent/schema/user.go:Fields:end -->

Things to notice:

- `id` is not listed, the parser adds it
- `password` is written by clients but never returned, `created_at` is the reverse
- `Validate()` and `DefaultFunc()` take a Go func from your `logic` package
- a comment above a field becomes a comment in the generated proto
- a field named after a reserved word, e.g. `table` or `order`, is quoted in the
  generated sql

Every field type and option is in [fields](../reference/fields.md), the column
types per dialect in [type mapping](../reference/type-mapping.md).

## Queries

Each entry becomes one sql query, one wrapper method and, with `PROTO()`, one rpc.

<!-- snippet:examples/01-basic-entity/sqlite/ent/schema/user.go:Queries:start -->
```go
func (User) Queries() []entlite.Query {
	return []entlite.Query{
		query.DefaultCRUD(),
		query.CreateBulk(),
		// Look up a user by email address
		query.GetBy("email"),
		query.ListAll(),
		query.DeleteAll(),
		query.ListBy("is_active").Name("ListActive"),
		query.ListBy(
			filter.Range("age"),   // age BETWEEN :min_age AND :max_age
			filter.Search("name"), // name LIKE :name
		).OrderBy("created_at").Count(),
	}
}
```
<!-- snippet:examples/01-basic-entity/sqlite/ent/schema/user.go:Queries:end -->

`DefaultCRUD()` expands to create, get, update and delete. `Name()` renames the
generated method, which you need when two queries would collide. Filters become
where clauses: `Eq` is `=`, `Range` is `BETWEEN`, `Search` is `LIKE`. Every
`ListBy` is paginated: the query gets `LIMIT`/`OFFSET` and the proto request a
required `limit` and an optional `offset`. `ListAll()` is not paginated.

See [queries](../reference/queries.md) and [filters](../reference/filters.md).

## Indexes

`Indexes()` is optional.

<!-- snippet:examples/01-basic-entity/sqlite/ent/schema/user.go:Indexes:start -->
```go
func (User) Indexes() []entlite.Index {
	return []entlite.Index{
		// compound primary key, it replaces the generated id column
		// index.Primary("email", "created_at"),
		// index on two columns
		index.Fields("age", "is_active"),
		// descending sort
		index.Fields("is_active").
			Desc("created_at"),
		// unique across two columns
		index.Fields("name", "email").Unique(),
		// explicit index name
		index.Fields("login_count", "rating").
			Name("idx_users_stats"),
	}
}
```
<!-- snippet:examples/01-basic-entity/sqlite/ent/schema/user.go:Indexes:end -->

`index.Primary()` replaces the automatic `id` primary key with a compound one.
See [indexes](../reference/indexes.md).

## What the parser rejects

| Rule | On failure |
| --- | --- |
| `Contracts()` exists and lists a layer | generation stops |
| a proto field number is unique in the entity | generation stops |
| a json `Default()` is valid json | generation stops |
| an index names a field that exists | generation stops |

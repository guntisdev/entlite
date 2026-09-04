# 04-contracts

`Contracts()` decides where an entity, a field or a query shows up. Four
entities, one for each combination.

<!-- teaches:start -->
- `SQLC()` only — a table with no proto message and no service
- `PROTO()` only — a message and a service with no table
- Both — a table served to clients
- `PROTO().ReadOnly()` — a table and a service, but no write rpc
- Query level `Contracts()`: a query the server runs that gets no rpc
- Field level `PROTO().ReadOnly()` for server managed timestamps
- Contracts on an entity are the default; a field or query can narrow them
- `index.Primary("name")` — a natural key, the entity gets no generated `id` column
<!-- teaches:end -->

## Four entities

A chess club. [`sqlite/ent/schema`](sqlite/ent/schema):

| Entity | Contracts | Table | Service |
|---|---|---|---|
| [`Match`](sqlite/ent/schema/match.go) | `SQLC()`, `PROTO()` | yes | create, get, delete, list |
| [`Player`](sqlite/ent/schema/player.go) | `SQLC()`, `PROTO().ReadOnly()` | yes | get, list |
| [`Standing`](sqlite/ent/schema/standing.go) | `PROTO()` | no | list |
| [`Audit`](sqlite/ent/schema/audit.go) | `SQLC()` | yes | none |

Check it against the generated files. `schema.sql` has three tables and no
`standing`. `schema.proto` has three messages and no `Audit`.

**Match** is the normal case. Both contracts, read and write.

**Player** is the roster. The proto contract is `ReadOnly()`, so there is no
create, update or delete rpc. `query.Create()` is still declared and still
becomes a database query — the server calls it in `SeedRoster`. The read only
contract only removes the rpc.

It is also the entity with a natural key:

```go
func (Player) Indexes() []entlite.Index {
	return []entlite.Index{
		index.Primary("name"),
	}
}
```

A club roster is identified by the player name, and `Match` already refers to
players that way, by name and not by id. So the name is the primary key and
there is no `id` column at all — `index.Primary` replaces the one entlite would
otherwise generate. Everything keyed by the primary key follows it: the column
is `PRIMARY KEY` in `schema.sql`, `query.Get()` becomes `GetPlayerByName` with
a `GetByName` rpc, and `CreatePlayer` returns only an `error`, because there is
no generated id left for the database to hand back.

**Standing** is a league table counted from matches on every request. There is
no `standing` table. The entity exists to give the response a typed shape. Its
`id` field carries the rank.

**Audit** is an internal log. No proto message, no service. The server writes
rows and reads them back through a plain HTTP endpoint, so you can see they
exist without exposing them.

## Narrowing a single query

An entity's contracts are the default for its queries. A query can take less.

```go
query.DeleteAll().Contracts(entlite.SQLC())   // Match, end of season cleanup
```

`Match` has both contracts, so its other queries get an rpc. This one does not.
It is a database query the server calls itself.

## Narrowing a single field

Same idea on fields:

```go
field.Time("created_at").Contracts(entlite.SQLC(), entlite.PROTO().ReadOnly())
```

The column exists and the field is in the response. It is not in create or
update requests, because `DefaultFunc(time.Now)` sets it on the server.

## Run

```bash
cd sqlite
make run     # serves on :8080
```

The page has a panel per entity. The Player one is read only: get by name and
list the roster.

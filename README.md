# Entlite
Entity-first generator for SQLC and Proto files. Maps DB and Protobuf types automatically to maintain a single source of truth in Go services.

## TODO
* Create /internal/naming/ to have in one place consistant naming
* Implement DefaultFunc for sql generation
* Add mcp for visual testing of examples/
* Add edge cases to examples - uuid as id, everything as optional, custom proto and queries files etc
* Split get/list/delete sqlc wraps in separate files
* Move query name to parser instead of generator
* Fix Optional() with Validate() - generated code passes a pointer to a value func and does not compile
* Comment out in examples Queries use case for: Count, GroupBy, Having, OrderBy
* Implement Queries Count()
* Implement Queries GroupBy()
* Implement Queries Having()
* Implement Queries OrderBy()
* Figure out migration

## Permissions refactor
`Permissions()` does two jobs at once. It says which contracts a field belongs to, and it says what can be done with the field. The names also do not match the rest of the DSL: entity level says `SQLC` and `PROTO`, field level says `Db` and `Api` for the same two things. And those names cannot be reused higher up, because `DbRead` or `ApiWrite` mean nothing for a whole entity or for a single query.

The refactor splits the two jobs. `Contracts()` says where something exists. `.ReadOnly()` and `.WriteOnly()` on a contract say what can be done with it there. The same two words then work at entity, field and query level. Queries also get `Contracts()`, so a query can be generated for sqlc only, proto only, or both. A query cannot take `.ReadOnly()`, because a query is already a read or a write, and the compiler blocks it.

New spelling:
```go
entlite.PROTO().ReadOnly()                                          // entity: no write rpc
field.String("password").Contracts(entlite.SQLC(), entlite.PROTO().WriteOnly())
field.Float("latest_value").Contracts(entlite.PROTO())              // no column
query.Create().Contracts(entlite.SQLC())                            // no rpc
```

Old to new:
* `permissions.Default` -> omit, inherits entity contracts
* `permissions.Internal` -> `Contracts(entlite.SQLC())`
* `permissions.Virtual` -> `Contracts(entlite.PROTO())`
* `permissions.ReadOnly` -> `Contracts(entlite.SQLC(), entlite.PROTO().ReadOnly())`
* `permissions.WriteOnly` -> `Contracts(entlite.SQLC(), entlite.PROTO().WriteOnly())`

`created_at` keeps a plain `entlite.SQLC()`, because the generated INSERT writes that column through `DefaultFunc`. `SQLC().ReadOnly()` means the generated SQL never writes the column, for example a value kept by a trigger.

Steps, contract level first:
* ~~Add `Layer` interface in pkg/entlite/schema.go: embeds `Contract`, adds `ReadOnly() Contract` and `WriteOnly() Contract`. `SQLC()` and `PROTO()` return `Layer`~~
* ~~Drop the unused variadic arg from `PROTO(contracts ...Contract)`~~
* ~~internal/schema: add access to `Contract`~~
* ~~Parser: read `.ReadOnly()` and `.WriteOnly()` chains in contracts.go~~
* ~~Generators: with `PROTO().ReadOnly()` skip create, create_bulk, update, delete and delete_all from proto, same for `SQLC().ReadOnly()` in sqlc~~

Then query level:
* ~~Add `Contracts(...Layer)` to `QueryOperations` and `ListByOperations`~~
* ~~internal/schema: add `Query.Contracts`, `Entity.SQLCQueries()` and `Entity.ProtoQueries()`~~
* ~~Parser: parse query contracts in queries.go, reuse `parseContractCall`~~
* ~~Parser: default empty query contracts to the entity contracts, after the method loop in parser.go, because `Queries()` can be parsed before `Contracts()`~~
* ~~Parser: error when a query contract is not in the entity contracts, or when the entity access already dropped that query~~
* ~~Parser: error on `.ReadOnly()` or `.WriteOnly()` inside a query, as a backstop for the compile error~~
* ~~proto generator: use `ProtoQueries()`, skip the service block when it is empty but keep the message~~
* ~~sqlc and sqlcWrap generators: use `SQLCQueries()`~~

Then field level:
* ~~Add `Contracts(...Contract)` to the field builder~~
* ~~internal/schema: replace `Field.Permissions` with contracts plus access, rewrite `Field.IsVirtual()` as "has no sqlc contract", with `Entity.IsFieldVirtual()` for entities that may be proto only~~
* ~~Parser: parse field contracts in fields.go, default them to the entity contracts, validate the subset~~
* ~~Generators: replace the permission bit checks in proto, protoValidate, sqlc and sqlcWrap~~
* ~~Delete pkg/entlite/permissions~~

Examples, each one shows only what fits it:
* 01-basic-entity, all three dialects: field level only. `password` becomes write only, `created_at` and `updated_at` become read only
* 03-optional: same field migration, no new concepts, it stays an example about optional types
* 02-custom: field migration, `latest_value` becomes `Contracts(entlite.PROTO())`. Then query level on Reading: expand `DefaultCRUD()` and give `query.Update()` the sqlc contract only, because a recorded measurement is not editable by clients. Remove `ReadingServer.Update` from server.go
* 04-contracts: keep Audit as sqlc only and Standing as proto only. Add `query.DeleteAll().Contracts(entlite.SQLC())` to Match for end of season cleanup with no rpc. Add a Player entity with `entlite.SQLC(), entlite.PROTO().ReadOnly()`, where the roster is written by the server and only read by clients, so `query.Create()` is dropped from proto by the entity access
* Regenerate all examples, check that servers and web clients still build


## Folder structure
```
└── ent/
    ├── schema/             # DSL entities
    ├── contract/
    │   ├── proto/          # generated from DSL: schema.proto. Custom proto could be added here
    │   └── sqlc/           # generated from DSL: schema.sql, queries.sql. Custom sql could be added here
    ├── gen/
    │   ├── db/             # small wrapper for type convertions - nullptr etc
    │   |   └── internal/   # generated from sqlc contract
    │   ├── pb/             # generated from proto contract
    │   └── ts/             # generated from proto contract
    ├── logic/              # optional, custom functions for DSL entities
    ├── buf.yaml
    ├── buf.gen.yaml
    ├── sqlc.yaml
    └── generate.go     # go generate - creates contracts, launches sqlc, buf, light db wrapper and convert
```

## Get started
sql dialect flag: postgresql (default) or sqlite or mysql
arguments: entity names
```bash
go run github.com/guntisdev/entlite/cmd/entlite new --dialect sqlite User Post
```

## Launch example
Each example has a Makefile that generates types, bundles the JavaScript and starts the web server
```bash
cd examples/01-basic-entity/sqlite
make run
```
The `postgres` and `mysql` variants run their database from `docker-compose.yml`, `make run` starts it
and waits until it is ready. Use `make down` to stop it and drop its data
```bash
cd examples/01-basic-entity/postgres
make run
```
Individual steps are available as `make gen`, `make web` and (for postgres/mysql) `make db`.
`make clean` removes node_modules, the bundle and the database

Doing it by hand, generate types
```bash
cd examples/01-basic-entity/sqlite
cd ent/
go generate
```
Build JavaScript
```bash
cd web/
npm install
npm run build
```
Run go web server
```bash
go run main.go
```

## Release
```bash
# Tag and push version
git tag v0.0.3
git push origin v0.0.3

# Force Go proxy cache update
GOPROXY=https://proxy.golang.org GOPRIVATE= go list -m github.com/guntisdev/entlite@v0.0.3

# Update CLI tool & library dependencies in projects
go install github.com/guntisdev/entlite/cmd/entlite@v0.0.3
go get github.com/guntisdev/entlite@v0.0.3
```


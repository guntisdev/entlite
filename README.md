# Entlite
Entity-first generator for SQLC and Proto files. Maps DB and Protobuf types automatically to maintain a single source of truth in Go services.

## TODO
* Gate sqlc generation on `SQLC()` in genCommand - currently it generates for all entities
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

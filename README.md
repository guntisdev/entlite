# Entlite
Entity-first generator for SQLC and Proto files. Maps DB and Protobuf types automatically to maintain a single source of truth in Go services.

## TODO
* Create /internal/naming/ to have in one place consistant naming
* Implement DefaultFunc for sql generation
* Add mcp for visual testing of examples/
* Add a uuid foreign key example - uuid as id, all optional and custom files are covered
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

## Examples
| Example | Teaches |
|---|---|
| [01-basic-entity](examples/01-basic-entity) | every field type, every query type, indexes, three dialects |
| [02-custom](examples/02-custom) | hand-written SQL and proto next to the generated files |
| [03-optional](examples/03-optional) | `Optional()` per type, optional filters, a uuid key |
| [04-contracts](examples/04-contracts) | `Contracts()` on an entity, a field and a query |

`make teaches` prints every list, for pulling into documentation.

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


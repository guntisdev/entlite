# Entlite
Entity-first generator for SQLC and Proto files. Maps DB and Protobuf types automatically to maintain a single source of truth in Go services.

## TODO
* Use `protovalidate-go` to intercept in `grpc.NewServer()` to call custom Validate() functions in proto exports
* get export directory for proto validate from yaml file
* improve integration test - less mocks and folder changing. Maybe copy all content to tmp dir, generate in same dir, compare and then put back from tmp?
* check each field if it added in further generation (for example Comment)
* Add edge cases to examples - uuid as id, everything as optional, custom proto and queries files etc
* Edge case with all types optional
* Split get/list/delete sqlc wraps in separate files

## Folder structure
```
└── ent/
    ├── schema/             # DSL entities
    ├── contract/
    │   ├── proto/          # generated from DSL: schema.prot. Custom proto could be added here
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
Go to one of examples and generate types
```bash
cd examples/01-sqlite-entity
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

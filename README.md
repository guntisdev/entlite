# Entlite
Entity-first generator for SQLC and Proto files. Maps DB and Protobuf types automatically to maintain a single source of truth in Go services.

## TODO
* Add to proto required field check `[(buf.validate.field).required = true];` 
* Use `protovalidate-go` to intercept in `grpc.NewServer()` to call custom Validate() functions in proto exports
* improve integration test - less mocks and folder changing. Maybe copy all content to tmp dir, generate in same dir, compare and then put back from tmp?
* handle sql dialect passing to newCommand
* read sql dialect from sqlc.yaml when generate.go - use for correct convertion/validation/wrapping
* fix sqlite convertion types (integer=int64, boolean=int64)
* check each field if it added in further generation (for example Comment)
* add more field types: int64? float32? float64?

## Folder structure
```
└── ent/
    ├── schema/         # DSL entities
    ├── contract/
    │   ├── proto/      # generated from DSL: schema.prot. Custom proto could be added here
    │   └── sqlc/       # generated from DSL: schema.sql, queries.sql. Custom sql could be added here
    ├── gen/
    │   ├── db/         # generated from sqlc contract
    │   ├── pb/         # generated from proto contract
    │   ├── convert/    # generated from DSL - convertion between db and pb types
    │   └── ts/         # generated from proto contract
    ├── logic/          # optional, custom functions for DSL entities
    ├── buf.yaml
    ├── buf.gen.yaml
    ├── sqlc.yaml
    └── generate.go     # go generate - creates contracts, launches sqlc, buf, light db wrapper and convert
```

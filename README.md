# Entlite
Entity-first generator for SQLC and Proto files. Maps DB and Protobuf types automatically to maintain a single source of truth in Go services.

## TODO
* Generate convert proto to db
* Add Validation() for field
* Add Validation func to sqlc output
* Add to proto required field check `[(buf.validate.field).required = true];` 
* Add Validate() method to proto export struct
* Use `protovalidate-go` to intercept in `grpc.NewServer()` to call custom Validate() functions in proto exports
* Fix getting import path and reuse functionality (currently some copy paste code)

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

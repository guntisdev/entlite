# Entlite
Entity-first generator for SQLC and Proto files. Maps DB and Protobuf types automatically to maintain a single source of truth in Go services.

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

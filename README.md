# Entlite
Entity-first generator for SQLC and Proto files. Maps DB and Protobuf types automatically to maintain a single source of truth in Go services.

## TODO
* Create /internal/naming/ to have in one place consistant naming
* improve integration test - less mocks and folder changing. Maybe copy all content to tmp dir, generate in same dir, compare and then put back from tmp?
* check each field if it added in further generation (for example Comment)
* Add edge cases to examples - uuid as id, everything as optional, custom proto and queries files etc
* Split get/list/delete sqlc wraps in separate files
* Add Queries:
```bash
func (User) Queries() []entlite.Query {
    return []entlite.Query{
        query.DefaultCRUD(), // Generates Create, Get, Update, Delete, List - by ID
        query.Get(), // one of default crud. Translates to .GetBy("ID")
        query.GetBy("email"), // validates unique
        query.GetBy("org_id", "email") // later handle composite indexes
        query.ListBy("org_id"), // allow either one param field name (default to .Eq)
        query.ListBy(
            filter.Range("age") // >= n <=
            filter.Search("name") // WHERE name LIKE 'Bob'
            filter.Search("surname").Optional() // set OR in sql
            filter.Eq("score")
        ),
        query.ListBy("org_id").Count(), // SELECT COUNT(*) FROM
        query.ListBy("org_id").OrderBy("created_at"), // adds also ASC/DESC 
    }
}
```
* Make annotations optional
* Add new field type JSON
* Add Name() option for Queries - for shorter or custom naming
* Add .Indexes() for fields (to index in sql)
```bash
func (User) Indexes() []entlite.Index {
	return []entlite.Index{
		// 1. Primary Key (Compound)
		index.Primary("country", "timestamp", "env"),

		// 2. Simple Single/Multi-Column Index
		index.Fields("env", "is_active"),

		// 3. Composite Index with Sort Ordering (ASC / DESC)
		index.Fields("country", "env").
			Desc("created_at"), // Sorting timestamp DESC for fast time-series queries

		// 4. Multi-Column Unique Constraint
		index.Fields("tenant_id", "email").Unique(),

		// 5. Named Index (Useful to avoid DB auto-generated name conflicts)
		index.Fields("login_count", "rating").
			Name("idx_users_stats"),
	}
}
```
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

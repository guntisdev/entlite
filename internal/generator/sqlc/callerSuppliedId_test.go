package sqlc

import (
	"strings"
	"testing"

	"github.com/guntisdev/entlite/internal/schema"
)

// index.Primary takes the key from the id field, so the caller sends the id
func suppliedIdEntity() schema.Entity {
	contracts := []schema.Contract{{Type: schema.ContractSQLC}, {Type: schema.ContractPROTO}}

	return schema.Entity{
		Name:      "Casino",
		Contracts: contracts,
		Fields: []schema.Field{
			{Name: "id", Type: schema.FieldTypeString, Unique: true, Immutable: true, Contracts: contracts},
			{Name: "env", Type: schema.FieldTypeString, Immutable: true, Contracts: contracts},
			{Name: "init_count", Type: schema.FieldTypeInt, Contracts: contracts},
		},
		Indexes: []schema.Index{{
			Type:    schema.IndexPrimary,
			Columns: []schema.IndexColumn{{Name: "id"}, {Name: "env"}},
		}},
		Queries: []schema.Query{
			{Type: schema.QueryCreate, Name: "CreateCasino", Contracts: contracts},
			{Type: schema.QueryGetBy, Name: "GetCasinoByIDEnv", Fields: []string{"ID", "env"}, PrimaryKey: true, Contracts: contracts},
		},
	}
}

// id is a plain column now: NOT NULL, no PRIMARY KEY
func TestCallerSuppliedIdTable(t *testing.T) {
	for _, dialect := range []schema.SQLDialect{schema.PostgreSQL, schema.SQLite, schema.MySQL} {
		t.Run(string(dialect), func(t *testing.T) {
			sql := NewGenerator(dialect).generateTableSQL(suppliedIdEntity())

			if !strings.Contains(sql, "id TEXT NOT NULL") && !strings.Contains(sql, "id VARCHAR(255) NOT NULL") {
				t.Errorf("expected the id column to be NOT NULL:\n%s", sql)
			}
			if !strings.Contains(sql, "PRIMARY KEY (id, env)") {
				t.Errorf("expected the compound primary key:\n%s", sql)
			}
			if strings.Contains(sql, "AUTOINCREMENT") || strings.Contains(sql, "AUTO_INCREMENT") || strings.Contains(sql, "SERIAL") {
				t.Errorf("expected no generated key:\n%s", sql)
			}
			// PRIMARY KEY belongs to the index, not to the column
			if strings.Contains(sql, "id TEXT PRIMARY KEY") {
				t.Errorf("expected the id column to give up the primary key:\n%s", sql)
			}
		})
	}
}

// the insert sends the id and returns nothing, the caller knows it already
func TestCallerSuppliedIdInsert(t *testing.T) {
	for _, dialect := range []schema.SQLDialect{schema.PostgreSQL, schema.SQLite, schema.MySQL} {
		t.Run(string(dialect), func(t *testing.T) {
			sql := NewGenerator(dialect).generateCRUDQueries(suppliedIdEntity())

			if !strings.Contains(sql, "-- name: CreateCasino :exec") {
				t.Errorf("expected the insert to return nothing:\n%s", sql)
			}
			if !strings.Contains(sql, " id,") {
				t.Errorf("expected the id in the insert columns:\n%s", sql)
			}
			if strings.Contains(sql, "RETURNING id") {
				t.Errorf("expected no RETURNING on the insert:\n%s", sql)
			}
		})
	}
}

// an id the db makes still stays out of the insert
func TestGeneratedIdStaysOutOfInsert(t *testing.T) {
	contracts := []schema.Contract{{Type: schema.ContractSQLC}, {Type: schema.ContractPROTO}}
	entity := schema.Entity{
		Name:      "Post",
		Contracts: contracts,
		Fields: []schema.Field{
			{Name: "id", Type: schema.FieldTypeInt, Primary: true, Unique: true, Contracts: contracts},
			{Name: "title", Type: schema.FieldTypeString, Contracts: contracts},
		},
		Queries: []schema.Query{
			{Type: schema.QueryCreate, Name: "CreatePost", Contracts: contracts},
		},
	}

	sql := NewGenerator(schema.PostgreSQL).generateCRUDQueries(entity)

	if !strings.Contains(sql, "-- name: CreatePost :one") {
		t.Errorf("expected the insert to return the generated id:\n%s", sql)
	}
	if strings.Contains(sql, " id,") {
		t.Errorf("expected the id out of the insert columns:\n%s", sql)
	}
	if !strings.Contains(sql, "RETURNING id") {
		t.Errorf("expected RETURNING on the insert:\n%s", sql)
	}
}

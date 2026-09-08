package sqlc

import (
	"strings"
	"testing"

	"github.com/guntisdev/entlite/internal/schema"
)

// the id field gives up the key to index.Primary, so the caller supplies its value
func suppliedIdEntity() schema.Entity {
	contracts := []schema.Contract{{Type: schema.ContractSQLC}, {Type: schema.ContractPROTO}}

	return schema.Entity{
		Name:      "Casino",
		Contracts: contracts,
		Fields: []schema.Field{
			{Name: "ID", Type: schema.FieldTypeString, Unique: true, Immutable: true, Contracts: contracts},
			{Name: "env", Type: schema.FieldTypeString, Immutable: true, Contracts: contracts},
			{Name: "init_count", Type: schema.FieldTypeInt, Contracts: contracts},
		},
		Indexes: []schema.Index{{
			Type:    schema.IndexPrimary,
			Columns: []schema.IndexColumn{{Name: "ID"}, {Name: "env"}},
		}},
		Queries: []schema.Query{
			{Type: schema.QueryCreate, Name: "CreateCasino", Contracts: contracts},
			{Type: schema.QueryGetBy, Name: "GetCasinoByIDEnv", Fields: []string{"ID", "env"}, PrimaryKey: true, Contracts: contracts},
		},
	}
}

// the id is a plain column now, it carries its own NOT NULL and no PRIMARY KEY
func TestCallerSuppliedIdTable(t *testing.T) {
	for _, dialect := range []schema.SQLDialect{schema.PostgreSQL, schema.SQLite, schema.MySQL} {
		t.Run(string(dialect), func(t *testing.T) {
			sql := NewGenerator(dialect).generateTableSQL(suppliedIdEntity())

			if !strings.Contains(sql, "ID TEXT NOT NULL") && !strings.Contains(sql, "ID VARCHAR(255) NOT NULL") {
				t.Errorf("expected the id column to be NOT NULL:\n%s", sql)
			}
			if !strings.Contains(sql, "PRIMARY KEY (ID, env)") {
				t.Errorf("expected the compound primary key:\n%s", sql)
			}
			if strings.Contains(sql, "AUTOINCREMENT") || strings.Contains(sql, "AUTO_INCREMENT") || strings.Contains(sql, "SERIAL") {
				t.Errorf("expected no generated key:\n%s", sql)
			}
			// the inline PRIMARY KEY belongs to the index, never to the column
			if strings.Contains(sql, "ID TEXT PRIMARY KEY") {
				t.Errorf("expected the id column to give up the primary key:\n%s", sql)
			}
		})
	}
}

// the insert carries the id and returns nothing, the caller already knows it
func TestCallerSuppliedIdInsert(t *testing.T) {
	for _, dialect := range []schema.SQLDialect{schema.PostgreSQL, schema.SQLite, schema.MySQL} {
		t.Run(string(dialect), func(t *testing.T) {
			sql := NewGenerator(dialect).generateCRUDQueries(suppliedIdEntity())

			if !strings.Contains(sql, "-- name: CreateCasino :exec") {
				t.Errorf("expected the insert to return nothing:\n%s", sql)
			}
			if !strings.Contains(sql, " ID,") {
				t.Errorf("expected the id in the insert columns:\n%s", sql)
			}
			if strings.Contains(sql, "RETURNING ID") {
				t.Errorf("expected no RETURNING on the insert:\n%s", sql)
			}
		})
	}
}

// an id the database assigns still stays out of the insert
func TestGeneratedIdStaysOutOfInsert(t *testing.T) {
	contracts := []schema.Contract{{Type: schema.ContractSQLC}, {Type: schema.ContractPROTO}}
	entity := schema.Entity{
		Name:      "Post",
		Contracts: contracts,
		Fields: []schema.Field{
			{Name: "ID", Type: schema.FieldTypeInt, Primary: true, Unique: true, Contracts: contracts},
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
	if strings.Contains(sql, " ID,") {
		t.Errorf("expected the id out of the insert columns:\n%s", sql)
	}
	if !strings.Contains(sql, "RETURNING ID") {
		t.Errorf("expected RETURNING on the insert:\n%s", sql)
	}
}

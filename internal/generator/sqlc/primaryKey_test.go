package sqlc

import (
	"strings"
	"testing"

	"github.com/guntisdev/entlite/internal/schema"
)

func compoundKeyEntity() schema.Entity {
	contracts := []schema.Contract{{Type: schema.ContractSQLC}, {Type: schema.ContractPROTO}}

	return schema.Entity{
		Name:      "Setting",
		Contracts: contracts,
		Fields: []schema.Field{
			{Name: "country", Type: schema.FieldTypeString, Contracts: contracts, Immutable: true},
			{Name: "env", Type: schema.FieldTypeString, Contracts: contracts, Immutable: true},
			{Name: "value", Type: schema.FieldTypeString, Contracts: contracts},
		},
		Indexes: []schema.Index{{
			Type:    schema.IndexPrimary,
			Columns: []schema.IndexColumn{{Name: "country"}, {Name: "env"}},
		}},
		Queries: []schema.Query{
			{Type: schema.QueryCreate, Contracts: contracts},
			{Type: schema.QueryGetBy, Fields: []string{"country", "env"}, PrimaryKey: true, Contracts: contracts},
			{Type: schema.QueryUpdate, Fields: []string{"country", "env"}, PrimaryKey: true, Contracts: contracts},
			{Type: schema.QueryDelete, Fields: []string{"country", "env"}, PrimaryKey: true, Contracts: contracts},
		},
	}
}

func TestCompoundPrimaryKeyTable(t *testing.T) {
	for _, dialect := range []schema.SQLDialect{schema.PostgreSQL, schema.SQLite, schema.MySQL} {
		t.Run(string(dialect), func(t *testing.T) {
			sql := NewGenerator(dialect).generateTableSQL(compoundKeyEntity())

			if strings.Contains(sql, "ID") {
				t.Errorf("expected no id column:\n%s", sql)
			}
			if !strings.Contains(sql, "PRIMARY KEY (country, env)") {
				t.Errorf("expected a compound primary key:\n%s", sql)
			}
			// the first column must not be preceded by a comma
			if strings.Contains(sql, "(\n,") {
				t.Errorf("stray comma before the first column:\n%s", sql)
			}
		})
	}
}

func TestCompoundPrimaryKeyQueries(t *testing.T) {
	tests := []struct {
		dialect schema.SQLDialect
		want    []string
	}{
		{schema.PostgreSQL, []string{
			"-- name: CreateSetting :exec",
			"DELETE FROM \"setting\" WHERE country = $1 AND env = $2;",
			"WHERE country = @country AND env = @env",
		}},
		{schema.SQLite, []string{
			"-- name: CreateSetting :exec",
			"DELETE FROM \"setting\" WHERE country = ? AND env = ?;",
			"WHERE country = @country AND env = @env",
		}},
		{schema.MySQL, []string{
			"-- name: CreateSetting :exec",
			"DELETE FROM `setting` WHERE country = ? AND env = ?;",
			"WHERE country = sqlc.arg('country') AND env = sqlc.arg('env')",
		}},
	}

	for _, tt := range tests {
		t.Run(string(tt.dialect), func(t *testing.T) {
			sql := NewGenerator(tt.dialect).generateCRUDQueries(compoundKeyEntity())

			for _, want := range tt.want {
				if !strings.Contains(sql, want) {
					t.Errorf("expected %q in the generated SQL:\n%s", want, sql)
				}
			}
			// nothing is returned, the caller supplies the whole key
			if strings.Contains(sql, "RETURNING country") {
				t.Errorf("expected no RETURNING on the insert:\n%s", sql)
			}
			// the key columns are not part of the SET list
			if strings.Contains(sql, "country = @country,") {
				t.Errorf("expected the primary key out of the update SET:\n%s", sql)
			}
		})
	}
}

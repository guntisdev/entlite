package sqlc

import (
	"strings"
	"testing"

	"github.com/guntisdev/entlite/internal/schema"
)

func timeNowFunc() any { return "time.Now" }
func newUUIDFunc() any { return "logic.NewUUID" }

func defaultFuncEntity() schema.Entity {
	contracts := []schema.Contract{{Type: schema.ContractSQLC}, {Type: schema.ContractPROTO}}

	return schema.Entity{
		Name:      "Article",
		Contracts: contracts,
		Fields: []schema.Field{
			{Name: "id", Type: schema.FieldTypeInt, Primary: true, Contracts: contracts},
			{Name: "title", Type: schema.FieldTypeString, Contracts: contracts},
			{Name: "api_key", Type: schema.FieldTypeString, Immutable: true, Contracts: contracts, DefaultFunc: newUUIDFunc},
			{Name: "created_at", Type: schema.FieldTypeTime, Immutable: true, Contracts: contracts, DefaultFunc: timeNowFunc},
		},
	}
}

func TestDefaultFuncWithKnownSQLEquivalentGetsDDLDefault(t *testing.T) {
	for _, dialect := range []schema.SQLDialect{schema.PostgreSQL, schema.SQLite, schema.MySQL} {
		t.Run(string(dialect), func(t *testing.T) {
			sql := NewGenerator(dialect).generateTableSQL(defaultFuncEntity())

			if !strings.Contains(sql, "created_at") || !strings.Contains(sql, "DEFAULT CURRENT_TIMESTAMP") {
				t.Errorf("expected created_at to get DEFAULT CURRENT_TIMESTAMP:\n%s", sql)
			}
		})
	}
}

func TestDefaultFuncWithoutSQLEquivalentGetsNoDDLDefault(t *testing.T) {
	sql := NewGenerator(schema.PostgreSQL).generateTableSQL(defaultFuncEntity())

	if strings.Contains(sql, "api_key TEXT DEFAULT") {
		t.Errorf("expected no DEFAULT clause for a DefaultFunc with no SQL equivalent:\n%s", sql)
	}
	if !strings.Contains(sql, "api_key TEXT NOT NULL") {
		t.Errorf("expected api_key to stay a plain NOT NULL column:\n%s", sql)
	}
}

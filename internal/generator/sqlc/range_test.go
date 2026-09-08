package sqlc

import (
	"strings"
	"testing"

	"github.com/guntisdev/entlite/internal/schema"
)

func rangeEntity() schema.Entity {
	contracts := []schema.Contract{{Type: schema.ContractSQLC}, {Type: schema.ContractPROTO}}

	return schema.Entity{
		Name:      "User",
		Contracts: contracts,
		Fields: []schema.Field{
			{Name: "ID", Type: schema.FieldTypeInt, Primary: true, Contracts: contracts},
			{Name: "name", Type: schema.FieldTypeString, Contracts: contracts},
			{Name: "age", Type: schema.FieldTypeInt, Contracts: contracts},
		},
		Queries: []schema.Query{
			// range first, then a second filter
			{Type: schema.QueryListBy, Name: "ListUserRangeFirst", Contracts: contracts,
				Filters: []schema.QueryFilter{
					{Type: schema.QueryFilterRange, Field: "age"},
					{Type: schema.QueryFilterSearch, Field: "name"},
				}},
			// range last, the order sqlc could not bind on sqlite
			{Type: schema.QueryListBy, Name: "ListUserRangeLast", Contracts: contracts,
				Filters: []schema.QueryFilter{
					{Type: schema.QueryFilterSearch, Field: "name"},
					{Type: schema.QueryFilterRange, Field: "age"},
				}},
		},
	}
}

func TestRangeClause(t *testing.T) {
	tests := []struct {
		dialect schema.SQLDialect
		want    []string
	}{
		{schema.PostgreSQL, []string{
			`SELECT * FROM "user" WHERE age BETWEEN @min_age AND @max_age AND name LIKE @name;`,
			`SELECT * FROM "user" WHERE name LIKE @name AND age BETWEEN @min_age AND @max_age;`,
		}},
		{schema.SQLite, []string{
			`SELECT * FROM "user" WHERE age >= @min_age AND age <= @max_age AND name LIKE @name;`,
			`SELECT * FROM "user" WHERE name LIKE @name AND age >= @min_age AND age <= @max_age;`,
		}},
		{schema.MySQL, []string{
			"SELECT * FROM `user` WHERE age BETWEEN sqlc.arg('min_age') AND sqlc.arg('max_age') AND name LIKE sqlc.arg('name');",
			"SELECT * FROM `user` WHERE name LIKE sqlc.arg('name') AND age BETWEEN sqlc.arg('min_age') AND sqlc.arg('max_age');",
		}},
	}

	for _, test := range tests {
		t.Run(string(test.dialect), func(t *testing.T) {
			sql := NewGenerator(test.dialect).generateCRUDQueries(rangeEntity())

			for _, want := range test.want {
				if !strings.Contains(sql, want) {
					t.Errorf("expected %s in the generated SQL:\n%s", want, sql)
				}
			}

			// sqlite has no BETWEEN left to lose the args
			if got := strings.Contains(sql, "BETWEEN"); got != (test.dialect != schema.SQLite) {
				t.Errorf("unexpected BETWEEN usage for %s:\n%s", test.dialect, sql)
			}
		})
	}
}

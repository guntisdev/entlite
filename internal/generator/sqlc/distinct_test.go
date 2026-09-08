package sqlc

import (
	"strings"
	"testing"

	"github.com/guntisdev/entlite/internal/schema"
)

func distinctEntity() schema.Entity {
	contracts := []schema.Contract{{Type: schema.ContractSQLC}, {Type: schema.ContractPROTO}}

	return schema.Entity{
		Name:      "Gametype",
		Contracts: contracts,
		Fields: []schema.Field{
			{Name: "ID", Type: schema.FieldTypeInt, Primary: true, Contracts: contracts},
			{Name: "env", Type: schema.FieldTypeString, Contracts: contracts},
			{Name: "gametype", Type: schema.FieldTypeString, Contracts: contracts},
			{Name: "name", Type: schema.FieldTypeString, Contracts: contracts},
			{Name: "order", Type: schema.FieldTypeInt, Contracts: contracts},
			{Name: "timestamp", Type: schema.FieldTypeTime, Contracts: contracts},
		},
		Queries: []schema.Query{
			// the use case: distinct values of one column, filtered and sorted
			{Type: schema.QueryListBy, Name: "ListGametypeNames",
				Filters: []schema.QueryFilter{
					{Field: "env", Type: schema.QueryFilterEq},
					{Field: "timestamp", Type: schema.QueryFilterRange},
				},
				Distinct: []string{"gametype"},
				OrderBy:  []schema.OrderColumn{{Name: "gametype"}}, Contracts: contracts},
			// several columns dedupe on the whole tuple
			{Type: schema.QueryListAll, Name: "ListReleasePairs",
				Distinct: []string{"name", "timestamp"},
				OrderBy:  []schema.OrderColumn{{Name: "name"}, {Name: "timestamp", Desc: true}}, Contracts: contracts},
			// a reserved column name still gets quoted
			{Type: schema.QueryListAll, Name: "ListGametypeOrders",
				Distinct: []string{"order"}, Contracts: contracts},
			// paging a distinct select
			{Type: schema.QueryListAll, Name: "ListGametypePagedNames",
				Distinct: []string{"gametype"}, HasLimit: true, HasOffset: true, Contracts: contracts},
			{Type: schema.QueryListAll, Name: "ListAllGametype", Contracts: contracts},
		},
	}
}

func TestDistinctSelect(t *testing.T) {
	tests := []struct {
		dialect schema.SQLDialect
		want    []string
	}{
		{schema.PostgreSQL, []string{
			`SELECT DISTINCT gametype FROM "gametype" WHERE env = @env AND timestamp BETWEEN @min_timestamp AND @max_timestamp ORDER BY gametype;`,
			`SELECT DISTINCT name, timestamp FROM "gametype" ORDER BY name, timestamp DESC;`,
			`SELECT DISTINCT "order" FROM "gametype";`,
			`SELECT DISTINCT gametype FROM "gametype" LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');`,
			`SELECT * FROM "gametype";`,
		}},
		{schema.SQLite, []string{
			`SELECT DISTINCT gametype FROM "gametype" WHERE env = @env AND timestamp >= @min_timestamp AND timestamp <= @max_timestamp ORDER BY gametype;`,
			`SELECT DISTINCT name, timestamp FROM "gametype" ORDER BY name, timestamp DESC;`,
			`SELECT DISTINCT "order" FROM "gametype";`,
			`SELECT * FROM "gametype";`,
		}},
		{schema.MySQL, []string{
			"SELECT DISTINCT gametype FROM `gametype` WHERE env = sqlc.arg('env') AND timestamp BETWEEN sqlc.arg('min_timestamp') AND sqlc.arg('max_timestamp') ORDER BY gametype;",
			"SELECT DISTINCT name, timestamp FROM `gametype` ORDER BY name, timestamp DESC;",
			"SELECT DISTINCT `order` FROM `gametype`;",
			"SELECT * FROM `gametype`;",
		}},
	}

	for _, test := range tests {
		t.Run(string(test.dialect), func(t *testing.T) {
			sql := NewGenerator(test.dialect).generateCRUDQueries(distinctEntity())

			for _, want := range test.want {
				if !strings.Contains(sql, want) {
					t.Errorf("expected %s in the generated SQL:\n%s", want, sql)
				}
			}

			// only the four distinct queries dedupe, the plain list still selects rows
			if got := strings.Count(sql, "SELECT DISTINCT"); got != 4 {
				t.Errorf("expected 4 distinct selects, got %d:\n%s", got, sql)
			}
		})
	}
}

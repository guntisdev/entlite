package sqlc

import (
	"strings"
	"testing"

	"github.com/guntisdev/entlite/internal/schema"
)

func aggregateEntity() schema.Entity {
	contracts := []schema.Contract{{Type: schema.ContractSQLC}, {Type: schema.ContractPROTO}}

	return schema.Entity{
		Name:      "Build",
		Contracts: contracts,
		Fields: []schema.Field{
			{Name: "id", Type: schema.FieldTypeInt, Primary: true, Contracts: contracts},
			{Name: "branch", Type: schema.FieldTypeString, Contracts: contracts},
			{Name: "env", Type: schema.FieldTypeString, Contracts: contracts},
			{Name: "duration_ms", Type: schema.FieldTypeInt64, Contracts: contracts},
			{Name: "coverage", Type: schema.FieldTypeFloat, Contracts: contracts},
			{Name: "order", Type: schema.FieldTypeInt, Contracts: contracts},
			{Name: "started_at", Type: schema.FieldTypeTime, Contracts: contracts},
		},
		Queries: []schema.Query{
			// the use case: one total per group, filtered and sorted
			{Type: schema.QueryListBy, Name: "SumBranchDurations",
				Filters:    []schema.QueryFilter{{Field: "env", Type: schema.QueryFilterEq}},
				GroupBy:    []string{"branch"},
				Aggregates: []schema.Aggregate{{Func: schema.AggregateSum, Field: "duration_ms"}},
				OrderBy:    []schema.OrderColumn{{Name: "branch"}}, Contracts: contracts},
			// several aggregates over several columns, in chain order
			{Type: schema.QueryListAll, Name: "BranchStats",
				GroupBy: []string{"branch", "env"},
				Aggregates: []schema.Aggregate{
					{Func: schema.AggregateAvg, Field: "duration_ms"},
					{Func: schema.AggregateMin, Field: "started_at"},
					{Func: schema.AggregateMax, Field: "started_at"},
				}, Contracts: contracts},
			// no group folds the whole table into one row
			{Type: schema.QueryListAll, Name: "TotalDuration",
				Aggregates: []schema.Aggregate{{Func: schema.AggregateSum, Field: "duration_ms"}}, Contracts: contracts},
			// a sum of floats stays a float
			{Type: schema.QueryListAll, Name: "TotalCoverage",
				Aggregates: []schema.Aggregate{{Func: schema.AggregateSum, Field: "coverage"}}, Contracts: contracts},
			// a reserved column name still gets quoted
			{Type: schema.QueryListAll, Name: "MaxOrder",
				Aggregates: []schema.Aggregate{{Func: schema.AggregateMax, Field: "order"}}, Contracts: contracts},
			// paging the groups
			{Type: schema.QueryListAll, Name: "TopBranches",
				GroupBy:    []string{"branch"},
				Aggregates: []schema.Aggregate{{Func: schema.AggregateSum, Field: "duration_ms"}},
				HasLimit:   true, HasOffset: true, Contracts: contracts},
		},
	}
}

func TestAggregateSelect(t *testing.T) {
	tests := []struct {
		dialect schema.SQLDialect
		want    []string
	}{
		{schema.PostgreSQL, []string{
			`SELECT branch, CAST(SUM(duration_ms) AS BIGINT) AS sum_duration_ms FROM "build" WHERE env = @env GROUP BY branch ORDER BY branch;`,
			`SELECT branch, env, CAST(AVG(duration_ms) AS DOUBLE PRECISION) AS avg_duration_ms, MIN(started_at) AS min_started_at, MAX(started_at) AS max_started_at FROM "build" GROUP BY branch, env;`,
			`SELECT CAST(SUM(duration_ms) AS BIGINT) AS sum_duration_ms FROM "build";`,
			`SELECT CAST(SUM(coverage) AS DOUBLE PRECISION) AS sum_coverage FROM "build";`,
			`SELECT MAX("order") AS max_order FROM "build";`,
			`SELECT branch, CAST(SUM(duration_ms) AS BIGINT) AS sum_duration_ms FROM "build" GROUP BY branch LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');`,
		}},
		{schema.SQLite, []string{
			`SELECT branch, CAST(SUM(duration_ms) AS INTEGER) AS sum_duration_ms FROM "build" WHERE env = @env GROUP BY branch ORDER BY branch;`,
			`SELECT branch, env, CAST(AVG(duration_ms) AS REAL) AS avg_duration_ms, MIN(started_at) AS min_started_at, MAX(started_at) AS max_started_at FROM "build" GROUP BY branch, env;`,
			`SELECT CAST(SUM(coverage) AS REAL) AS sum_coverage FROM "build";`,
			`SELECT MAX("order") AS max_order FROM "build";`,
		}},
		{schema.MySQL, []string{
			"SELECT branch, CAST(SUM(duration_ms) AS SIGNED) AS sum_duration_ms FROM `build` WHERE env = sqlc.arg('env') GROUP BY branch ORDER BY branch;",
			"SELECT branch, env, CAST(AVG(duration_ms) AS DOUBLE) AS avg_duration_ms, MIN(started_at) AS min_started_at, MAX(started_at) AS max_started_at FROM `build` GROUP BY branch, env;",
			"SELECT CAST(SUM(coverage) AS DOUBLE) AS sum_coverage FROM `build`;",
			"SELECT MAX(`order`) AS max_order FROM `build`;",
		}},
	}

	for _, test := range tests {
		t.Run(string(test.dialect), func(t *testing.T) {
			sql := NewGenerator(test.dialect).generateCRUDQueries(aggregateEntity())

			for _, want := range test.want {
				if !strings.Contains(sql, want) {
					t.Errorf("expected %s in the generated SQL:\n%s", want, sql)
				}
			}

			// a group returns many rows, folding the whole table returns one
			if got := strings.Count(sql, ":one"); got != 3 {
				t.Errorf("expected 3 one row queries, got %d:\n%s", got, sql)
			}
			if got := strings.Count(sql, "GROUP BY"); got != 3 {
				t.Errorf("expected 3 grouped queries, got %d:\n%s", got, sql)
			}
		})
	}
}

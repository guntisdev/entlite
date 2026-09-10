package parser

import (
	"strings"
	"testing"

	"github.com/guntisdev/entlite/internal/schema"
)

func TestParseQueryAggregates(t *testing.T) {
	tests := []struct {
		name    string
		queries string
		want    []schema.Aggregate
		wantErr string
	}{
		{
			name:    "no aggregate",
			queries: `query.ListBy("email"),`,
		},
		{
			name:    "sum",
			queries: `query.ListAll().GroupBy("env").Sum("age").Name("Folded"),`,
			want:    []schema.Aggregate{{Func: schema.AggregateSum, Field: "age"}},
		},
		{
			name:    "avg",
			queries: `query.ListAll().GroupBy("env").Avg("age").Name("Folded"),`,
			want:    []schema.Aggregate{{Func: schema.AggregateAvg, Field: "age"}},
		},
		{
			name:    "min and max keep the chain order",
			queries: `query.ListAll().GroupBy("env").Max("age").Min("age").Name("Folded"),`,
			want: []schema.Aggregate{
				{Func: schema.AggregateMax, Field: "age"},
				{Func: schema.AggregateMin, Field: "age"},
			},
		},
		{
			name:    "several aggregates over several columns",
			queries: `query.ListBy("env").Sum("age").Avg("score").Name("Folded"),`,
			want: []schema.Aggregate{
				{Func: schema.AggregateSum, Field: "age"},
				{Func: schema.AggregateAvg, Field: "score"},
			},
		},
		{
			name:    "aggregate on a non list query",
			queries: `query.Get().Sum("age").Name("Folded"),`,
			wantErr: "Sum is only supported for list queries",
		},
		{
			name:    "aggregate needs a column",
			queries: `query.ListAll().Avg().Name("Folded"),`,
			wantErr: "Avg expects exactly one string field",
		},
		{
			name:    "aggregate takes one column",
			queries: `query.ListAll().Min("age", "score").Name("Folded"),`,
			wantErr: "Min expects exactly one string field",
		},
		{
			name:    "aggregate takes a string",
			queries: `query.ListAll().Max(3).Name("Folded"),`,
			wantErr: "Max expects exactly one string field",
		},
		{
			name:    "nonexisting column",
			queries: `query.ListAll().Sum("nope").Name("Folded"),`,
			wantErr: `Sum() references nonexisting field "nope"`,
		},
		{
			name:    "sum of a string",
			queries: `query.ListAll().Sum("email").Name("Folded"),`,
			wantErr: `Sum() cannot fold field "email" of type string`,
		},
		{
			name:    "avg of a time",
			queries: `query.ListAll().Avg("seen_at").Name("Folded"),`,
			wantErr: `Avg() cannot fold field "seen_at" of type time`,
		},
		{
			name:    "min of a string orders it",
			queries: `query.ListAll().Min("email").Name("Folded"),`,
			want:    []schema.Aggregate{{Func: schema.AggregateMin, Field: "email"}},
		},
		{
			name:    "max of a time",
			queries: `query.ListAll().Max("seen_at").Name("Folded"),`,
			wantErr: `Max() cannot fold field "seen_at" of type time`,
		},
		{
			name:    "max of a bool",
			queries: `query.ListAll().Max("active").Name("Folded"),`,
			wantErr: `Max() cannot fold field "active" of type bool`,
		},
		{
			name:    "the same aggregate twice",
			queries: `query.ListAll().Sum("age").Sum("age").Name("Folded"),`,
			wantErr: `repeats Sum() of field "age"`,
		},
		{
			name:    "folding the grouped column",
			queries: `query.ListAll().GroupBy("env").Sum("env").Name("Folded"),`,
			wantErr: `Sum() cannot fold field "env" of type string`,
		},
		{
			name:    "folding a column the group already holds",
			queries: `query.ListAll().GroupBy("age").Sum("age").Name("Folded"),`,
			wantErr: `Sum() folds field "age", which GroupBy() already groups by`,
		},
		{
			name:    "paging one row",
			queries: `query.ListAll().Sum("age").Limit().Name("Folded"),`,
			wantErr: "has Limit() on an aggregate without GroupBy()",
		},
		{
			name:    "sorting one row",
			queries: `query.ListAll().Sum("age").Desc("age").Name("Folded"),`,
			wantErr: "sorts an aggregate without GroupBy()",
		},
		{
			name:    "aggregate without a name",
			queries: `query.ListAll().Sum("age"),`,
			wantErr: "has an aggregate without Name()",
		},
		{
			name:    "aggregate with distinct",
			queries: `query.ListAll().Distinct("env").Sum("age").Name("Folded"),`,
			wantErr: "has an aggregate with Distinct()",
		},
	}

	const fields = `field.Int("age"),
		field.Float("score"),
		field.Bool("active"),
		field.Time("seen_at"),`

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			entity, err := parseDistinctEntity(t, bothContracts, fields, test.queries)

			if test.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), test.wantErr) {
					t.Fatalf("expected error containing %q, got: %v", test.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}

			got := entity.Queries[0].Aggregates
			if len(got) != len(test.want) {
				t.Fatalf("Aggregates: expected %v, got %v", test.want, got)
			}
			for i := range got {
				if got[i] != test.want[i] {
					t.Errorf("Aggregates: expected %v, got %v", test.want, got)
				}
			}
		})
	}
}

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
			queries: `query.ListAll().GroupBy("env").Sum("age"),`,
			want:    []schema.Aggregate{{Func: schema.AggregateSum, Field: "age"}},
		},
		{
			name:    "avg",
			queries: `query.ListAll().GroupBy("env").Avg("age"),`,
			want:    []schema.Aggregate{{Func: schema.AggregateAvg, Field: "age"}},
		},
		{
			name:    "min and max keep the chain order",
			queries: `query.ListAll().GroupBy("env").Max("age").Min("age"),`,
			want: []schema.Aggregate{
				{Func: schema.AggregateMax, Field: "age"},
				{Func: schema.AggregateMin, Field: "age"},
			},
		},
		{
			name:    "several aggregates over several columns",
			queries: `query.ListBy("env").Sum("age").Avg("score"),`,
			want: []schema.Aggregate{
				{Func: schema.AggregateSum, Field: "age"},
				{Func: schema.AggregateAvg, Field: "score"},
			},
		},
		{
			name:    "aggregate on a non list query",
			queries: `query.Get().Sum("age"),`,
			wantErr: "Sum is only supported for list queries",
		},
		{
			name:    "aggregate needs a column",
			queries: `query.ListAll().Avg(),`,
			wantErr: "Avg expects exactly one string field",
		},
		{
			name:    "aggregate takes one column",
			queries: `query.ListAll().Min("age", "score"),`,
			wantErr: "Min expects exactly one string field",
		},
		{
			name:    "aggregate takes a string",
			queries: `query.ListAll().Max(3),`,
			wantErr: "Max expects exactly one string field",
		},
	}

	const fields = `field.Int("age"),
		field.Float("score"),`

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

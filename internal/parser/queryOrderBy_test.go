package parser

import (
	"strings"
	"testing"

	"github.com/guntisdev/entlite/internal/schema"
)

func TestParseQueryOrderBy(t *testing.T) {
	tests := []struct {
		name    string
		queries string
		want    []schema.OrderColumn
		wantErr string
	}{
		{
			name:    "no order by",
			queries: `query.ListBy("email"),`,
		},
		{
			name:    "ascending",
			queries: `query.ListBy("email").Asc("name"),`,
			want:    []schema.OrderColumn{{Name: "name"}},
		},
		{
			name:    "descending",
			queries: `query.ListBy("email").Desc("name"),`,
			want:    []schema.OrderColumn{{Name: "name", Desc: true}},
		},
		{
			name:    "two columns keep the chain order",
			queries: `query.ListBy("email").Desc("name").Asc("email"),`,
			want:    []schema.OrderColumn{{Name: "name", Desc: true}, {Name: "email"}},
		},
		{
			name:    "sorted with pagination",
			queries: `query.ListBy("email").Desc("email").Limit().Offset(),`,
			want:    []schema.OrderColumn{{Name: "email", Desc: true}},
		},
		{
			name:    "sorted by a filter field",
			queries: `query.ListBy(filter.Search("email")).Asc("email"),`,
			want:    []schema.OrderColumn{{Name: "email"}},
		},
		{
			name:    "sorted list all",
			queries: `query.ListAll().Desc("name").Limit(),`,
			want:    []schema.OrderColumn{{Name: "name", Desc: true}},
		},
		{
			name:    "unknown field",
			queries: `query.ListBy("email").Asc("created_at"),`,
			wantErr: `order by references nonexisting field "created_at"`,
		},
		{
			name:    "repeated field",
			queries: `query.ListBy("email").Asc("name").Desc("name"),`,
			wantErr: `order by repeats field "name"`,
		},
		{
			name:    "empty field name",
			queries: `query.ListBy("email").Asc(""),`,
			wantErr: "Asc expects a field name",
		},
		{
			name:    "no argument",
			queries: `query.ListBy("email").Asc(),`,
			wantErr: "Asc expects exactly one string field",
		},
		{
			name:    "two arguments",
			queries: `query.ListBy("email").Desc("email", "name"),`,
			wantErr: "Desc expects exactly one string field",
		},
		{
			name:    "not a string",
			queries: `query.ListBy("email").Desc(1),`,
			wantErr: "Desc expects exactly one string field",
		},
		{
			name:    "order by on a non list query",
			queries: `query.Get().Asc("email"),`,
			wantErr: "Asc is only supported for list queries",
		},
		{
			name:    "unknown operation",
			queries: `query.ListBy("email").Sort("email"),`,
			wantErr: `unsupported query operation "Sort"`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			entity, err := parseQueryContractEntity(t, bothContracts, test.queries)

			if test.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), test.wantErr) {
					t.Fatalf("expected error containing %q, got: %v", test.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}

			got := entity.Queries[0].OrderBy
			if len(got) != len(test.want) {
				t.Fatalf("OrderBy: expected %v, got %v", test.want, got)
			}
			for i := range got {
				if got[i] != test.want[i] {
					t.Errorf("OrderBy column %d: expected %v, got %v", i, test.want[i], got[i])
				}
			}
		})
	}
}

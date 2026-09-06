package parser

import (
	"strings"
	"testing"
)

func TestParseQueryOrderBy(t *testing.T) {
	tests := []struct {
		name    string
		queries string
		want    string
		wantErr string
	}{
		{
			name:    "no order by",
			queries: `query.ListBy("email"),`,
		},
		{
			name:    "sorted by a field",
			queries: `query.ListBy("email").OrderBy("email"),`,
			want:    "email",
		},
		{
			name:    "sorted with pagination",
			queries: `query.ListBy("email").OrderBy("email").Limit().Offset(),`,
			want:    "email",
		},
		{
			name:    "sorted by a filter field",
			queries: `query.ListBy(filter.Search("email")).OrderBy("email"),`,
			want:    "email",
		},
		{
			name:    "unknown field",
			queries: `query.ListBy("email").OrderBy("created_at"),`,
			wantErr: `order_by references nonexisting field "created_at"`,
		},
		{
			name:    "empty field name",
			queries: `query.ListBy("email").OrderBy(""),`,
			wantErr: "OrderBy expects a field name",
		},
		{
			name:    "no argument",
			queries: `query.ListBy("email").OrderBy(),`,
			wantErr: "OrderBy expects exactly one string field",
		},
		{
			name:    "two arguments",
			queries: `query.ListBy("email").OrderBy("email", "email"),`,
			wantErr: "OrderBy expects exactly one string field",
		},
		{
			name:    "not a string",
			queries: `query.ListBy("email").OrderBy(1),`,
			wantErr: "OrderBy expects exactly one string field",
		},
		{
			name:    "order by on a non list query",
			queries: `query.Get().OrderBy("email"),`,
			wantErr: "OrderBy is only supported for ListBy queries",
		},
		{
			name:    "order by on list all",
			queries: `query.ListAll().OrderBy("email"),`,
			wantErr: "OrderBy is only supported for ListBy queries",
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

			if got := entity.Queries[0].OrderBy; got != test.want {
				t.Errorf("OrderBy: expected %q, got %q", test.want, got)
			}
		})
	}
}

package parser

import (
	"strings"
	"testing"
)

func TestParseQueryCount(t *testing.T) {
	tests := []struct {
		name      string
		queries   string
		wantCount bool
		wantErr   string
	}{
		{
			name:    "no count",
			queries: `query.ListBy("email"),`,
		},
		{
			name:      "list by with count",
			queries:   `query.ListBy("email").Count(),`,
			wantCount: true,
		},
		{
			name:      "list all with count",
			queries:   `query.ListAll().Count(),`,
			wantCount: true,
		},
		{
			name:      "count next to pagination",
			queries:   `query.ListBy("email").Asc("email").Limit().Offset().Count(),`,
			wantCount: true,
		},
		{
			name:      "count before pagination",
			queries:   `query.ListBy("email").Count().Limit(),`,
			wantCount: true,
		},
		{
			name:    "count on a non list query",
			queries: `query.Get().Count(),`,
			wantErr: "Count is only supported for list queries",
		},
		{
			name:    "count takes no arguments",
			queries: `query.ListBy("email").Count("email"),`,
			wantErr: "Count does not accept arguments",
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

			if got := entity.Queries[0].Count; got != test.wantCount {
				t.Errorf("Count: expected %v, got %v", test.wantCount, got)
			}
		})
	}
}

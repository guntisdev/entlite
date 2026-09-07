package parser

import (
	"strings"
	"testing"
)

func TestParseQueryPagination(t *testing.T) {
	tests := []struct {
		name      string
		queries   string
		wantLimit bool
		wantRows  int
		wantOff   bool
		wantErr   string
	}{
		{
			name:    "no pagination",
			queries: `query.ListBy("email"),`,
		},
		{
			name:      "caller sets the limit",
			queries:   `query.ListBy("email").Limit(),`,
			wantLimit: true,
		},
		{
			name:      "fixed limit",
			queries:   `query.ListBy("email").Limit(50),`,
			wantLimit: true,
			wantRows:  50,
		},
		{
			name:      "limit and offset",
			queries:   `query.ListBy("email").Limit().Offset(),`,
			wantLimit: true,
			wantOff:   true,
		},
		{
			name:      "fixed page size, caller sets the page",
			queries:   `query.ListBy("email").Limit(100).Offset(),`,
			wantLimit: true,
			wantRows:  100,
			wantOff:   true,
		},
		{
			name:      "list all takes pagination too",
			queries:   `query.ListAll().Limit(10),`,
			wantLimit: true,
			wantRows:  10,
		},
		{
			name:      "chained after other operations",
			queries:   `query.ListBy("email").Asc("email").Limit().Offset(),`,
			wantLimit: true,
			wantOff:   true,
		},
		{
			name:    "offset without limit",
			queries: `query.ListBy("email").Offset(),`,
			wantErr: "has Offset() without Limit()",
		},
		{
			name:    "limit on a non list query",
			queries: `query.Get().Limit(),`,
			wantErr: "Limit is only supported for list queries",
		},
		{
			name:    "offset takes no arguments",
			queries: `query.ListBy("email").Offset(10),`,
			wantErr: "Offset does not accept arguments",
		},
		{
			name:    "limit takes at most one argument",
			queries: `query.ListBy("email").Limit(10, 20),`,
			wantErr: "Limit expects no arguments or a single row count",
		},
		{
			name:    "limit rejects a non number",
			queries: `query.ListBy("email").Limit("10"),`,
			wantErr: "Limit expects no arguments or a single row count",
		},
		{
			name:    "limit rejects zero",
			queries: `query.ListBy("email").Limit(0),`,
			wantErr: "Limit 0 must be at least 1",
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

			query := entity.Queries[0]
			if query.HasLimit != test.wantLimit {
				t.Errorf("HasLimit: expected %v, got %v", test.wantLimit, query.HasLimit)
			}
			if query.Limit != test.wantRows {
				t.Errorf("Limit: expected %d, got %d", test.wantRows, query.Limit)
			}
			if query.HasOffset != test.wantOff {
				t.Errorf("HasOffset: expected %v, got %v", test.wantOff, query.HasOffset)
			}
			// a fixed limit stays out of the request
			wantFromRequest := test.wantLimit && test.wantRows == 0
			if query.LimitFromRequest() != wantFromRequest {
				t.Errorf("LimitFromRequest: expected %v, got %v", wantFromRequest, query.LimitFromRequest())
			}
		})
	}
}

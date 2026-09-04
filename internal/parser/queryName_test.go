package parser

import (
	"strings"
	"testing"
)

func TestParseQueryName(t *testing.T) {
	tests := []struct {
		name    string
		queries string
		wantErr string
	}{
		{
			name:    "custom name",
			queries: `query.ListAll().Name("ListActive"),`,
		},
		{
			name:    "not an identifier",
			queries: `query.ListAll().Name("list active"),`,
			wantErr: `Name "list active" is not a valid identifier`,
		},
		{
			name:    "reserved request suffix",
			queries: `query.ListAll().Name("ListActiveRequest"),`,
			wantErr: `Name "ListActiveRequest" cannot end with Request, the generator appends it`,
		},
		{
			name:    "reserved params suffix",
			queries: `query.ListAll().Name("ListActiveParams"),`,
			wantErr: `Name "ListActiveParams" cannot end with Params, the generator appends it`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			entity, err := parseQueryContractEntity(t, "entlite.SQLC(), entlite.PROTO(),", test.queries)

			if test.wantErr == "" {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				if got := entity.Queries[0].Name; got != "ListActive" {
					t.Errorf("expected query name ListActive, got %q", got)
				}
				return
			}

			if err == nil {
				t.Fatalf("expected error containing %q, got none", test.wantErr)
			}
			if !strings.Contains(err.Error(), test.wantErr) {
				t.Errorf("expected error containing %q, got %q", test.wantErr, err.Error())
			}
		})
	}
}

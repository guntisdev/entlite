package parser

import (
	"strings"
	"testing"
)

// validation of the grouped columns lives with the aggregates, this only covers parsing
func TestParseQueryGroupBy(t *testing.T) {
	tests := []struct {
		name    string
		queries string
		want    []string
		wantErr string
	}{
		{
			name:    "no group by",
			queries: `query.ListBy("email"),`,
		},
		{
			name:    "one column",
			queries: `query.ListBy("env").GroupBy("name"),`,
			want:    []string{"name"},
		},
		{
			name:    "several columns keep the given order",
			queries: `query.ListAll().GroupBy("name", "env"),`,
			want:    []string{"name", "env"},
		},
		{
			name:    "group by on a non list query",
			queries: `query.Get().GroupBy("name"),`,
			wantErr: "GroupBy is only supported for list queries",
		},
		{
			name:    "group by needs a column",
			queries: `query.ListAll().GroupBy(),`,
			wantErr: "GroupBy expects at least one field name",
		},
		{
			name:    "group by takes strings",
			queries: `query.ListAll().GroupBy(3),`,
			wantErr: "GroupBy expects string field args",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			entity, err := parseDistinctEntity(t, bothContracts, "", test.queries)

			if test.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), test.wantErr) {
					t.Fatalf("expected error containing %q, got: %v", test.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}

			got := entity.Queries[0].GroupBy
			if len(got) != len(test.want) {
				t.Fatalf("GroupBy: expected %v, got %v", test.want, got)
			}
			for i := range got {
				if got[i] != test.want[i] {
					t.Errorf("GroupBy: expected %v, got %v", test.want, got)
				}
			}
		})
	}
}

package parser

import (
	"strings"
	"testing"
)

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
			queries: `query.ListBy("env").GroupBy("name").Sum("age").Name("Folded"),`,
			want:    []string{"name"},
		},
		{
			name:    "several columns keep the given order",
			queries: `query.ListAll().GroupBy("name", "env").Sum("age").Name("Folded"),`,
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
		{
			name:    "group by without an aggregate is a distinct",
			queries: `query.ListAll().GroupBy("name"),`,
			wantErr: "has GroupBy() without an aggregate",
		},
		{
			name:    "group by together with distinct",
			queries: `query.ListAll().Distinct("env").GroupBy("name").Sum("age").Name("Folded"),`,
			wantErr: "has both Distinct() and GroupBy()",
		},
		{
			name:    "group by with count",
			queries: `query.ListAll().GroupBy("name").Sum("age").Count().Name("Folded"),`,
			wantErr: "has an aggregate with Count()",
		},
		{
			name:    "nonexisting column",
			queries: `query.ListAll().GroupBy("nope").Sum("age").Name("Folded"),`,
			wantErr: `GroupBy references nonexisting field "nope"`,
		},
		{
			name:    "repeated column",
			queries: `query.ListAll().GroupBy("name", "name").Sum("age").Name("Folded"),`,
			wantErr: `GroupBy repeats field "name"`,
		},
		{
			name:    "the primary key makes every group hold one row",
			queries: `query.ListAll().GroupBy("id", "name").Sum("age").Name("Folded"),`,
			wantErr: "GroupBy groups by the unique key (id)",
		},
		{
			name:    "a unique column makes every group hold one row",
			queries: `query.ListAll().GroupBy("email").Sum("age").Name("Folded"),`,
			wantErr: "GroupBy groups by the unique key (email)",
		},
		{
			name:    "sorted by a column the group does not hold",
			queries: `query.ListAll().GroupBy("name").Sum("age").Asc("env").Name("Folded"),`,
			wantErr: `sorts by field "env", which GroupBy() does not group by`,
		},
		{
			name:    "sorted and paged by the grouped column",
			queries: `query.ListAll().GroupBy("name").Sum("age").Asc("name").Limit().Offset().Name("Folded"),`,
			want:    []string{"name"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			entity, err := parseDistinctEntity(t, bothContracts, `field.Int("age"),`, test.queries)

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

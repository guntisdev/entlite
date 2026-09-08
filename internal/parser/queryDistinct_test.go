package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/guntisdev/entlite/internal/schema"
)

const distinctTemplate = `package schema

import (
	"github.com/guntisdev/entlite/pkg/entlite"
	"github.com/guntisdev/entlite/pkg/entlite/field"
	"github.com/guntisdev/entlite/pkg/entlite/query"
)

type User struct {
	entlite.Schema
}

func (User) Contracts() []entlite.Contract {
	return []entlite.Contract{
		%s
	}
}

func (User) Fields() []entlite.Field {
	return []entlite.Field{
		field.String("email").Unique(),
		field.String("name"),
		field.String("env"),
		%s
	}
}

func (User) Queries() []entlite.Query {
	return []entlite.Query{
		%s
	}
}
`

func parseDistinctEntity(t *testing.T, contracts, fields, queries string) (schema.Entity, error) {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "user.go")
	source := strings.Replace(distinctTemplate, "%s", contracts, 1)
	source = strings.Replace(source, "%s", fields, 1)
	source = strings.Replace(source, "%s", queries, 1)

	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		t.Fatalf("failed to write entity file: %v", err)
	}

	entities, err := ParseEntities([]DiscoveredEntity{{Name: "User", Path: path}})
	if err != nil {
		return schema.Entity{}, err
	}

	return entities[0], nil
}

func TestParseQueryDistinct(t *testing.T) {
	tests := []struct {
		name    string
		queries string
		want    []string
		wantErr string
	}{
		{
			name:    "no distinct",
			queries: `query.ListBy("email"),`,
		},
		{
			name:    "one column",
			queries: `query.ListBy("env").Distinct("name"),`,
			want:    []string{"name"},
		},
		{
			name:    "several columns keep the given order",
			queries: `query.ListAll().Distinct("name", "env"),`,
			want:    []string{"name", "env"},
		},
		{
			name:    "sorted by the distinct column",
			queries: `query.ListAll().Distinct("name").Asc("name"),`,
			want:    []string{"name"},
		},
		{
			name:    "paged distinct",
			queries: `query.ListBy("env").Distinct("name").Limit().Offset(),`,
			want:    []string{"name"},
		},
		{
			name:    "distinct on a non list query",
			queries: `query.Get().Distinct("name"),`,
			wantErr: "Distinct is only supported for list queries",
		},
		{
			name:    "distinct needs a column",
			queries: `query.ListAll().Distinct(),`,
			wantErr: "Distinct expects at least one field name",
		},
		{
			name:    "distinct takes strings",
			queries: `query.ListAll().Distinct(3),`,
			wantErr: "Distinct expects string field args",
		},
		{
			name:    "nonexisting column",
			queries: `query.ListAll().Distinct("nope"),`,
			wantErr: `Distinct references nonexisting field "nope"`,
		},
		{
			name:    "repeated column",
			queries: `query.ListAll().Distinct("name", "name"),`,
			wantErr: `Distinct repeats field "name"`,
		},
		{
			name:    "count would total the rows the dedupe drops",
			queries: `query.ListAll().Distinct("name").Count(),`,
			wantErr: "has Distinct() with Count()",
		},
		{
			name:    "sorted by a column the select does not hold",
			queries: `query.ListAll().Distinct("name").Asc("env"),`,
			wantErr: `sorts by field "env", which Distinct() does not select`,
		},
		{
			name:    "the primary key makes every row unique",
			queries: `query.ListAll().Distinct("id", "name"),`,
			wantErr: "Distinct selects the unique key (id)",
		},
		{
			name:    "a unique column makes every row unique",
			queries: `query.ListAll().Distinct("email"),`,
			wantErr: "Distinct selects the unique key (email)",
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

			got := entity.Queries[0].Distinct
			if len(got) != len(test.want) {
				t.Fatalf("Distinct: expected %v, got %v", test.want, got)
			}
			for i := range got {
				if got[i] != test.want[i] {
					t.Errorf("Distinct: expected %v, got %v", test.want, got)
				}
			}
		})
	}
}

// a unique index over several columns cannot be deduplicated either
func TestDistinctOnUniqueIndex(t *testing.T) {
	const source = `package schema

import (
	"github.com/guntisdev/entlite/pkg/entlite"
	"github.com/guntisdev/entlite/pkg/entlite/field"
	"github.com/guntisdev/entlite/pkg/entlite/index"
	"github.com/guntisdev/entlite/pkg/entlite/query"
)

type User struct {
	entlite.Schema
}

func (User) Contracts() []entlite.Contract {
	return []entlite.Contract{
		entlite.SQLC(),
		entlite.PROTO(),
	}
}

func (User) Fields() []entlite.Field {
	return []entlite.Field{
		field.String("name"),
		field.String("env"),
	}
}

func (User) Indexes() []entlite.Index {
	return []entlite.Index{
		index.Asc("name", "env").Unique(),
	}
}

func (User) Queries() []entlite.Query {
	return []entlite.Query{
		query.ListAll().Distinct("name", "env"),
	}
}
`

	dir := t.TempDir()
	path := filepath.Join(dir, "user.go")
	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		t.Fatalf("failed to write entity file: %v", err)
	}

	_, err := ParseEntities([]DiscoveredEntity{{Name: "User", Path: path}})
	want := "Distinct selects the unique key (name, env)"
	if err == nil || !strings.Contains(err.Error(), want) {
		t.Fatalf("expected error containing %q, got: %v", want, err)
	}
}

// the distinct columns are part of the generated name, so two distinct queries on the
// same fields do not collide
func TestDistinctQueryNames(t *testing.T) {
	entity, err := parseDistinctEntity(t, bothContracts, "",
		"query.ListBy(\"env\"),\n\t\tquery.ListBy(\"env\").Distinct(\"name\"),\n\t\tquery.ListAll().Distinct(\"name\", \"env\"),")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	want := []string{"ListUserByEnv", "ListUserDistinctNameByEnv", "ListAllUserDistinctNameEnv"}
	for i, name := range want {
		if entity.Queries[i].Name != name {
			t.Errorf("query %d: expected name %q, got %q", i, name, entity.Queries[i].Name)
		}
	}
}

// a distinct select returns the values, so it cannot name a field the api hides
func TestDistinctOnUnreadableField(t *testing.T) {
	fields := `field.String("password").Contracts(entlite.SQLC(), entlite.PROTO().WriteOnly()),`
	_, err := parseDistinctEntity(t, bothContracts, fields, `query.ListAll().Distinct("password"),`)

	want := `Distinct returns field "password", which the proto contract cannot read`
	if err == nil || !strings.Contains(err.Error(), want) {
		t.Fatalf("expected error containing %q, got: %v", want, err)
	}
}

// without a proto contract the value never leaves the database layer
func TestDistinctOnUnreadableFieldSQLCOnly(t *testing.T) {
	fields := `field.String("password").Contracts(entlite.SQLC()),`
	_, err := parseDistinctEntity(t, "entlite.SQLC(),", fields, `query.ListAll().Distinct("password"),`)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

// a virtual field has no column to select
func TestDistinctOnVirtualField(t *testing.T) {
	fields := `field.String("token").Contracts(entlite.PROTO()),`
	_, err := parseDistinctEntity(t, bothContracts, fields, `query.ListAll().Distinct("token"),`)

	want := `Distinct references virtual field "token"`
	if err == nil || !strings.Contains(err.Error(), want) {
		t.Fatalf("expected error containing %q, got: %v", want, err)
	}
}

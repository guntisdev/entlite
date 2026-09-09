package parser

import (
	"os"
	"path/filepath"
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

const generatedNameTemplate = `package schema

import (
	"github.com/guntisdev/entlite/pkg/entlite"
	"github.com/guntisdev/entlite/pkg/entlite/field"
	"github.com/guntisdev/entlite/pkg/entlite/filter"
	"github.com/guntisdev/entlite/pkg/entlite/query"
)

type User struct {
	entlite.Schema
}

func (User) Contracts() []entlite.Contract {
	return []entlite.Contract{
		entlite.SQLC(), entlite.PROTO(),
	}
}

func (User) Fields() []entlite.Field {
	return []entlite.Field{
		field.String("email").Unique(),
		field.Int("org_id"),
		field.Int("age"),
	}
}

func (User) Queries() []entlite.Query {
	return []entlite.Query{
		%s
	}
}
`

func TestGeneratedQueryName(t *testing.T) {
	tests := []struct {
		name    string
		queries string
		want    []string
	}{
		{
			name:    "default crud",
			queries: `query.DefaultCRUD(),`,
			want:    []string{"CreateUser", "GetUserById", "UpdateUser", "DeleteUser"},
		},
		{
			name:    "create bulk and delete all",
			queries: `query.CreateBulk(), query.DeleteAll(),`,
			want:    []string{"CreateBulkUser", "DeleteAllUser"},
		},
		{
			name:    "get by fields",
			queries: `query.GetBy("email"), query.GetBy("org_id", "email"),`,
			want:    []string{"GetUserByEmail", "GetUserByOrgIdEmail"},
		},
		{
			name:    "list queries",
			queries: `query.ListAll(), query.ListBy("email"), query.ListBy(filter.Range("age")),`,
			want:    []string{"ListAllUser", "ListUserByEmail", "ListUserFilterByAge"},
		},
		{
			name:    "custom name wins",
			queries: `query.ListAll().Name("ListActive"),`,
			want:    []string{"ListActive"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "user.go")
			source := strings.Replace(generatedNameTemplate, "%s", test.queries, 1)
			if err := os.WriteFile(path, []byte(source), 0644); err != nil {
				t.Fatalf("failed to write entity file: %v", err)
			}

			entities, err := ParseEntities([]DiscoveredEntity{{Name: "User", Path: path}})
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			var got []string
			for _, query := range entities[0].Queries {
				got = append(got, query.Name)
			}

			if len(got) != len(test.want) {
				t.Fatalf("expected names %v, got %v", test.want, got)
			}
			for i := range got {
				if got[i] != test.want[i] {
					t.Errorf("expected name %q, got %q", test.want[i], got[i])
				}
			}
		})
	}
}

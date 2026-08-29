package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/guntisdev/entlite/internal/schema"
)

const queryContractTemplate = `package schema

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
	}
}

func (User) Queries() []entlite.Query {
	return []entlite.Query{
		%s
	}
}
`

func parseQueryContractEntity(t *testing.T, contracts, queries string) (schema.Entity, error) {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "user.go")
	source := strings.Replace(queryContractTemplate, "%s", contracts, 1)
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

func queryTypes(queries []schema.Query) []schema.QueryType {
	types := make([]schema.QueryType, 0, len(queries))
	for _, q := range queries {
		types = append(types, q.Type)
	}

	return types
}

func equalTypes(got, want []schema.QueryType) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}

	return true
}

const bothContracts = "entlite.SQLC(),\n\t\tentlite.PROTO(),"

func TestQueryContracts(t *testing.T) {
	tests := []struct {
		name      string
		contracts string
		queries   string
		wantSQLC  []schema.QueryType
		wantProto []schema.QueryType
		wantErr   string
	}{
		{
			name:      "no contracts inherits the entity",
			contracts: bothContracts,
			queries:   "query.Create(),\n\t\tquery.ListAll(),",
			wantSQLC:  []schema.QueryType{schema.QueryCreate, schema.QueryListAll},
			wantProto: []schema.QueryType{schema.QueryCreate, schema.QueryListAll},
		},
		{
			name:      "sqlc only query has no rpc",
			contracts: bothContracts,
			queries:   "query.Create().Contracts(entlite.SQLC()),\n\t\tquery.ListAll(),",
			wantSQLC:  []schema.QueryType{schema.QueryCreate, schema.QueryListAll},
			wantProto: []schema.QueryType{schema.QueryListAll},
		},
		{
			name:      "proto only query has no sql",
			contracts: bothContracts,
			queries:   "query.Create(),\n\t\tquery.ListAll().Contracts(entlite.PROTO()),",
			wantSQLC:  []schema.QueryType{schema.QueryCreate},
			wantProto: []schema.QueryType{schema.QueryCreate, schema.QueryListAll},
		},
		{
			name:      "contracts chained after other operations",
			contracts: bothContracts,
			queries:   `query.ListBy("email").OrderBy("email").Contracts(entlite.SQLC()),`,
			wantSQLC:  []schema.QueryType{schema.QueryListBy},
			wantProto: nil,
		},
		{
			name:      "contract the entity does not have",
			contracts: "entlite.SQLC(),",
			queries:   "query.Create().Contracts(entlite.PROTO()),",
			wantErr:   "which the entity does not have",
		},
		{
			name:      "read only entity drops the write query",
			contracts: "entlite.SQLC(),\n\t\tentlite.PROTO().ReadOnly(),",
			queries:   "query.Create(),\n\t\tquery.ListAll(),",
			wantSQLC:  []schema.QueryType{schema.QueryCreate, schema.QueryListAll},
			wantProto: []schema.QueryType{schema.QueryListAll},
		},
		{
			name:      "read only entity is a ceiling",
			contracts: "entlite.SQLC(),\n\t\tentlite.PROTO().ReadOnly(),",
			queries:   "query.Create().Contracts(entlite.SQLC(), entlite.PROTO()),",
			wantErr:   "that entity contract is read only",
		},
		{
			name:      "access is rejected on a query",
			contracts: bothContracts,
			queries:   "query.Create().Contracts(entlite.PROTO().ReadOnly()),",
			wantErr:   "a query is already a read or a write",
		},
		{
			name:      "empty contracts",
			contracts: bothContracts,
			queries:   "query.Create().Contracts(),",
			wantErr:   "expects at least one",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity, err := parseQueryContractEntity(t, tt.contracts, tt.queries)

			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected error containing %q, got: %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}

			if got := queryTypes(entity.SQLCQueries()); !equalTypes(got, tt.wantSQLC) {
				t.Fatalf("sqlc queries: expected %v, got %v", tt.wantSQLC, got)
			}
			if got := queryTypes(entity.ProtoQueries()); !equalTypes(got, tt.wantProto) {
				t.Fatalf("proto queries: expected %v, got %v", tt.wantProto, got)
			}
		})
	}
}

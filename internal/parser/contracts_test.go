package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/guntisdev/entlite/internal/schema"
)

const contractEntityTemplate = `package schema

import (
	"github.com/guntisdev/entlite/pkg/entlite"
	"github.com/guntisdev/entlite/pkg/entlite/field"
)

type User struct {
	entlite.Schema
}

%s

func (User) Fields() []entlite.Field {
	return []entlite.Field{
		field.String("email").Unique(),
	}
}
`

func parseContractEntity(t *testing.T, contractsMethod string) error {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "user.go")
	source := strings.Replace(contractEntityTemplate, "%s", contractsMethod, 1)

	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		t.Fatalf("failed to write entity file: %v", err)
	}

	_, err := ParseEntities([]DiscoveredEntity{{Name: "User", Path: path}})
	return err
}

func TestContractsValidation(t *testing.T) {
	tests := []struct {
		name            string
		contractsMethod string
		wantErr         string
	}{
		{
			name: "both contracts",
			contractsMethod: `func (User) Contracts() []entlite.Contract {
	return []entlite.Contract{
		entlite.SQLC(),
		entlite.PROTO(),
	}
}`,
		},
		{
			name: "only sqlc",
			contractsMethod: `func (User) Contracts() []entlite.Contract {
	return []entlite.Contract{
		entlite.SQLC(),
	}
}`,
		},
		{
			name: "only proto",
			contractsMethod: `func (User) Contracts() []entlite.Contract {
	return []entlite.Contract{
		entlite.PROTO(),
	}
}`,
		},
		{
			name:            "missing method",
			contractsMethod: "",
			wantErr:         `is missing Contracts() method`,
		},
		{
			name: "empty list",
			contractsMethod: `func (User) Contracts() []entlite.Contract {
	return []entlite.Contract{}
}`,
			wantErr: `has empty Contracts()`,
		},
		{
			name: "method on another type",
			contractsMethod: `type Post struct {
	entlite.Schema
}

func (Post) Contracts() []entlite.Contract {
	return []entlite.Contract{
		entlite.SQLC(),
	}
}`,
			wantErr: `is missing Contracts() method`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := parseContractEntity(t, tt.contractsMethod)

			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("expected no error, got: %v", err)
				}
				return
			}

			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("expected error containing %q, got: %v", tt.wantErr, err)
			}
		})
	}
}

const accessEntityTemplate = `package schema

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
		entlite.SQLC(),
		entlite.PROTO()%s,
	}
}

func (User) Fields() []entlite.Field {
	return []entlite.Field{
		field.String("email").Unique(),
	}
}

func (User) Queries() []entlite.Query {
	return []entlite.Query{
		query.Create(),
		query.Get(),
		query.ListAll(),
	}
}
`

func parseAccessEntity(t *testing.T, chain string) (schema.Entity, error) {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "user.go")
	source := strings.Replace(accessEntityTemplate, "%s", chain, 1)

	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		t.Fatalf("failed to write entity file: %v", err)
	}

	entities, err := ParseEntities([]DiscoveredEntity{{Name: "User", Path: path}})
	if err != nil {
		return schema.Entity{}, err
	}

	return entities[0], nil
}

func TestContractAccess(t *testing.T) {
	tests := []struct {
		name       string
		chain      string
		wantAccess schema.Access
		wantProto  []schema.QueryType
		wantErr    string
	}{
		{
			name:       "no access is full",
			chain:      "",
			wantAccess: schema.AccessFull,
			wantProto:  []schema.QueryType{schema.QueryCreate, schema.QueryGetBy, schema.QueryListAll},
		},
		{
			name:       "read only drops write queries",
			chain:      ".ReadOnly()",
			wantAccess: schema.AccessRead,
			wantProto:  []schema.QueryType{schema.QueryGetBy, schema.QueryListAll},
		},
		{
			name:       "write only drops read queries",
			chain:      ".WriteOnly()",
			wantAccess: schema.AccessWrite,
			wantProto:  []schema.QueryType{schema.QueryCreate},
		},
		{
			name:    "unknown operation",
			chain:   ".Readonly()",
			wantErr: "unsupported contract operation",
		},
		{
			name:    "operation with arguments",
			chain:   `.ReadOnly("x")`,
			wantErr: "does not accept arguments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity, err := parseAccessEntity(t, tt.chain)

			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected error containing %q, got: %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}

			contract, ok := entity.GetContract(schema.ContractPROTO)
			if !ok {
				t.Fatal("expected proto contract")
			}
			if contract.Access != tt.wantAccess {
				t.Fatalf("expected access %q, got %q", tt.wantAccess, contract.Access)
			}

			var got []schema.QueryType
			for _, q := range entity.ProtoQueries() {
				got = append(got, q.Type)
			}
			if len(got) != len(tt.wantProto) {
				t.Fatalf("expected proto queries %v, got %v", tt.wantProto, got)
			}
			for i := range got {
				if got[i] != tt.wantProto[i] {
					t.Fatalf("expected proto queries %v, got %v", tt.wantProto, got)
				}
			}

			// sqlc contract has no access, so it keeps every query
			if len(entity.SQLCQueries()) != len(entity.Queries) {
				t.Fatalf("expected all %d queries for sqlc, got %d", len(entity.Queries), len(entity.SQLCQueries()))
			}
		})
	}
}

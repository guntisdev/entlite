package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const gatingSchemaTemplate = `package schema

import (
	"github.com/guntisdev/entlite/pkg/entlite"
	"github.com/guntisdev/entlite/pkg/entlite/field"
	"github.com/guntisdev/entlite/pkg/entlite/query"
)

type Note struct {
	entlite.Schema
}

func (Note) Contracts() []entlite.Contract {
	return []entlite.Contract{
		%s
	}
}

func (Note) Fields() []entlite.Field {
	return []entlite.Field{
		field.String("title"),
	}
}

func (Note) Queries() []entlite.Query {
	return []entlite.Query{
		query.Create(),
	}
}
`

const gatingSqlcYaml = `version: "2"
sql:
  - schema: "contract/sqlc/schema.sql"
    queries: "contract/sqlc/queries.sql"
    engine: "postgresql"
    gen:
      go:
        package: "internal"
        out: "gen/db/internal"
`

// runGenWithContracts generates one entity declaring only the given contracts
func runGenWithContracts(t *testing.T, contracts string) string {
	t.Helper()

	tmpDir := t.TempDir()
	entDir := filepath.Join(tmpDir, "ent")
	schemaDir := filepath.Join(entDir, "schema")

	if err := os.MkdirAll(schemaDir, 0755); err != nil {
		t.Fatalf("failed to create schema directory: %v", err)
	}

	source := strings.Replace(gatingSchemaTemplate, "%s", contracts, 1)
	if err := os.WriteFile(filepath.Join(schemaDir, "note.go"), []byte(source), 0644); err != nil {
		t.Fatalf("failed to write entity file: %v", err)
	}

	if err := os.WriteFile(filepath.Join(entDir, "sqlc.yaml"), []byte(gatingSqlcYaml), 0644); err != nil {
		t.Fatalf("failed to write sqlc.yaml: %v", err)
	}

	genCommand([]string{schemaDir})

	return entDir
}

func fileExists(t *testing.T, path string) bool {
	t.Helper()

	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}

	t.Fatalf("failed to stat %s: %v", path, err)
	return false
}

func TestGenCommandContractGating(t *testing.T) {
	tests := []struct {
		name      string
		contracts string
		wantProto bool
		wantSqlc  bool
	}{
		{
			name:      "both contracts",
			contracts: "entlite.SQLC(),\n\t\tentlite.PROTO(),",
			wantProto: true,
			wantSqlc:  true,
		},
		{
			name:      "sqlc only",
			contracts: "entlite.SQLC(),",
			wantProto: false,
			wantSqlc:  true,
		},
		{
			name:      "proto only",
			contracts: "entlite.PROTO(),",
			wantProto: true,
			wantSqlc:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entDir := runGenWithContracts(t, tt.contracts)

			protoPath := filepath.Join(entDir, "contract", "proto", "schema.proto")
			schemaPath := filepath.Join(entDir, "contract", "sqlc", "schema.sql")
			queriesPath := filepath.Join(entDir, "contract", "sqlc", "queries.sql")

			if got := fileExists(t, protoPath); got != tt.wantProto {
				t.Errorf("schema.proto exists = %v, want %v", got, tt.wantProto)
			}
			if got := fileExists(t, schemaPath); got != tt.wantSqlc {
				t.Errorf("schema.sql exists = %v, want %v", got, tt.wantSqlc)
			}
			if got := fileExists(t, queriesPath); got != tt.wantSqlc {
				t.Errorf("queries.sql exists = %v, want %v", got, tt.wantSqlc)
			}
		})
	}
}

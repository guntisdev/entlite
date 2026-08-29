package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/guntisdev/entlite/internal/schema"
)

const fieldContractTemplate = `package schema

import (
	"github.com/guntisdev/entlite/pkg/entlite"
	"github.com/guntisdev/entlite/pkg/entlite/field"
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
		%s
	}
}
`

func parseFieldContractEntity(t *testing.T, contracts, fields string) (schema.Entity, error) {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "user.go")
	source := strings.Replace(fieldContractTemplate, "%s", contracts, 1)
	source = strings.Replace(source, "%s", fields, 1)

	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		t.Fatalf("failed to write entity file: %v", err)
	}

	entities, err := ParseEntities([]DiscoveredEntity{{Name: "User", Path: path}})
	if err != nil {
		return schema.Entity{}, err
	}

	return entities[0], nil
}

func TestFieldContracts(t *testing.T) {
	tests := []struct {
		name         string
		contracts    string
		field        string
		wantDbRead   bool
		wantDbWrite  bool
		wantApiRead  bool
		wantApiWrite bool
		wantVirtual  bool
		wantErr      string
	}{
		{
			name:         "no contracts inherits the entity",
			contracts:    bothContracts,
			field:        `field.String("email"),`,
			wantDbRead:   true,
			wantDbWrite:  true,
			wantApiRead:  true,
			wantApiWrite: true,
		},
		{
			name:        "old ReadOnly, server writes and clients read",
			contracts:   bothContracts,
			field:       `field.String("email").Contracts(entlite.SQLC(), entlite.PROTO().ReadOnly()),`,
			wantDbRead:  true,
			wantDbWrite: true,
			wantApiRead: true,
		},
		{
			name:         "old WriteOnly, clients send but never read back",
			contracts:    bothContracts,
			field:        `field.String("password").Contracts(entlite.SQLC(), entlite.PROTO().WriteOnly()),`,
			wantDbRead:   true,
			wantDbWrite:  true,
			wantApiWrite: true,
		},
		{
			name:        "old Internal, no proto at all",
			contracts:   bothContracts,
			field:       `field.String("note").Contracts(entlite.SQLC()),`,
			wantDbRead:  true,
			wantDbWrite: true,
		},
		{
			name:         "old Virtual, no column",
			contracts:    bothContracts,
			field:        `field.Float("latest").Contracts(entlite.PROTO()),`,
			wantApiRead:  true,
			wantApiWrite: true,
			wantVirtual:  true,
		},
		{
			name:        "sqlc read only, a column the generated SQL never writes",
			contracts:   bothContracts,
			field:       `field.Time("touched_at").Contracts(entlite.SQLC().ReadOnly(), entlite.PROTO().ReadOnly()),`,
			wantDbRead:  true,
			wantApiRead: true,
		},
		{
			name:      "contract the entity does not have",
			contracts: "entlite.SQLC(),",
			field:     `field.String("email").Contracts(entlite.PROTO()),`,
			wantErr:   "which the entity does not have",
		},
		{
			name:      "empty contracts",
			contracts: bothContracts,
			field:     `field.String("email").Contracts(),`,
			wantErr:   "expects at least one",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity, err := parseFieldContractEntity(t, tt.contracts, tt.field)

			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected error containing %q, got: %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}

			// the entity always gets an id field first, the declared field is last
			field := entity.Fields[len(entity.Fields)-1]

			if got := field.CanDbRead(); got != tt.wantDbRead {
				t.Errorf("CanDbRead: expected %v, got %v", tt.wantDbRead, got)
			}
			if got := field.CanDbWrite(); got != tt.wantDbWrite {
				t.Errorf("CanDbWrite: expected %v, got %v", tt.wantDbWrite, got)
			}
			if got := field.CanApiRead(); got != tt.wantApiRead {
				t.Errorf("CanApiRead: expected %v, got %v", tt.wantApiRead, got)
			}
			if got := field.CanApiWrite(); got != tt.wantApiWrite {
				t.Errorf("CanApiWrite: expected %v, got %v", tt.wantApiWrite, got)
			}
			if got := entity.IsFieldVirtual(field); got != tt.wantVirtual {
				t.Errorf("IsFieldVirtual: expected %v, got %v", tt.wantVirtual, got)
			}
		})
	}
}

func TestProtoOnlyEntityHasNoVirtualFields(t *testing.T) {
	entity, err := parseFieldContractEntity(t, "entlite.PROTO(),", `field.String("player"),`)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	for _, field := range entity.Fields {
		if entity.IsFieldVirtual(field) {
			t.Errorf("field %q should not be virtual on a proto only entity", field.Name)
		}
	}
}

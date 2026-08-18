package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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

package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const indexEntityTemplate = `package schema

import (
	"github.com/guntisdev/entlite/pkg/entlite"
	"github.com/guntisdev/entlite/pkg/entlite/field"
	"github.com/guntisdev/entlite/pkg/entlite/index"
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
		field.String("email").Unique(),
		field.String("name"),
		field.Bool("is_active").Default(true),
		field.String("captcha").Contracts(entlite.PROTO()),
	}
}

func (User) Indexes() []entlite.Index {
	return []entlite.Index{
		%s
	}
}
`

func parseIndexEntity(t *testing.T, indexes string) error {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "user.go")
	source := strings.Replace(indexEntityTemplate, "%s", indexes, 1)

	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		t.Fatalf("failed to write entity file: %v", err)
	}

	_, err := ParseEntities([]DiscoveredEntity{{Name: "User", Path: path}})
	return err
}

func TestIndexFieldValidation(t *testing.T) {
	tests := []struct {
		name    string
		indexes string
		wantErr string
	}{
		{
			name:    "existing fields",
			indexes: `index.Asc("email", "is_active"),`,
		},
		{
			name:    "unknown field",
			indexes: `index.Asc("tenant_id", "email").Unique(),`,
			wantErr: `index references nonexisting field "tenant_id"`,
		},
		{
			name:    "unknown field in Desc modifier",
			indexes: `index.Asc("is_active").Desc("created_at"),`,
			wantErr: `index references nonexisting field "created_at"`,
		},
		{
			name:    "descending first column",
			indexes: `index.Desc("email").Asc("is_active"),`,
		},
		{
			name:    "same field twice",
			indexes: `index.Asc("email").Desc("email"),`,
			wantErr: `index lists field "email" twice`,
		},
		{
			name:    "no fields",
			indexes: `index.Asc(),`,
			wantErr: `index.Asc requires at least one field`,
		},
		{
			name:    "unknown field in Primary",
			indexes: `index.Primary("country", "env"),`,
			wantErr: `index references nonexisting field "country"`,
		},
		{
			name:    "virtual field",
			indexes: `index.Asc("captcha"),`,
			wantErr: `index references virtual field "captcha"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := parseIndexEntity(t, tt.indexes)

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

package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const namesEntityTemplate = `package schema

import (
	"github.com/guntisdev/entlite/pkg/entlite"
	"github.com/guntisdev/entlite/pkg/entlite/field"
	"github.com/guntisdev/entlite/pkg/entlite/query"
)

type ENTITY struct {
	entlite.Schema
}

func (ENTITY) Contracts() []entlite.Contract {
	return []entlite.Contract{entlite.SQLC(), entlite.PROTO()}
}

func (ENTITY) Fields() []entlite.Field {
	return []entlite.Field{
		FIELDS
	}
}

func (ENTITY) Queries() []entlite.Query {
	return []entlite.Query{query.Create()}
}
`

func parseNamedEntity(t *testing.T, entityName, fields string) error {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "entity.go")
	source := strings.ReplaceAll(namesEntityTemplate, "ENTITY", entityName)
	source = strings.Replace(source, "FIELDS", fields, 1)

	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		t.Fatalf("failed to write entity file: %v", err)
	}

	_, err := ParseEntities([]DiscoveredEntity{{Name: entityName, Path: path}})
	return err
}

const okFields = `field.String("env"),`

func TestEntityNameIsChecked(t *testing.T) {
	if err := parseNamedEntity(t, "PopularProduct", okFields); err != nil {
		t.Fatalf("expected PascalCase to be accepted, got: %v", err)
	}

	rejected := map[string]string{
		"popularProduct":  "must start with a capital",
		"Popular_Product": "must be PascalCase",
		"HTTPServer":     "two capitals in a row",
	}
	for name, want := range rejected {
		err := parseNamedEntity(t, name, okFields)
		if err == nil {
			t.Errorf("expected entity %q to be rejected", name)
			continue
		}
		if !strings.Contains(err.Error(), want) {
			t.Errorf("entity %q: expected an error about %q, got: %v", name, want, err)
		}
	}
}

func TestFieldNameIsChecked(t *testing.T) {
	if err := parseNamedEntity(t, "Setting", `field.String("is_active"),`); err != nil {
		t.Fatalf("expected snake_case to be accepted, got: %v", err)
	}

	rejected := map[string]string{
		`field.String("isActive"),`:   "must be lower snake_case",
		`field.String("IsActive"),`:   "must be lower snake_case",
		`field.String("is__active"),`: "empty word",
		`field.String("ID"),`:         "must be lower snake_case",
	}
	for fields, want := range rejected {
		err := parseNamedEntity(t, "Setting", fields)
		if err == nil {
			t.Errorf("expected %s to be rejected", fields)
			continue
		}
		if !strings.Contains(err.Error(), want) {
			t.Errorf("%s: expected an error about %q, got: %v", fields, want, err)
		}
	}
}

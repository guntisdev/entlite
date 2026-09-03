package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/guntisdev/entlite/internal/schema"
)

const primaryEntityTemplate = `package schema

import (
	"github.com/guntisdev/entlite/pkg/entlite"
	"github.com/guntisdev/entlite/pkg/entlite/field"
	"github.com/guntisdev/entlite/pkg/entlite/index"
	"github.com/guntisdev/entlite/pkg/entlite/query"
)

type Setting struct {
	entlite.Schema
}

func (Setting) Contracts() []entlite.Contract {
	return []entlite.Contract{
		entlite.SQLC(),
		entlite.PROTO(),
	}
}

func (Setting) Fields() []entlite.Field {
	return []entlite.Field{
		FIELDS
	}
}

func (Setting) Queries() []entlite.Query {
	return []entlite.Query{
		query.DefaultCRUD(),
	}
}

func (Setting) Indexes() []entlite.Index {
	return []entlite.Index{
		INDEXES
	}
}
`

const defaultPrimaryFields = `field.String("country"),
		field.String("env"),
		field.String("value"),`

func parsePrimaryEntity(t *testing.T, fields, indexes string) (schema.Entity, error) {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "setting.go")
	source := strings.Replace(primaryEntityTemplate, "FIELDS", fields, 1)
	source = strings.Replace(source, "INDEXES", indexes, 1)

	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		t.Fatalf("failed to write entity file: %v", err)
	}

	entities, err := ParseEntities([]DiscoveredEntity{{Name: "Setting", Path: path}})
	if err != nil {
		return schema.Entity{}, err
	}

	return entities[0], nil
}

func fieldNames(entity schema.Entity) []string {
	names := make([]string, 0, len(entity.Fields))
	for _, field := range entity.Fields {
		names = append(names, field.Name)
	}

	return names
}

func queryFields(t *testing.T, entity schema.Entity, queryType schema.QueryType) []string {
	t.Helper()

	for _, query := range entity.Queries {
		if query.Type == queryType {
			return query.Fields
		}
	}

	t.Fatalf("entity has no %q query", queryType)
	return nil
}

func TestPrimaryIndexDropsIdField(t *testing.T) {
	entity, err := parsePrimaryEntity(t, defaultPrimaryFields, `index.Primary("country", "env"),`)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if entity.HasIdField() {
		t.Fatalf("expected no id field, got fields %v", fieldNames(entity))
	}

	want := []string{"country", "env"}
	for _, queryType := range []schema.QueryType{schema.QueryGetBy, schema.QueryUpdate, schema.QueryDelete} {
		got := queryFields(t, entity, queryType)
		if strings.Join(got, ",") != strings.Join(want, ",") {
			t.Errorf("query %q keyed by %v, expected %v", queryType, got, want)
		}
	}

	// the proto field numbers start at the first declared field
	if entity.Fields[0].ProtoField != 1 {
		t.Errorf("expected the first field to be proto field 1, got %d", entity.Fields[0].ProtoField)
	}
}

func TestWithoutPrimaryIndexIdFieldIsAdded(t *testing.T) {
	entity, err := parsePrimaryEntity(t, defaultPrimaryFields, `index.Fields("value"),`)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if !entity.HasIdField() {
		t.Fatalf("expected an id field, got fields %v", fieldNames(entity))
	}
	if !entity.GetIdField().Primary {
		t.Error("expected the id field to be the primary key")
	}

	for _, queryType := range []schema.QueryType{schema.QueryGetBy, schema.QueryUpdate, schema.QueryDelete} {
		got := queryFields(t, entity, queryType)
		if strings.Join(got, ",") != "ID" {
			t.Errorf("query %q keyed by %v, expected [ID]", queryType, got)
		}
	}
}

// an index.Primary("id") still needs the id column, only its auto increment goes away
func TestPrimaryIndexOnIdKeepsIdField(t *testing.T) {
	entity, err := parsePrimaryEntity(t, defaultPrimaryFields, `index.Primary("id", "env"),`)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if !entity.HasIdField() {
		t.Fatalf("expected an id field, got fields %v", fieldNames(entity))
	}
	if entity.GetIdField().Primary {
		t.Error("expected the id field to give up the primary key to the index")
	}
	if got := queryFields(t, entity, schema.QueryGetBy); strings.Join(got, ",") != "ID,env" {
		t.Errorf("get keyed by %v, expected [ID env]", got)
	}
}

// a declared id field is never dropped, it stays a plain column
func TestPrimaryIndexKeepsDeclaredIdField(t *testing.T) {
	fields := `field.String("id"),
		field.String("country"),
		field.String("env"),`

	entity, err := parsePrimaryEntity(t, fields, `index.Primary("country", "env"),`)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if !entity.HasIdField() {
		t.Fatalf("expected the declared id field, got fields %v", fieldNames(entity))
	}
	if entity.GetIdField().Primary {
		t.Error("expected the id field to give up the primary key to the index")
	}
	if got := queryFields(t, entity, schema.QueryDelete); strings.Join(got, ",") != "country,env" {
		t.Errorf("delete keyed by %v, expected [country env]", got)
	}
}

// the index error comes first, the queries keyed by it have nothing to resolve to
func TestPrimaryIndexRejectsUnknownField(t *testing.T) {
	_, err := parsePrimaryEntity(t, defaultPrimaryFields, `index.Primary("tenant", "env"),`)
	if err == nil {
		t.Fatal("expected an error for an unknown primary key column")
	}
	if !strings.Contains(err.Error(), `index references nonexisting field "tenant"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPrimaryIndexRejectsOptionalField(t *testing.T) {
	fields := `field.String("country"),
		field.String("env").Optional(),`

	_, err := parsePrimaryEntity(t, fields, `index.Primary("country", "env"),`)
	if err == nil {
		t.Fatal("expected an error for an optional primary key column")
	}
	if !strings.Contains(err.Error(), `index.Primary references optional field "env"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

package proto

import (
	"strings"
	"testing"

	"github.com/guntisdev/entlite/internal/schema"
)

// the id field gives up the key to index.Primary, so the caller supplies its value
func suppliedIdEntity() schema.Entity {
	contracts := []schema.Contract{{Type: schema.ContractSQLC}, {Type: schema.ContractPROTO}}

	return schema.Entity{
		Name:      "Casino",
		Contracts: contracts,
		Fields: []schema.Field{
			{Name: "id", Type: schema.FieldTypeString, ProtoField: 1, Unique: true, Immutable: true, Contracts: contracts},
			{Name: "env", Type: schema.FieldTypeString, ProtoField: 2, Immutable: true, Contracts: contracts},
			{Name: "init_count", Type: schema.FieldTypeInt, ProtoField: 3, Contracts: contracts},
		},
		Indexes: []schema.Index{{
			Type:    schema.IndexPrimary,
			Columns: []schema.IndexColumn{{Name: "id"}, {Name: "env"}},
		}},
		Queries: []schema.Query{
			{Type: schema.QueryCreate, Name: "CreateCasino", Contracts: contracts},
			{Type: schema.QueryCreateBulk, Name: "CreateBulkCasino", Contracts: contracts},
		},
	}
}

func TestCallerSuppliedIdInCreateRequest(t *testing.T) {
	content := generateSchemaProto([]schema.Entity{suppliedIdEntity()}, "example/gen/pb")

	create := `message CreateCasinoRequest {
  string id = 1 [(buf.validate.field).required = true];`
	if !strings.Contains(content, create) {
		t.Errorf("expected the create request to require the id:\n%s", content)
	}

	bulk := `message CreateBulkCasinoRow {
  string id = 1 [(buf.validate.field).required = true];`
	if !strings.Contains(content, bulk) {
		t.Errorf("expected the bulk row to require the id:\n%s", content)
	}
}

// an id the database assigns never reaches the client on create
func TestGeneratedIdStaysOutOfCreateRequest(t *testing.T) {
	entity := suppliedIdEntity()
	entity.Indexes = nil
	entity.Fields[0].Primary = true

	content := generateSchemaProto([]schema.Entity{entity}, "example/gen/pb")

	create := `message CreateCasinoRequest {
  string env = 2`
	if !strings.Contains(content, create) {
		t.Errorf("expected the create request to start after the id:\n%s", content)
	}
}

package proto

import (
	"strings"
	"testing"

	"github.com/guntisdev/entlite/internal/schema"
)

func countEntity() schema.Entity {
	contracts := []schema.Contract{{Type: schema.ContractSQLC}, {Type: schema.ContractPROTO}}

	return schema.Entity{
		Name:      "Post",
		Contracts: contracts,
		Fields: []schema.Field{
			{Name: "id", Type: schema.FieldTypeInt, Primary: true, ProtoField: 1, Contracts: contracts},
			{Name: "title", Type: schema.FieldTypeString, ProtoField: 2, Contracts: contracts},
		},
		Queries: []schema.Query{
			{Type: schema.QueryListBy, Name: "ListPostPaged", Fields: []string{"title"},
				Count: true, HasLimit: true, HasOffset: true, Contracts: contracts},
			{Type: schema.QueryListBy, Name: "ListPostPlain", Fields: []string{"title"}, Contracts: contracts},
		},
	}
}

func TestCountResponseField(t *testing.T) {
	content := generateSchemaProto([]schema.Entity{countEntity()}, "example/gen/pb")

	counted := `message ListPostPagedResponse {
  repeated Post rows = 1;
  int64 total_size = 2;
}`
	if !strings.Contains(content, counted) {
		t.Errorf("expected the counted response to carry total_size:\n%s", content)
	}

	plain := `message ListPostPlainResponse {
  repeated Post rows = 1;
}`
	if !strings.Contains(content, plain) {
		t.Errorf("expected the plain response to stay without total_size:\n%s", content)
	}
}

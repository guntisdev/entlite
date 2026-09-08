package proto

import (
	"strings"
	"testing"

	"github.com/guntisdev/entlite/internal/schema"
)

func distinctEntity() schema.Entity {
	contracts := []schema.Contract{{Type: schema.ContractSQLC}, {Type: schema.ContractPROTO}}

	return schema.Entity{
		Name:      "Release",
		Contracts: contracts,
		Fields: []schema.Field{
			{Name: "id", Type: schema.FieldTypeInt, Primary: true, ProtoField: 1, Contracts: contracts},
			{Name: "env", Type: schema.FieldTypeString, ProtoField: 2, Contracts: contracts},
			{Name: "name", Type: schema.FieldTypeString, ProtoField: 3, Contracts: contracts},
			{Name: "timestamp", Type: schema.FieldTypeTime, ProtoField: 4, Contracts: contracts},
			{Name: "build", Type: schema.FieldTypeString, Optional: true, ProtoField: 5, Contracts: contracts},
		},
		Queries: []schema.Query{
			{Type: schema.QueryListBy, Name: "ListReleaseNames",
				Filters:  []schema.QueryFilter{{Field: "env", Type: schema.QueryFilterEq}},
				Distinct: []string{"name"}, Contracts: contracts},
			{Type: schema.QueryListAll, Name: "ListReleaseLatest",
				Distinct: []string{"name", "timestamp"}, Contracts: contracts},
			{Type: schema.QueryListAll, Name: "ListReleaseBuilds",
				Distinct: []string{"build"}, Contracts: contracts},
			{Type: schema.QueryListAll, Name: "ListAllRelease", Contracts: contracts},
		},
	}
}

func TestDistinctResponseMessages(t *testing.T) {
	content := generateSchemaProto([]schema.Entity{distinctEntity()}, "example/gen/pb")

	// one column returns the values themselves
	single := `message ListReleaseNamesResponse {
  repeated string name = 1;
}`
	if !strings.Contains(content, single) {
		t.Errorf("expected a repeated value response:\n%s", content)
	}

	// several columns need a row message of their own
	several := `message ListReleaseLatestRow {
  string name = 1;
  google.protobuf.Timestamp timestamp = 2;
}

message ListReleaseLatestResponse {
  repeated ListReleaseLatestRow rows = 1;
}`
	if !strings.Contains(content, several) {
		t.Errorf("expected a row message response:\n%s", content)
	}

	// a plain list still returns entities
	plain := `message ListAllReleaseResponse {
  repeated Release rows = 1;
}`
	if !strings.Contains(content, plain) {
		t.Errorf("expected the plain list to keep returning entities:\n%s", content)
	}

	// a distinct query returns columns, so it never returns the entity
	if strings.Contains(content, "message ListReleaseNamesResponse {\n  repeated Release") {
		t.Errorf("expected no entity in a distinct response:\n%s", content)
	}
}

// a repeated field cannot be optional in proto3, but the value type still has to match
func TestDistinctOptionalColumnResponse(t *testing.T) {
	content := generateSchemaProto([]schema.Entity{distinctEntity()}, "example/gen/pb")

	want := `message ListReleaseBuildsResponse {
  repeated string build = 1;
}`
	if !strings.Contains(content, want) {
		t.Errorf("expected an optional column to stay a plain repeated value:\n%s", content)
	}
}

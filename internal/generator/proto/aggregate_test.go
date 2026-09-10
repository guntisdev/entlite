package proto

import (
	"strings"
	"testing"

	"github.com/guntisdev/entlite/internal/schema"
)

func aggregateEntity() schema.Entity {
	contracts := []schema.Contract{{Type: schema.ContractSQLC}, {Type: schema.ContractPROTO}}

	return schema.Entity{
		Name:      "Build",
		Contracts: contracts,
		Fields: []schema.Field{
			{Name: "id", Type: schema.FieldTypeInt, Primary: true, ProtoField: 1, Contracts: contracts},
			{Name: "branch", Type: schema.FieldTypeString, ProtoField: 2, Contracts: contracts},
			{Name: "env", Type: schema.FieldTypeString, ProtoField: 3, Contracts: contracts},
			{Name: "failed_tests", Type: schema.FieldTypeInt, ProtoField: 4, Contracts: contracts},
			{Name: "duration_ms", Type: schema.FieldTypeInt64, ProtoField: 5, Contracts: contracts},
			{Name: "coverage", Type: schema.FieldTypeFloat, ProtoField: 6, Contracts: contracts},
		},
		Queries: []schema.Query{
			{Type: schema.QueryListBy, Name: "SumBranchDurations",
				Filters:    []schema.QueryFilter{{Field: "env", Type: schema.QueryFilterEq}},
				GroupBy:    []string{"branch"},
				Aggregates: []schema.Aggregate{{Func: schema.AggregateSum, Field: "duration_ms"}}, Contracts: contracts},
			{Type: schema.QueryListAll, Name: "BranchStats",
				GroupBy: []string{"branch"},
				Aggregates: []schema.Aggregate{
					{Func: schema.AggregateAvg, Field: "coverage"},
					{Func: schema.AggregateSum, Field: "failed_tests"},
					{Func: schema.AggregateMin, Field: "env"},
				}, Contracts: contracts},
			{Type: schema.QueryListAll, Name: "TotalDuration",
				Aggregates: []schema.Aggregate{{Func: schema.AggregateSum, Field: "duration_ms"}}, Contracts: contracts},
			{Type: schema.QueryListAll, Name: "BuildTotals",
				Aggregates: []schema.Aggregate{
					{Func: schema.AggregateAvg, Field: "coverage"},
					{Func: schema.AggregateMax, Field: "failed_tests"},
				}, Contracts: contracts},
		},
	}
}

func TestAggregateResponseMessages(t *testing.T) {
	content := generateSchemaProto([]schema.Entity{aggregateEntity()}, "example/gen/pb")

	// a group returns one row per group, keyed by the grouped columns
	grouped := `message SumBranchDurationsRow {
  string branch = 1;
  int64 sum_duration_ms = 2;
}

message SumBranchDurationsResponse {
  repeated SumBranchDurationsRow rows = 1;
}`
	if !strings.Contains(content, grouped) {
		t.Errorf("expected a grouped row message:\n%s", content)
	}

	// an average is a double and a sum of an int32 widens
	stats := `message BranchStatsRow {
  string branch = 1;
  double avg_coverage = 2;
  int64 sum_failed_tests = 3;
  string min_env = 4;
}`
	if !strings.Contains(content, stats) {
		t.Errorf("expected the aggregate types on the row message:\n%s", content)
	}

	// without a group the aggregates sit on the response itself
	single := `message TotalDurationResponse {
  int64 sum_duration_ms = 1;
}`
	if !strings.Contains(content, single) {
		t.Errorf("expected a single value response:\n%s", content)
	}

	several := `message BuildTotalsResponse {
  double avg_coverage = 1;
  int64 max_failed_tests = 2;
}`
	if !strings.Contains(content, several) {
		t.Errorf("expected one response field per aggregate:\n%s", content)
	}

	// an aggregate replaces the rows, so the entity never comes back
	if strings.Contains(content, "message TotalDurationResponse {\n  repeated Build") {
		t.Errorf("expected no entity in an aggregate response:\n%s", content)
	}
}

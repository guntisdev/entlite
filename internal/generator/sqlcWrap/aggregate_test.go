package sqlcwrap

import (
	"strings"
	"testing"

	"github.com/guntisdev/entlite/internal/schema"
)

func aggregateTestEntity() schema.Entity {
	contracts := []schema.Contract{{Type: schema.ContractSQLC}, {Type: schema.ContractPROTO}}

	return schema.Entity{
		Name:      "Build",
		Contracts: contracts,
		Fields: []schema.Field{
			{Name: "id", Type: schema.FieldTypeInt64, Primary: true, Contracts: contracts},
			{Name: "branch", Type: schema.FieldTypeString, Contracts: contracts},
			{Name: "failed_tests", Type: schema.FieldTypeInt, Contracts: contracts},
			{Name: "duration_ms", Type: schema.FieldTypeInt64, Contracts: contracts},
			{Name: "coverage", Type: schema.FieldTypeFloat, Contracts: contracts},
		},
	}
}

func TestGroupedAggregate(t *testing.T) {
	entity := aggregateTestEntity()
	query := schema.Query{Type: schema.QueryListBy, Name: "SumBranchDurations",
		Fields:     []string{"branch"},
		GroupBy:    []string{"branch"},
		Aggregates: []schema.Aggregate{{Func: schema.AggregateSum, Field: "duration_ms"}}}
	funcDecl := parseListMethod(t, `func (q *Queries) SumBranchDurations(ctx context.Context, branch string) ([]SumBranchDurationsRow, error) {
	return nil, nil
}`)

	ctx := countContext(query, entity)
	got := ctx.generateListQuery(funcDecl, entity) + generateAggregateRowStruct(entity, query)

	want := []string{
		"func (q *Queries) SumBranchDurations(ctx context.Context, branch string) ([]SumBranchDurationsRow, error) {",
		"dbResult, err := (*internal.Queries)(q).SumBranchDurations(ctx, branch)",
		"result := make([]SumBranchDurationsRow, len(dbResult))",
		"Branch: dbResult[i].Branch,",
		"SumDurationMs: dbResult[i].SumDurationMs,",
		"type SumBranchDurationsRow struct {",
		"Branch string",
		"SumDurationMs int64",
	}
	for _, part := range want {
		if !strings.Contains(got, part) {
			t.Errorf("expected %q in the wrapper:\n%s", part, got)
		}
	}
}

// one aggregate and no group folds the whole table into a single value
func TestWholeTableAggregate(t *testing.T) {
	entity := aggregateTestEntity()
	query := schema.Query{Type: schema.QueryListAll, Name: "AvgCoverage",
		Aggregates: []schema.Aggregate{{Func: schema.AggregateAvg, Field: "coverage"}}}
	funcDecl := parseListMethod(t, `func (q *Queries) AvgCoverage(ctx context.Context) (float64, error) {
	return 0, nil
}`)

	got := countContext(query, entity).generateListQuery(funcDecl, entity)

	want := []string{
		"func (q *Queries) AvgCoverage(ctx context.Context) (float64, error) {",
		"return 0, err",
		"return dbResult, nil",
	}
	for _, part := range want {
		if !strings.Contains(got, part) {
			t.Errorf("expected %q in the wrapper:\n%s", part, got)
		}
	}
}

// several aggregates and no group still fold into one row
func TestWholeTableAggregateRow(t *testing.T) {
	entity := aggregateTestEntity()
	query := schema.Query{Type: schema.QueryListAll, Name: "BuildTotals",
		Aggregates: []schema.Aggregate{
			{Func: schema.AggregateSum, Field: "failed_tests"},
			{Func: schema.AggregateMin, Field: "branch"},
		}}
	funcDecl := parseListMethod(t, `func (q *Queries) BuildTotals(ctx context.Context) (BuildTotalsRow, error) {
	return BuildTotalsRow{}, nil
}`)

	ctx := countContext(query, entity)
	got := ctx.generateListQuery(funcDecl, entity) + generateAggregateRowStruct(entity, query)

	want := []string{
		"func (q *Queries) BuildTotals(ctx context.Context) (BuildTotalsRow, error) {",
		"return BuildTotalsRow{}, err",
		"return BuildTotalsRow{",
		"SumFailedTests: dbResult.SumFailedTests,",
		"MinBranch: dbResult.MinBranch,",
		// an int32 column sums into an int64, so the row does not restate the field type
		"SumFailedTests int64",
		"MinBranch string",
	}
	for _, part := range want {
		if !strings.Contains(got, part) {
			t.Errorf("expected %q in the wrapper:\n%s", part, got)
		}
	}
}

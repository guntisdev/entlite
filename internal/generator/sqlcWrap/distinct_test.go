package sqlcwrap

import (
	"strings"
	"testing"

	"github.com/guntisdev/entlite/internal/schema"
)

func distinctTestEntity() schema.Entity {
	contracts := []schema.Contract{{Type: schema.ContractSQLC}, {Type: schema.ContractPROTO}}

	return schema.Entity{
		Name:      "Release",
		Contracts: contracts,
		Fields: []schema.Field{
			{Name: "id", Type: schema.FieldTypeInt64, Primary: true, Contracts: contracts},
			{Name: "env", Type: schema.FieldTypeString, Contracts: contracts},
			{Name: "name", Type: schema.FieldTypeString, Contracts: contracts},
			{Name: "timestamp", Type: schema.FieldTypeTime, Contracts: contracts},
			{Name: "build", Type: schema.FieldTypeString, Optional: true, Contracts: contracts},
		},
	}
}

func distinctContext(query schema.Query, entity schema.Entity) *generationContext {
	ctx := countContext(query, entity)
	ctx.sqlDialect = schema.PostgreSQL
	return ctx
}

func TestDistinctSingleColumn(t *testing.T) {
	entity := distinctTestEntity()
	query := schema.Query{Type: schema.QueryListBy, Name: "ListReleaseNames",
		Filters:  []schema.QueryFilter{{Field: "env", Type: schema.QueryFilterEq}},
		Distinct: []string{"name"}}
	funcDecl := parseListMethod(t, `func (q *Queries) ListReleaseNames(ctx context.Context, env string) ([]string, error) {
	return nil, nil
}`)

	got := distinctContext(query, entity).generateListQuery(funcDecl, entity)

	want := []string{
		"func (q *Queries) ListReleaseNames(ctx context.Context, env string) ([]string, error) {",
		"dbResults, err := (*internal.Queries)(q).ListReleaseNames(ctx, env)",
		"result := make([]string, len(dbResults))",
		"result[i] = dbResults[i]",
		"return result, nil",
	}
	for _, part := range want {
		if !strings.Contains(got, part) {
			t.Errorf("expected %q in the wrapper:\n%s", part, got)
		}
	}

	// a distinct query returns column values, so no entity is built
	if strings.Contains(got, "Release{") || strings.Contains(got, "ReleaseFromSQL") {
		t.Errorf("expected no entity in a distinct list:\n%s", got)
	}
}

// an optional column arrives as sql.NullString and has to be converted back
func TestDistinctOptionalColumn(t *testing.T) {
	entity := distinctTestEntity()
	query := schema.Query{Type: schema.QueryListAll, Name: "ListReleaseBuilds", Distinct: []string{"build"}}
	funcDecl := parseListMethod(t, `func (q *Queries) ListReleaseBuilds(ctx context.Context) ([]sql.NullString, error) {
	return nil, nil
}`)

	got := distinctContext(query, entity).generateListQuery(funcDecl, entity)

	want := []string{
		"func (q *Queries) ListReleaseBuilds(ctx context.Context) ([]*string, error) {",
		"result[i] = NullStringToPtr(dbResults[i])",
	}
	for _, part := range want {
		if !strings.Contains(got, part) {
			t.Errorf("expected %q in the wrapper:\n%s", part, got)
		}
	}
}

func TestDistinctSeveralColumns(t *testing.T) {
	entity := distinctTestEntity()
	query := schema.Query{Type: schema.QueryListAll, Name: "ListReleaseLatest", Distinct: []string{"name", "timestamp"}}
	funcDecl := parseListMethod(t, `func (q *Queries) ListReleaseLatest(ctx context.Context) ([]ListReleaseLatestRow, error) {
	return nil, nil
}`)

	ctx := distinctContext(query, entity)
	got := ctx.generateListQuery(funcDecl, entity)

	want := []string{
		"func (q *Queries) ListReleaseLatest(ctx context.Context) ([]ListReleaseLatestRow, error) {",
		"result := make([]ListReleaseLatestRow, len(dbResults))",
		"result[i] = ListReleaseLatestRow{",
		"Name: dbResults[i].Name,",
		"Timestamp: dbResults[i].Timestamp,",
	}
	for _, part := range want {
		if !strings.Contains(got, part) {
			t.Errorf("expected %q in the wrapper:\n%s", part, got)
		}
	}

	// the row struct is restated with the wrapper's types, not aliased to sqlc's
	target, ok := ctx.distinctRowQuery("ListReleaseLatestRow")
	if !ok {
		t.Fatal("expected ListReleaseLatestRow to belong to the distinct query")
	}

	structDef := generateDistinctRowStruct(target.entity, target.query)
	wantStruct := `type ListReleaseLatestRow struct {
	Name string
	Timestamp time.Time
}`
	if !strings.Contains(structDef, wantStruct) {
		t.Errorf("expected the restated row struct:\n%s", structDef)
	}
}

// one column needs no row struct, so sqlc declares none to restate
func TestDistinctSingleColumnHasNoRowStruct(t *testing.T) {
	entity := distinctTestEntity()
	query := schema.Query{Type: schema.QueryListAll, Name: "ListReleaseNames", Distinct: []string{"name"}}

	if _, ok := distinctContext(query, entity).distinctRowQuery("ListReleaseNamesRow"); ok {
		t.Error("expected no row struct for a single column distinct")
	}
}

// a plain list query keeps sqlc's row struct alias
func TestPlainListRowStructIsNotClaimed(t *testing.T) {
	entity := distinctTestEntity()
	query := schema.Query{Type: schema.QueryListBy, Name: "ListReleaseByEnv", Fields: []string{"env"}, Count: true}

	if _, ok := distinctContext(query, entity).distinctRowQuery("ListReleaseByEnvRow"); ok {
		t.Error("expected a counted list row struct to stay untouched")
	}
}

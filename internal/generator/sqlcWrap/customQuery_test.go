package sqlcwrap

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/guntisdev/entlite/internal/schema"
)

func casinoEntity() schema.Entity {
	contracts := []schema.Contract{{Type: schema.ContractSQLC}}

	return schema.Entity{
		Name:      "Casino",
		Contracts: contracts,
		Fields: []schema.Field{
			{Name: "id", Type: schema.FieldTypeInt64, Primary: true, Contracts: contracts},
			{Name: "casino", Type: schema.FieldTypeString, Contracts: contracts},
			{Name: "initialized", Type: schema.FieldTypeInt64, Contracts: contracts},
		},
	}
}

func wrapQueryFile(t *testing.T, entity schema.Entity, queries []schema.Query, source string) string {
	t.Helper()

	node, err := parser.ParseFile(token.NewFileSet(), "queries.sql.go", "package internal\n\n"+source, parser.ParseComments)
	if err != nil {
		t.Fatalf("failed to parse source: %v", err)
	}

	dslQueries := make(map[string]dslQuery)
	for _, query := range queries {
		dslQueries[query.Name] = dslQuery{entity: entity, query: query}
	}

	ctx := &generationContext{
		fileType:            FileTypeQuery,
		inputPackageName:    "internal",
		node:                node,
		sqlDialect:          schema.PostgreSQL,
		entityMap:           map[string]schema.Entity{entity.Name: entity},
		dslQueries:          dslQueries,
		parsedEntities:      []schema.Entity{entity},
		methods:             make(map[string]*ast.FuncDecl),
		createParamsStructs: make(map[string]*ast.StructType),
		updateParamsStructs: make(map[string]*ast.StructType),
		filterParamsStructs: make(map[string]*ast.StructType),
	}
	ctx.collectDeclarations()

	return ctx.generateQueryFileDeclarations()
}

// a hand-written query can be named like a dsl query
func TestCustomListQueryWithOwnRowTypeIsForwarded(t *testing.T) {
	got := wrapQueryFile(t, casinoEntity(), nil, `type ListCasinoByUsageParams struct {
	Limit  int64
	Offset int64
}

type ListCasinoByUsageRow struct {
	Casino string
	Total  int64
}

func (q *Queries) ListCasinoByUsage(ctx context.Context, arg ListCasinoByUsageParams) ([]ListCasinoByUsageRow, error) {
	return nil, nil
}`)

	want := []string{
		"type ListCasinoByUsageRow = internal.ListCasinoByUsageRow",
		"type ListCasinoByUsageParams = internal.ListCasinoByUsageParams",
		"func (q *Queries) ListCasinoByUsage(ctx context.Context, arg internal.ListCasinoByUsageParams) ([]internal.ListCasinoByUsageRow, error) {",
		"return (*internal.Queries)(q).ListCasinoByUsage(ctx, arg)",
	}
	for _, part := range want {
		if !strings.Contains(got, part) {
			t.Errorf("expected %q in the wrapper:\n%s", part, got)
		}
	}

	if strings.Contains(got, "CasinoFromSQL") {
		t.Errorf("expected no model converter for an aggregate row:\n%s", got)
	}
}

// the same name, but the select is the plain table, so the wrapper does fit
func TestCustomListQueryOfEntityRowsIsWrapped(t *testing.T) {
	got := wrapQueryFile(t, casinoEntity(), nil, `func (q *Queries) ListCasinoByUsage(ctx context.Context) ([]Casino, error) {
	return nil, nil
}`)

	want := []string{
		"func (q *Queries) ListCasinoByUsage(ctx context.Context) ([]*Casino, error) {",
		"result[i] = CasinoFromSQL(&dbResults[i])",
	}
	for _, part := range want {
		if !strings.Contains(got, part) {
			t.Errorf("expected %q in the wrapper:\n%s", part, got)
		}
	}
}

func TestCustomGetQueryWithOwnRowTypeIsForwarded(t *testing.T) {
	got := wrapQueryFile(t, casinoEntity(), nil, `type GetCasinoByErrorsRow struct {
	Casino string
	Errors int64
}

func (q *Queries) GetCasinoByErrors(ctx context.Context, casino string) (GetCasinoByErrorsRow, error) {
	return GetCasinoByErrorsRow{}, nil
}`)

	want := "func (q *Queries) GetCasinoByErrors(ctx context.Context, casino string) (internal.GetCasinoByErrorsRow, error) {"
	if !strings.Contains(got, want) {
		t.Errorf("expected %q in the wrapper:\n%s", want, got)
	}

	if strings.Contains(got, "CasinoFromSQL") {
		t.Errorf("expected no model converter for an aggregate row:\n%s", got)
	}
}

// :execrows returns a count, so the delete wrapper, which returns only an error, does not fit
func TestCustomDeleteQueryReturningRowsIsForwarded(t *testing.T) {
	got := wrapQueryFile(t, casinoEntity(), nil, `func (q *Queries) DeleteCasino(ctx context.Context, id int64) (int64, error) {
	return 0, nil
}`)

	want := "func (q *Queries) DeleteCasino(ctx context.Context, id int64) (int64, error) {"
	if !strings.Contains(got, want) {
		t.Errorf("expected %q in the wrapper:\n%s", want, got)
	}
}

// a dsl query keeps its wrapper, its row struct comes from the same schema
func TestDslListQueryIsStillWrapped(t *testing.T) {
	entity := casinoEntity()
	query := schema.Query{Type: schema.QueryListBy, Name: "ListCasinoByCasino", Fields: []string{"casino"}}
	got := wrapQueryFile(t, entity, []schema.Query{query}, `func (q *Queries) ListCasinoByCasino(ctx context.Context, casino string) ([]Casino, error) {
	return nil, nil
}`)

	want := []string{
		"func (q *Queries) ListCasinoByCasino(ctx context.Context, casino string) ([]*Casino, error) {",
		"result[i] = CasinoFromSQL(&dbResults[i])",
	}
	for _, part := range want {
		if !strings.Contains(got, part) {
			t.Errorf("expected %q in the wrapper:\n%s", part, got)
		}
	}
}

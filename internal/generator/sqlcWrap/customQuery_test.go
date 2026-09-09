package sqlcwrap

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/guntisdev/entlite/internal/schema"
)

func productEntity() schema.Entity {
	contracts := []schema.Contract{{Type: schema.ContractSQLC}}

	return schema.Entity{
		Name:      "Product",
		Contracts: contracts,
		Fields: []schema.Field{
			{Name: "id", Type: schema.FieldTypeInt64, Primary: true, Contracts: contracts},
			{Name: "name", Type: schema.FieldTypeString, Contracts: contracts},
			{Name: "sold", Type: schema.FieldTypeInt64, Contracts: contracts},
		},
	}
}

// an entity whose id the caller supplies, so entlite's own insert returns nothing
func countryEntity() schema.Entity {
	contracts := []schema.Contract{{Type: schema.ContractSQLC}}

	return schema.Entity{
		Name:      "Country",
		Contracts: contracts,
		Indexes:   []schema.Index{{Type: schema.IndexPrimary, Columns: []schema.IndexColumn{{Name: "id"}}}},
		Fields: []schema.Field{
			{Name: "id", Type: schema.FieldTypeString, Contracts: contracts},
			{Name: "name", Type: schema.FieldTypeString, Contracts: contracts},
		},
	}
}

// wraps a whole sqlc query file, the way the sqlc-wrap command does
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

// a hand-written query can be named like a dsl query, but it selects an aggregate, so
// sqlc gives it its own row struct and the entity wrapper does not fit
func TestCustomListQueryWithOwnRowTypeIsForwarded(t *testing.T) {
	got := wrapQueryFile(t, productEntity(), nil, `type ListProductByUsageParams struct {
	Limit  int64
	Offset int64
}

type ListProductByUsageRow struct {
	Name  string
	Total int64
}

func (q *Queries) ListProductByUsage(ctx context.Context, arg ListProductByUsageParams) ([]ListProductByUsageRow, error) {
	return nil, nil
}`)

	want := []string{
		"type ListProductByUsageRow = internal.ListProductByUsageRow",
		"type ListProductByUsageParams = internal.ListProductByUsageParams",
		"func (q *Queries) ListProductByUsage(ctx context.Context, arg internal.ListProductByUsageParams) ([]internal.ListProductByUsageRow, error) {",
		"return (*internal.Queries)(q).ListProductByUsage(ctx, arg)",
	}
	for _, part := range want {
		if !strings.Contains(got, part) {
			t.Errorf("expected %q in the wrapper:\n%s", part, got)
		}
	}

	if strings.Contains(got, "ProductFromSQL") {
		t.Errorf("expected no model converter for an aggregate row:\n%s", got)
	}
}

// the same name, but the select is the plain table, so the wrapper does fit
func TestCustomListQueryOfEntityRowsIsWrapped(t *testing.T) {
	got := wrapQueryFile(t, productEntity(), nil, `func (q *Queries) ListProductByUsage(ctx context.Context) ([]Product, error) {
	return nil, nil
}`)

	want := []string{
		"func (q *Queries) ListProductByUsage(ctx context.Context) ([]*Product, error) {",
		"result[i] = ProductFromSQL(&dbResults[i])",
	}
	for _, part := range want {
		if !strings.Contains(got, part) {
			t.Errorf("expected %q in the wrapper:\n%s", part, got)
		}
	}
}

func TestCustomGetQueryWithOwnRowTypeIsForwarded(t *testing.T) {
	got := wrapQueryFile(t, productEntity(), nil, `type GetProductByErrorsRow struct {
	Name   string
	Errors int64
}

func (q *Queries) GetProductByErrors(ctx context.Context, name string) (GetProductByErrorsRow, error) {
	return GetProductByErrorsRow{}, nil
}`)

	want := "func (q *Queries) GetProductByErrors(ctx context.Context, name string) (internal.GetProductByErrorsRow, error) {"
	if !strings.Contains(got, want) {
		t.Errorf("expected %q in the wrapper:\n%s", want, got)
	}

	if strings.Contains(got, "ProductFromSQL") {
		t.Errorf("expected no model converter for an aggregate row:\n%s", got)
	}
}

// :execrows returns a count, so the delete wrapper, which returns only an error, does not fit
func TestCustomDeleteQueryReturningRowsIsForwarded(t *testing.T) {
	got := wrapQueryFile(t, productEntity(), nil, `func (q *Queries) DeleteProduct(ctx context.Context, id int64) (int64, error) {
	return 0, nil
}`)

	want := "func (q *Queries) DeleteProduct(ctx context.Context, id int64) (int64, error) {"
	if !strings.Contains(got, want) {
		t.Errorf("expected %q in the wrapper:\n%s", want, got)
	}
}

// a dsl query keeps its wrapper, its row struct comes from the same schema
func TestDslListQueryIsStillWrapped(t *testing.T) {
	entity := productEntity()
	query := schema.Query{Type: schema.QueryListBy, Name: "ListProductByName", Fields: []string{"name"}}
	got := wrapQueryFile(t, entity, []schema.Query{query}, `func (q *Queries) ListProductByName(ctx context.Context, name string) ([]Product, error) {
	return nil, nil
}`)

	want := []string{
		"func (q *Queries) ListProductByName(ctx context.Context, name string) ([]*Product, error) {",
		"result[i] = ProductFromSQL(&dbResults[i])",
	}
	for _, part := range want {
		if !strings.Contains(got, part) {
			t.Errorf("expected %q in the wrapper:\n%s", part, got)
		}
	}
}

// a hand-written insert with RETURNING * hands back the whole row, the create wrapper
// returns only an error for a caller supplied id, so the two do not fit
func TestCustomCreateQueryWithReturningRowIsForwarded(t *testing.T) {
	entity := countryEntity()
	if entity.InsertReturnsID() {
		t.Fatal("the caller supplies the id, the insert should return nothing")
	}

	got := wrapQueryFile(t, entity, nil, `type CreateCountryParams struct {
	ID   string
	Name string
}

func (q *Queries) CreateCountry(ctx context.Context, arg CreateCountryParams) (Country, error) {
	return Country{}, nil
}`)

	want := []string{
		"type CreateCountryParams = internal.CreateCountryParams",
		"func (q *Queries) CreateCountry(ctx context.Context, arg internal.CreateCountryParams) (internal.Country, error) {",
		"return (*internal.Queries)(q).CreateCountry(ctx, arg)",
	}
	for _, part := range want {
		if !strings.Contains(got, part) {
			t.Errorf("expected %q in the wrapper:\n%s", part, got)
		}
	}

	if strings.Contains(got, "internalArg") {
		t.Errorf("expected no create wrapper for a hand-written insert:\n%s", got)
	}
}

// the same, for an entity whose insert does hand an id back: RETURNING * is still not an id
func TestCustomCreateQueryWithReturningRowIsForwardedWhenInsertReturnsID(t *testing.T) {
	entity := productEntity()
	if !entity.InsertReturnsID() {
		t.Fatal("the db assigns the id, the insert should return it")
	}

	got := wrapQueryFile(t, entity, nil, `type CreateProductParams struct {
	Name string
}

func (q *Queries) CreateProduct(ctx context.Context, arg CreateProductParams) (Product, error) {
	return Product{}, nil
}`)

	want := "func (q *Queries) CreateProduct(ctx context.Context, arg internal.CreateProductParams) (internal.Product, error) {"
	if !strings.Contains(got, want) {
		t.Errorf("expected %q in the wrapper:\n%s", want, got)
	}
}

// an insert that returns the id, the shape the create wrapper is written for
func TestCustomCreateQueryReturningIdIsWrapped(t *testing.T) {
	got := wrapQueryFile(t, productEntity(), nil, `type CreateProductParams struct {
	Name string
	Sold int64
}

func (q *Queries) CreateProduct(ctx context.Context, arg CreateProductParams) (int64, error) {
	return 0, nil
}`)

	want := []string{
		"func (q *Queries) CreateProduct(ctx context.Context, arg CreateProductParams) (int64, error) {",
		"internalArg := internal.CreateProductParams{",
		"Name: arg.Name,",
		"Sold: arg.Sold,",
		"return (*internal.Queries)(q).CreateProduct(ctx, internalArg)",
	}
	for _, part := range want {
		if !strings.Contains(got, part) {
			t.Errorf("expected %q in the wrapper:\n%s", part, got)
		}
	}
}

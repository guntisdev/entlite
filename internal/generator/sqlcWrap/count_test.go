package sqlcwrap

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/guntisdev/entlite/internal/schema"
)

func countTestEntity() schema.Entity {
	contracts := []schema.Contract{{Type: schema.ContractSQLC}, {Type: schema.ContractPROTO}}

	return schema.Entity{
		Name:      "Post",
		Contracts: contracts,
		Fields: []schema.Field{
			{Name: "id", Type: schema.FieldTypeInt64, Primary: true, Contracts: contracts},
			{Name: "title", Type: schema.FieldTypeString, Contracts: contracts},
		},
	}
}

// returns the sqlc method declaration the wrapper is generated from
func parseListMethod(t *testing.T, source string) *ast.FuncDecl {
	t.Helper()

	file, err := parser.ParseFile(token.NewFileSet(), "queries.sql.go", "package internal\n\n"+source, 0)
	if err != nil {
		t.Fatalf("failed to parse source: %v", err)
	}

	for _, decl := range file.Decls {
		if funcDecl, ok := decl.(*ast.FuncDecl); ok {
			return funcDecl
		}
	}

	t.Fatal("no func declaration in source")
	return nil
}

func countContext(query schema.Query, entity schema.Entity) *generationContext {
	return &generationContext{
		inputPackageName:    "internal",
		sqlDialect:          schema.PostgreSQL,
		entityMap:           map[string]schema.Entity{entity.Name: entity},
		parsedEntities:      []schema.Entity{entity},
		dslQueries:          map[string]dslQuery{query.Name: {entity: entity, query: query}},
		filterParamsStructs: make(map[string]*ast.StructType),
	}
}

func TestListQueryWithCount(t *testing.T) {
	entity := countTestEntity()
	query := schema.Query{Type: schema.QueryListBy, Name: "ListPostByTitle", Fields: []string{"title"}, Count: true}
	funcDecl := parseListMethod(t, `func (q *Queries) ListPostByTitle(ctx context.Context, title string) ([]ListPostByTitleRow, error) {
	return nil, nil
}`)

	got := countContext(query, entity).generateListQuery(funcDecl, entity)

	want := []string{
		"func (q *Queries) ListPostByTitle(ctx context.Context, title string) ([]*Post, int64, error) {",
		"return nil, 0, err",
		"result[i] = &Post{",
		"ID: dbResults[i].ID,",
		"Title: dbResults[i].Title,",
		"totalCount = dbResults[0].TotalCount",
		"return result, totalCount, nil",
	}
	for _, part := range want {
		if !strings.Contains(got, part) {
			t.Errorf("expected %q in the wrapper:\n%s", part, got)
		}
	}

	// the count query has its own row struct, so the model converter does not fit
	if strings.Contains(got, "PostFromSQL") {
		t.Errorf("expected no model converter in a counted list:\n%s", got)
	}
}

func TestListQueryWithoutCount(t *testing.T) {
	entity := countTestEntity()
	query := schema.Query{Type: schema.QueryListBy, Name: "ListPostByTitle", Fields: []string{"title"}}
	funcDecl := parseListMethod(t, `func (q *Queries) ListPostByTitle(ctx context.Context, title string) ([]Post, error) {
	return nil, nil
}`)

	got := countContext(query, entity).generateListQuery(funcDecl, entity)

	want := []string{
		"func (q *Queries) ListPostByTitle(ctx context.Context, title string) ([]*Post, error) {",
		"return nil, err",
		"result[i] = PostFromSQL(&dbResults[i])",
		"return result, nil",
	}
	for _, part := range want {
		if !strings.Contains(got, part) {
			t.Errorf("expected %q in the wrapper:\n%s", part, got)
		}
	}

	if strings.Contains(got, "totalCount") {
		t.Errorf("expected no count in a plain list:\n%s", got)
	}
}

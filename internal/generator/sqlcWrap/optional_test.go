package sqlcwrap

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/guntisdev/entlite/internal/schema"
)

func optionalFilterTestEntity() schema.Entity {
	contracts := []schema.Contract{{Type: schema.ContractSQLC}, {Type: schema.ContractPROTO}}

	return schema.Entity{
		Name:      "Article",
		Contracts: contracts,
		Fields: []schema.Field{
			{Name: "id", Type: schema.FieldTypeInt64, Primary: true, Contracts: contracts},
			{Name: "title", Type: schema.FieldTypeString, Contracts: contracts},
			{Name: "is_featured", Type: schema.FieldTypeBool, Contracts: contracts},
			{Name: "views", Type: schema.FieldTypeInt, Contracts: contracts},
		},
	}
}

func parseParamsStruct(t *testing.T, source string) *ast.StructType {
	t.Helper()

	file, err := parser.ParseFile(token.NewFileSet(), "queries.sql.go", "package internal\n\n"+source, 0)
	if err != nil {
		t.Fatalf("failed to parse source: %v", err)
	}

	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, spec := range genDecl.Specs {
			if typeSpec, ok := spec.(*ast.TypeSpec); ok {
				if structType, ok := typeSpec.Type.(*ast.StructType); ok {
					return structType
				}
			}
		}
	}

	t.Fatal("no struct type in source")
	return nil
}

// filter.Eq/Search/Range(...).Optional() must make the param a pointer even when
// the underlying column is a required (non-nullable) field - the column's own
// nullability says nothing about whether the filter itself can be omitted.
func TestFilterParamsStructPointerFollowsFilterOptional(t *testing.T) {
	entity := optionalFilterTestEntity()
	query := schema.Query{
		Type: schema.QueryListBy,
		Name: "ListArticleFilterByIsFeaturedTitleViews",
		Filters: []schema.QueryFilter{
			{Type: schema.QueryFilterEq, Field: "is_featured", Optional: true},
			{Type: schema.QueryFilterSearch, Field: "title", Optional: false},
			{Type: schema.QueryFilterRange, Field: "views", Optional: true},
		},
	}

	structType := parseParamsStruct(t, `type ListArticleFilterByIsFeaturedTitleViewsParams struct {
	IsFeatured bool
	Title      string
	MinViews   int32
	MaxViews   int32
}`)

	got := generateFilterParamsStruct("ListArticleFilterByIsFeaturedTitleViewsParams", structType, entity, query)

	want := []string{"IsFeatured *bool", "MinViews *int32", "MaxViews *int32"}
	for _, part := range want {
		if !strings.Contains(got, part) {
			t.Errorf("expected %q in the generated struct:\n%s", part, got)
		}
	}

	// title's filter is mandatory, so it stays a plain value
	if !strings.Contains(got, "Title string") {
		t.Errorf("expected a plain Title field for a mandatory filter:\n%s", got)
	}
	if strings.Contains(got, "Title *string") {
		t.Errorf("mandatory filter should not become a pointer:\n%s", got)
	}
}

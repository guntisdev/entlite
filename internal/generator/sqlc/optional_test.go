package sqlc

import (
	"strings"
	"testing"

	"github.com/guntisdev/entlite/internal/schema"
)

func optionalFilterEntity() schema.Entity {
	contracts := []schema.Contract{{Type: schema.ContractSQLC}, {Type: schema.ContractPROTO}}

	return schema.Entity{
		Name:      "Article",
		Contracts: contracts,
		Fields: []schema.Field{
			{Name: "id", Type: schema.FieldTypeInt, Primary: true, Contracts: contracts},
			{Name: "title", Type: schema.FieldTypeString, Contracts: contracts},
			{Name: "is_featured", Type: schema.FieldTypeBool, Contracts: contracts},
			{Name: "views", Type: schema.FieldTypeInt, Contracts: contracts},
		},
		Queries: []schema.Query{
			{Type: schema.QueryListBy, Name: "ListArticleFilterByIsFeaturedTitleViews", Contracts: contracts,
				Filters: []schema.QueryFilter{
					{Type: schema.QueryFilterEq, Field: "is_featured", Optional: true},
					{Type: schema.QueryFilterSearch, Field: "title", Optional: true},
					{Type: schema.QueryFilterRange, Field: "views", Optional: true},
				}},
		},
	}
}

// An Optional() filter must be skippable at call time: omitting it should not
// narrow the result set, so the WHERE clause needs an IS NULL escape hatch
// rather than a plain mandatory comparison.
func TestOptionalFilterSkipsWhenArgIsNull(t *testing.T) {
	// sqlc.narg() is the only way to declare a nullable named arg, so it is used
	// verbatim across all three dialects (unlike the mandatory-arg helper, which
	// switches between @name and sqlc.arg() per dialect).
	want := []string{
		`(sqlc.narg('is_featured') IS NULL OR is_featured = sqlc.narg('is_featured'))`,
		`(sqlc.narg('title') IS NULL OR title LIKE sqlc.narg('title'))`,
		`(sqlc.narg('min_views') IS NULL OR views >= sqlc.narg('min_views')) AND (sqlc.narg('max_views') IS NULL OR views <= sqlc.narg('max_views'))`,
	}

	for _, dialect := range []schema.SQLDialect{schema.PostgreSQL, schema.SQLite, schema.MySQL} {
		t.Run(string(dialect), func(t *testing.T) {
			sql := NewGenerator(dialect).generateCRUDQueries(optionalFilterEntity())

			for _, part := range want {
				if !strings.Contains(sql, part) {
					t.Errorf("expected %s in the generated SQL:\n%s", part, sql)
				}
			}

			// no mandatory comparison should remain for any of the optional filters
			for _, unwanted := range []string{"is_featured = @is_featured", "title LIKE @title"} {
				if strings.Contains(sql, unwanted) {
					t.Errorf("unexpected mandatory comparison %q in generated SQL:\n%s", unwanted, sql)
				}
			}
		})
	}
}

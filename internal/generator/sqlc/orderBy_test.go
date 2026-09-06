package sqlc

import (
	"strings"
	"testing"

	"github.com/guntisdev/entlite/internal/schema"
)

func orderByEntity() schema.Entity {
	contracts := []schema.Contract{{Type: schema.ContractSQLC}, {Type: schema.ContractPROTO}}

	return schema.Entity{
		Name:      "Post",
		Contracts: contracts,
		Fields: []schema.Field{
			{Name: "ID", Type: schema.FieldTypeInt, Primary: true, Contracts: contracts},
			{Name: "title", Type: schema.FieldTypeString, Contracts: contracts},
			{Name: "created_at", Type: schema.FieldTypeTime, Contracts: contracts},
			{Name: "order", Type: schema.FieldTypeInt, Contracts: contracts},
		},
		Queries: []schema.Query{
			{Type: schema.QueryListBy, Name: "ListPostPaged", Fields: []string{"title"},
				OrderBy: "created_at", HasLimit: true, HasOffset: true, Contracts: contracts},
			{Type: schema.QueryListBy, Name: "ListPostByOrder", Fields: []string{"title"},
				OrderBy: "order", Contracts: contracts},
			{Type: schema.QueryListBy, Name: "ListPostPlain", Fields: []string{"title"}, Contracts: contracts},
		},
	}
}

func TestOrderByClause(t *testing.T) {
	tests := []struct {
		dialect schema.SQLDialect
		want    []string
	}{
		{schema.PostgreSQL, []string{
			`SELECT * FROM "post" WHERE title = @title ORDER BY created_at LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');`,
			`SELECT * FROM "post" WHERE title = @title ORDER BY "order";`,
			`SELECT * FROM "post" WHERE title = @title;`,
		}},
		{schema.SQLite, []string{
			`SELECT * FROM "post" WHERE title = @title ORDER BY created_at LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');`,
			`SELECT * FROM "post" WHERE title = @title ORDER BY "order";`,
			`SELECT * FROM "post" WHERE title = @title;`,
		}},
		{schema.MySQL, []string{
			"SELECT * FROM `post` WHERE title = sqlc.arg('title') ORDER BY created_at LIMIT ? OFFSET ?;",
			"SELECT * FROM `post` WHERE title = sqlc.arg('title') ORDER BY `order`;",
			"SELECT * FROM `post` WHERE title = sqlc.arg('title');",
		}},
	}

	for _, test := range tests {
		t.Run(string(test.dialect), func(t *testing.T) {
			sql := NewGenerator(test.dialect).generateCRUDQueries(orderByEntity())

			for _, want := range test.want {
				if !strings.Contains(sql, want) {
					t.Errorf("expected %s in the generated SQL:\n%s", want, sql)
				}
			}

			// only the two sorted queries get an ORDER BY
			if got := strings.Count(sql, "ORDER BY"); got != 2 {
				t.Errorf("expected 2 ORDER BY clauses, got %d:\n%s", got, sql)
			}
		})
	}
}

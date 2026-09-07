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
				OrderBy:  []schema.OrderColumn{{Name: "created_at"}},
				HasLimit: true, HasOffset: true, Contracts: contracts},
			{Type: schema.QueryListBy, Name: "ListPostByOrder", Fields: []string{"title"},
				OrderBy: []schema.OrderColumn{{Name: "order", Desc: true}}, Contracts: contracts},
			{Type: schema.QueryListBy, Name: "ListPostNewest", Fields: []string{"title"},
				OrderBy: []schema.OrderColumn{{Name: "created_at", Desc: true}, {Name: "title"}}, Contracts: contracts},
			{Type: schema.QueryListAll, Name: "ListAllPostSorted",
				OrderBy: []schema.OrderColumn{{Name: "created_at", Desc: true}}, Contracts: contracts},
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
			`SELECT * FROM "post" WHERE title = @title ORDER BY "order" DESC;`,
			`SELECT * FROM "post" WHERE title = @title ORDER BY created_at DESC, title;`,
			`SELECT * FROM "post" ORDER BY created_at DESC;`,
			`SELECT * FROM "post" WHERE title = @title;`,
		}},
		{schema.SQLite, []string{
			`SELECT * FROM "post" WHERE title = @title ORDER BY created_at LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');`,
			`SELECT * FROM "post" WHERE title = @title ORDER BY "order" DESC;`,
			`SELECT * FROM "post" WHERE title = @title ORDER BY created_at DESC, title;`,
			`SELECT * FROM "post" ORDER BY created_at DESC;`,
			`SELECT * FROM "post" WHERE title = @title;`,
		}},
		{schema.MySQL, []string{
			"SELECT * FROM `post` WHERE title = sqlc.arg('title') ORDER BY created_at LIMIT ? OFFSET ?;",
			"SELECT * FROM `post` WHERE title = sqlc.arg('title') ORDER BY `order` DESC;",
			"SELECT * FROM `post` WHERE title = sqlc.arg('title') ORDER BY created_at DESC, title;",
			"SELECT * FROM `post` ORDER BY created_at DESC;",
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

			// only the four sorted queries get an ORDER BY
			if got := strings.Count(sql, "ORDER BY"); got != 4 {
				t.Errorf("expected 4 ORDER BY clauses, got %d:\n%s", got, sql)
			}
		})
	}
}

package sqlc

import (
	"strings"
	"testing"

	"github.com/guntisdev/entlite/internal/schema"
)

func countEntity() schema.Entity {
	contracts := []schema.Contract{{Type: schema.ContractSQLC}, {Type: schema.ContractPROTO}}

	return schema.Entity{
		Name:      "Post",
		Contracts: contracts,
		Fields: []schema.Field{
			{Name: "id", Type: schema.FieldTypeInt, Primary: true, Contracts: contracts},
			{Name: "title", Type: schema.FieldTypeString, Contracts: contracts},
		},
		Queries: []schema.Query{
			{Type: schema.QueryListBy, Name: "ListPostPaged", Fields: []string{"title"},
				Count: true, HasLimit: true, HasOffset: true, Contracts: contracts},
			{Type: schema.QueryListAll, Name: "ListAllPostCounted", Count: true, Contracts: contracts},
			{Type: schema.QueryListBy, Name: "ListPostPlain", Fields: []string{"title"}, Contracts: contracts},
		},
	}
}

func TestCountColumn(t *testing.T) {
	tests := []struct {
		dialect schema.SQLDialect
		want    []string
	}{
		{schema.PostgreSQL, []string{
			`SELECT *, COUNT(*) OVER() AS total_size FROM "post" WHERE title = @title LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');`,
			`SELECT *, COUNT(*) OVER() AS total_size FROM "post";`,
			`SELECT * FROM "post" WHERE title = @title;`,
		}},
		{schema.SQLite, []string{
			`SELECT *, COUNT(*) OVER() AS total_size FROM "post" WHERE title = @title LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');`,
			`SELECT *, COUNT(*) OVER() AS total_size FROM "post";`,
			`SELECT * FROM "post" WHERE title = @title;`,
		}},
		{schema.MySQL, []string{
			"SELECT *, CAST(COUNT(*) OVER() AS SIGNED) AS total_size FROM `post` WHERE title = sqlc.arg('title') LIMIT ? OFFSET ?;",
			"SELECT *, CAST(COUNT(*) OVER() AS SIGNED) AS total_size FROM `post`;",
			"SELECT * FROM `post` WHERE title = sqlc.arg('title');",
		}},
	}

	for _, test := range tests {
		t.Run(string(test.dialect), func(t *testing.T) {
			sql := NewGenerator(test.dialect).generateCRUDQueries(countEntity())

			for _, want := range test.want {
				if !strings.Contains(sql, want) {
					t.Errorf("expected %s in the generated SQL:\n%s", want, sql)
				}
			}

			// only the two counted queries get a window count
			if got := strings.Count(sql, "total_size"); got != 2 {
				t.Errorf("expected 2 total_size columns, got %d:\n%s", got, sql)
			}
		})
	}
}

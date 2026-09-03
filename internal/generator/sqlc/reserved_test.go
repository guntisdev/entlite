package sqlc

import (
	"strings"
	"testing"

	"github.com/guntisdev/entlite/internal/schema"
)

// an entity whose columns are all reserved words, plus one that is not
func reservedEntity() schema.Entity {
	sqlc := []schema.Contract{{Type: schema.ContractSQLC}, {Type: schema.ContractPROTO}}

	return schema.Entity{
		Name:      "Booking",
		Contracts: sqlc,
		Fields: []schema.Field{
			{Name: "ID", Type: schema.FieldTypeInt, Primary: true, Contracts: sqlc},
			{Name: "table", Type: schema.FieldTypeString, Contracts: sqlc},
			{Name: "order", Type: schema.FieldTypeInt, Contracts: sqlc},
			{Name: "label", Type: schema.FieldTypeString, Contracts: sqlc},
		},
		Indexes: []schema.Index{{
			Type:    schema.IndexRegular,
			Columns: []schema.IndexColumn{{Name: "table"}, {Name: "order", Desc: true}},
		}},
		Queries: []schema.Query{
			{Type: schema.QueryCreate, Contracts: sqlc},
			{Type: schema.QueryGetBy, Fields: []string{"table"}, Contracts: sqlc},
			{Type: schema.QueryUpdate, Contracts: sqlc},
			{Type: schema.QueryDelete, Contracts: sqlc},
			{Type: schema.QueryListBy, Filters: []schema.QueryFilter{
				{Type: schema.QueryFilterEq, Field: "table"},
				{Type: schema.QueryFilterEq, Field: "label"},
			}, Contracts: sqlc},
		},
	}
}

func TestReservedColumnsAreQuoted(t *testing.T) {
	tests := []struct {
		dialect schema.SQLDialect
		quoted  []string
	}{
		{schema.PostgreSQL, []string{`"table"`, `"order"`}},
		{schema.SQLite, []string{`"table"`, `"order"`}},
		{schema.MySQL, []string{"`table`", "`order`"}},
	}

	for _, tt := range tests {
		t.Run(string(tt.dialect), func(t *testing.T) {
			g := NewGenerator(tt.dialect)
			sql := g.generateTableSQL(reservedEntity()) + g.generateCRUDQueries(reservedEntity())

			for _, want := range tt.quoted {
				if !strings.Contains(sql, want) {
					t.Errorf("Expected %s in the generated SQL:\n%s", want, sql)
				}
			}

			for line := range strings.SplitSeq(sql, "\n") {
				if strings.HasPrefix(strings.TrimSpace(line), "--") {
					continue
				}
				stripped := strings.NewReplacer("sqlc.arg('table')", "", "sqlc.arg('order')", "",
					"sqlc.narg('table')", "", "sqlc.narg('order')", "").Replace(line)
				for _, word := range []string{"table", "order"} {
					for _, bare := range []string{" " + word + " ", " " + word + ",", " " + word + "="} {
						if strings.Contains(stripped, bare) {
							t.Errorf("Unquoted reserved word %q in: %s", word, line)
						}
					}
				}
			}
		})
	}
}

func TestUnreservedColumnsStayBare(t *testing.T) {
	g := NewGenerator(schema.PostgreSQL)
	sql := g.generateTableSQL(reservedEntity()) + g.generateCRUDQueries(reservedEntity())

	if strings.Contains(sql, `"label"`) {
		t.Errorf("Expected label to stay unquoted:\n%s", sql)
	}
	if !strings.Contains(sql, "label = @label") {
		t.Errorf("Expected an @label arg for the unreserved column:\n%s", sql)
	}
}

func TestReservedNamedArgUsesSqlcArg(t *testing.T) {
	for _, dialect := range []schema.SQLDialect{schema.PostgreSQL, schema.SQLite} {
		g := NewGenerator(dialect)
		// postgres cannot parse @table, @order or @limit
		if got := g.namedArg("table"); got != "sqlc.arg('table')" {
			t.Errorf("%s: namedArg(table) = %s", dialect, got)
		}
		if got := g.namedArg("label"); got != "@label" {
			t.Errorf("%s: namedArg(label) = %s", dialect, got)
		}
	}
}

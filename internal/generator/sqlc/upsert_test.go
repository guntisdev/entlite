package sqlc

import (
	"strings"
	"testing"

	"github.com/guntisdev/entlite/internal/schema"
)

var upsertContracts = []schema.Contract{{Type: schema.ContractSQLC}, {Type: schema.ContractPROTO}}

// one unique field, so mysql cannot pick the wrong key
func upsertEntity(query schema.Query) schema.Entity {
	return schema.Entity{
		Name:      "Post",
		Contracts: upsertContracts,
		Fields: []schema.Field{
			{Name: "ID", Type: schema.FieldTypeInt, Primary: true, Unique: true, Contracts: upsertContracts},
			{Name: "slug", Type: schema.FieldTypeString, Unique: true, Contracts: upsertContracts},
			{Name: "title", Type: schema.FieldTypeString, Contracts: upsertContracts},
			{Name: "views", Type: schema.FieldTypeInt, Contracts: upsertContracts},
			{Name: "created_at", Type: schema.FieldTypeTime, Immutable: true, Contracts: upsertContracts},
		},
		Queries: []schema.Query{query},
	}
}

func createUpsert(fields []string, ignore bool) schema.Query {
	return schema.Query{
		Type:         schema.QueryCreate,
		Name:         "CreatePost",
		Upsert:       true,
		UpsertFields: fields,
		UpsertIgnore: ignore,
		Contracts:    upsertContracts,
	}
}

func TestUpsertClause(t *testing.T) {
	tests := []struct {
		name    string
		dialect schema.SQLDialect
		query   schema.Query
		want    []string
		absent  []string
	}{
		{
			name:    "postgres updates from excluded",
			dialect: schema.PostgreSQL,
			query:   createUpsert([]string{"slug"}, false),
			want: []string{
				"ON CONFLICT (slug) DO UPDATE SET",
				"  title = excluded.title,",
				"  views = excluded.views",
				"RETURNING ID;",
			},
			// the target, the key and an immutable column are never overwritten
			absent: []string{"slug = excluded.slug", "ID = excluded.ID", "created_at = excluded.created_at"},
		},
		{
			name:    "sqlite updates from excluded",
			dialect: schema.SQLite,
			query:   createUpsert([]string{"slug"}, false),
			want:    []string{"ON CONFLICT (slug) DO UPDATE SET", "  title = excluded.title,", "RETURNING ID;"},
		},
		{
			name:    "postgres does nothing",
			dialect: schema.PostgreSQL,
			query:   createUpsert([]string{"slug"}, true),
			want:    []string{"ON CONFLICT (slug) DO NOTHING\nRETURNING ID;"},
			absent:  []string{"DO UPDATE"},
		},
		{
			name:    "mysql has no conflict target",
			dialect: schema.MySQL,
			query:   createUpsert([]string{"slug"}, false),
			want: []string{
				"ON DUPLICATE KEY UPDATE",
				"  title = VALUES(title),",
				"  views = VALUES(views);",
			},
			absent: []string{"ON CONFLICT", "excluded.", "RETURNING"},
		},
		{
			name:    "mysql ignores with a no-op assignment",
			dialect: schema.MySQL,
			query:   createUpsert([]string{"slug"}, true),
			want:    []string{"ON DUPLICATE KEY UPDATE slug = slug;"},
			// INSERT IGNORE would hide unrelated errors too
			absent: []string{"INSERT IGNORE", "DO NOTHING"},
		},
		{
			name:    "no upsert leaves the insert alone",
			dialect: schema.PostgreSQL,
			query:   schema.Query{Type: schema.QueryCreate, Name: "CreatePost", Contracts: upsertContracts},
			want:    []string{") RETURNING ID;"},
			absent:  []string{"ON CONFLICT", "ON DUPLICATE KEY"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sql := NewGenerator(test.dialect).generateCRUDQueries(upsertEntity(test.query))

			for _, want := range test.want {
				if !strings.Contains(sql, want) {
					t.Errorf("expected %q in the generated SQL:\n%s", want, sql)
				}
			}
			for _, absent := range test.absent {
				if strings.Contains(sql, absent) {
					t.Errorf("did not expect %q in the generated SQL:\n%s", absent, sql)
				}
			}
		})
	}
}

// the conflict target defaults to the primary key
func TestUpsertCompoundKey(t *testing.T) {
	entity := schema.Entity{
		Name:      "Setting",
		Contracts: upsertContracts,
		Fields: []schema.Field{
			{Name: "country", Type: schema.FieldTypeString, Contracts: upsertContracts},
			{Name: "env", Type: schema.FieldTypeString, Contracts: upsertContracts},
			{Name: "value", Type: schema.FieldTypeString, Contracts: upsertContracts},
		},
		Indexes: []schema.Index{{
			Type:    schema.IndexPrimary,
			Columns: []schema.IndexColumn{{Name: "country"}, {Name: "env"}},
		}},
		Queries: []schema.Query{{
			Type: schema.QueryCreate, Name: "CreateSetting", Upsert: true, Contracts: upsertContracts,
		}},
	}

	sql := NewGenerator(schema.PostgreSQL).generateCRUDQueries(entity)

	for _, want := range []string{"ON CONFLICT (country, env) DO UPDATE SET", "  value = excluded.value"} {
		if !strings.Contains(sql, want) {
			t.Errorf("expected %q in the generated SQL:\n%s", want, sql)
		}
	}
	// the key columns are the target, they are not overwritten
	for _, absent := range []string{"country = excluded.country", "env = excluded.env"} {
		if strings.Contains(sql, absent) {
			t.Errorf("did not expect %q in the generated SQL:\n%s", absent, sql)
		}
	}
}

func TestUpsertMySQLWarning(t *testing.T) {
	// a second unique field mysql cannot be told to ignore
	twoUnique := func() schema.Entity {
		entity := upsertEntity(createUpsert([]string{"slug"}, false))
		entity.Fields = append(entity.Fields, schema.Field{
			Name: "code", Type: schema.FieldTypeString, Unique: true, Contracts: upsertContracts,
		})
		return entity
	}

	tests := []struct {
		name    string
		dialect schema.SQLDialect
		entity  schema.Entity
		want    string
	}{
		{
			name:    "one unique key is unambiguous",
			dialect: schema.MySQL,
			entity:  upsertEntity(createUpsert([]string{"slug"}, false)),
		},
		{
			name:    "a second unique key is not",
			dialect: schema.MySQL,
			entity:  twoUnique(),
			want:    "so it ignores Upsert(slug) and fires on (code) as well",
		},
		{
			name:    "postgres honours the target, so it never warns",
			dialect: schema.PostgreSQL,
			entity:  twoUnique(),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewGenerator(test.dialect)
			sql := g.generateCRUDQueries(test.entity)

			if test.want == "" {
				if len(g.warnings) != 0 {
					t.Fatalf("expected no warnings, got %v", g.warnings)
				}
				if strings.Contains(sql, "-- warning:") {
					t.Errorf("expected no warning comment:\n%s", sql)
				}
				return
			}

			if len(g.warnings) != 1 {
				t.Fatalf("expected 1 warning, got %v", g.warnings)
			}
			if !strings.Contains(g.warnings[0], test.want) {
				t.Errorf("expected warning containing %q, got %q", test.want, g.warnings[0])
			}
			// the warning also survives in the generated file
			if !strings.Contains(sql, "-- warning: mysql ON DUPLICATE KEY UPDATE") {
				t.Errorf("expected the warning as a comment:\n%s", sql)
			}
		})
	}
}

// a unique index that contains a unique field can only fire when that field does
func TestUpsertRedundantConstraintDoesNotWarn(t *testing.T) {
	entity := upsertEntity(createUpsert([]string{"slug"}, false))
	entity.Indexes = []schema.Index{{
		Type:    schema.IndexRegular,
		Unique:  true,
		Columns: []schema.IndexColumn{{Name: "slug"}, {Name: "title"}},
	}}

	g := NewGenerator(schema.MySQL)
	g.generateCRUDQueries(entity)

	if len(g.warnings) != 0 {
		t.Errorf("expected no warnings, got %v", g.warnings)
	}
}

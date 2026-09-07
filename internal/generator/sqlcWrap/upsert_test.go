package sqlcwrap

import (
	"strings"
	"testing"

	"github.com/guntisdev/entlite/internal/schema"
)

func bulkUpsertEntity(ignore bool) schema.Entity {
	contracts := []schema.Contract{{Type: schema.ContractSQLC}, {Type: schema.ContractPROTO}}

	return schema.Entity{
		Name:      "Post",
		Contracts: contracts,
		Fields: []schema.Field{
			{Name: "ID", Type: schema.FieldTypeInt64, Primary: true, Contracts: contracts},
			{Name: "slug", Type: schema.FieldTypeString, Unique: true, Contracts: contracts},
			{Name: "title", Type: schema.FieldTypeString, Contracts: contracts},
		},
		Queries: []schema.Query{{
			Type:         schema.QueryCreateBulk,
			Name:         "CreateBulkPost",
			Upsert:       true,
			UpsertFields: []string{"slug"},
			UpsertIgnore: ignore,
			Contracts:    contracts,
		}},
	}
}

const bulkInsertSource = `func (q *Queries) CreateBulkPost(ctx context.Context, arg CreateBulkPostParams) (int64, error) {
	return 0, nil
}`

// an ignored row is one skipped row, the rest of the batch still commits
func TestCreateBulkIgnoreSkipsRow(t *testing.T) {
	tests := []struct {
		name    string
		ignore  bool
		dialect schema.SQLDialect
		wantErr bool
	}{
		{name: "ignore on sqlite", ignore: true, dialect: schema.SQLite, wantErr: true},
		{name: "ignore on postgres", ignore: true, dialect: schema.PostgreSQL, wantErr: true},
		// mysql has no RETURNING, so the insert never reports ErrNoRows
		{name: "ignore on mysql", ignore: true, dialect: schema.MySQL, wantErr: false},
		{name: "upsert that updates always returns a row", ignore: false, dialect: schema.PostgreSQL, wantErr: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			entity := bulkUpsertEntity(test.ignore)
			funcDecl := parseListMethod(t, bulkInsertSource)
			got := generateCreateBulkQuery(funcDecl, entity, "internal", test.dialect)

			hasErrCheck := strings.Contains(got, "errors.Is(err, sql.ErrNoRows)")
			if hasErrCheck != test.wantErr {
				t.Fatalf("errors.Is(err, sql.ErrNoRows) present: %v, expected %v\n%s", hasErrCheck, test.wantErr, got)
			}
			if !test.wantErr {
				return
			}

			for _, want := range []string{
				"var kept int64",
				"results = append(results, kept)",
				"continue",
				"// A row the upsert kept holds the zero id, so the results line up with args.",
			} {
				if !strings.Contains(got, want) {
					t.Errorf("expected %q in the wrapper:\n%s", want, got)
				}
			}

			// a real failure still aborts the batch
			if !strings.Contains(got, "return nil, err") {
				t.Errorf("expected other errors to abort:\n%s", got)
			}
		})
	}
}

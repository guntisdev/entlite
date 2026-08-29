package parser

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/guntisdev/entlite/internal/schema"
)

const commentEntitySource = `package schema

import (
	"github.com/guntisdev/entlite/pkg/entlite"
	"github.com/guntisdev/entlite/pkg/entlite/field"
)

// User is a person with an account.
// Second doc line.
type User struct {
	entlite.Schema
}

func (User) Contracts() []entlite.Contract {
	return []entlite.Contract{
		entlite.SQLC(),
		entlite.PROTO(),
	}
}

func (User) Fields() []entlite.Field {
	return []entlite.Field{
		// Login address
		field.String("email").Unique(),

		// --- section note, detached by the blank line below ---

		field.String("name"),
		field.Int("age").Optional(), // trailing note, not a doc
		field.Bool("is_active").Default(true),
		//go:generate something
		field.String("captcha").Contracts(entlite.PROTO()),
		// Free-form,
		// over two lines
		field.String("bio").Optional(),
	}
}
`

func parseCommentEntity(t *testing.T) schema.Entity {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "user.go")
	if err := os.WriteFile(path, []byte(commentEntitySource), 0644); err != nil {
		t.Fatalf("failed to write entity file: %v", err)
	}

	entities, err := ParseEntities([]DiscoveredEntity{{Name: "User", Path: path}})
	if err != nil {
		t.Fatalf("failed to parse entity: %v", err)
	}

	return entities[0]
}

func TestEntityComment(t *testing.T) {
	entity := parseCommentEntity(t)

	want := "User is a person with an account.\nSecond doc line."
	if entity.Comment != want {
		t.Errorf("entity comment = %q, want %q", entity.Comment, want)
	}
}

func TestFieldComments(t *testing.T) {
	entity := parseCommentEntity(t)

	tests := []struct {
		field string
		want  string
	}{
		{"email", "Login address"},
		{"name", ""},
		{"age", ""},
		{"is_active", ""},
		{"captcha", ""},
		{"bio", "Free-form,\nover two lines"},
	}

	for _, test := range tests {
		field, ok := entity.GetFieldByName(test.field)
		if !ok {
			t.Fatalf("field %q not found", test.field)
		}
		if field.Comment != test.want {
			t.Errorf("field %q comment = %q, want %q", test.field, field.Comment, test.want)
		}
	}
}

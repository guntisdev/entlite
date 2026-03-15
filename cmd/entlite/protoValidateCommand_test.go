package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/guntisdev/entlite/internal/util"
)

func TestProtoValidateCommand(t *testing.T) {

	tmpDir := t.TempDir()

	schemaDir := filepath.Join(tmpDir, "ent", "schema")
	logicDir := filepath.Join(tmpDir, "ent", "logic")
	pbDir := filepath.Join(tmpDir, "ent", "gen", "pb")

	if err := os.MkdirAll(schemaDir, 0755); err != nil {
		t.Fatalf("Failed to create schema directory: %v", err)
	}

	if err := os.MkdirAll(logicDir, 0755); err != nil {
		t.Fatalf("Failed to create logic directory: %v", err)
	}

	if err := os.MkdirAll(pbDir, 0755); err != nil {
		t.Fatalf("Failed to create pb directory: %v", err)
	}

	goModContent := `module github.com/guntisdev/entlite/examples/test-validate

go 1.26.0
`
	goModPath := filepath.Join(tmpDir, "go.mod")
	if err := os.WriteFile(goModPath, []byte(goModContent), 0644); err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}

	userSchemaContent := `package ent

import (
	"time"

	"github.com/guntisdev/entlite/examples/test-validate/ent/logic"
	"github.com/guntisdev/entlite/pkg/entlite"
	"github.com/guntisdev/entlite/pkg/entlite/field"
)

type User struct {
	entlite.Schema
}

func (User) Annotations() []entlite.Annotation {
	return []entlite.Annotation{
		entlite.Message(),
		entlite.Service(),
	}
}

func (User) Fields() []entlite.Field {
	return []entlite.Field{
		field.String("email").Unique().ProtoField(2).Validate(logic.IsValidEmail),
		field.String("name").Validate(logic.StartsWithCapital).Comment("First name and surname"),
		field.Int32("age").Optional(),
		field.Float("score").Default(0.0),
		field.String("uuid").Immutable().DefaultFunc(logic.GetUuidStr),
		field.Bool("is_admin").ProtoField(5),
		field.Byte("api_key").DefaultFunc(logic.GenerateAPIKey).Immutable(),
		field.Time("created_at").DefaultFunc(time.Now).ProtoField(6).Immutable(),
		field.Time("updated_at").DefaultFunc(time.Now).ProtoField(7),
	}
}`

	userSchemaPath := filepath.Join(schemaDir, "user.go")
	if err := os.WriteFile(userSchemaPath, []byte(userSchemaContent), 0644); err != nil {
		t.Fatalf("Failed to write user schema: %v", err)
	}

	postSchemaContent := `package ent

import (
	"github.com/guntisdev/entlite/examples/test-validate/ent/logic"
	"github.com/guntisdev/entlite/pkg/entlite"
	"github.com/guntisdev/entlite/pkg/entlite/field"
)

type Post struct {
	entlite.Schema
}

func (Post) Annotations() []entlite.Annotation {
	return []entlite.Annotation{
		entlite.Message(),
		entlite.Service(),
	}
}

func (Post) Fields() []entlite.Field {
	return []entlite.Field{
		field.String("title").Validate(logic.MinLength5),
		field.String("content"),
		field.Int32("user_id"),
	}
}`

	postSchemaPath := filepath.Join(schemaDir, "post.go")
	if err := os.WriteFile(postSchemaPath, []byte(postSchemaContent), 0644); err != nil {
		t.Fatalf("Failed to write post schema: %v", err)
	}

	logicContent := `package logic

import (
	"crypto/rand"
	"fmt"
	"strings"
	"unicode"

	"github.com/google/uuid"
)

func GetUuidStr() string {
	return uuid.New().String()
}

func StartsWithCapital(s string) bool {
	if len(s) == 0 {
		return false
	}
	return unicode.IsUpper(rune(s[0]))
}

func IsValidEmail(s string) bool {
	return strings.Contains(s, "@") && strings.Contains(s, ".")
}

func MinLength5(s string) bool {
	return len(s) >= 5
}

func GenerateAPIKey() []byte {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		panic(fmt.Sprintf("failed to generate secure random bytes: %v", err))
	}
	return key
}`

	logicPath := filepath.Join(logicDir, "logic.go")
	if err := os.WriteFile(logicPath, []byte(logicContent), 0644); err != nil {
		t.Fatalf("Failed to write logic file: %v", err)
	}

	// Change to tmpDir/ent to simulate working in the project
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	if err := os.Chdir(filepath.Join(tmpDir, "ent")); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}
	defer os.Chdir(originalDir)

	protoValidate([]string{pbDir})

	validatePath := filepath.Join(pbDir, "proto_validate.go")
	if _, err := os.Stat(validatePath); os.IsNotExist(err) {
		t.Fatalf("Expected proto_validate.go was not created at %s", validatePath)
	}

	actualContent, err := os.ReadFile(validatePath)
	if err != nil {
		t.Fatalf("Failed to read generated proto_validate.go: %v", err)
	}

	expectedContent := `package pb

import (
	"fmt"
	"github.com/guntisdev/entlite/examples/test-validate/ent/logic"
)

func (r *CreatePostRequest) Validate() error {
	if !logic.MinLength5(r.Title) {
		return fmt.Errorf("Validation failed for field name: Title")
	}
	return nil
}

func (r *UpdatePostRequest) Validate() error {
	if !logic.MinLength5(r.Title) {
		return fmt.Errorf("Validation failed for field name: Title")
	}
	return nil
}

func (r *CreateUserRequest) Validate() error {
	if !logic.IsValidEmail(r.Email) {
		return fmt.Errorf("Validation failed for field name: Email")
	}
	if !logic.StartsWithCapital(r.Name) {
		return fmt.Errorf("Validation failed for field name: Name")
	}
	return nil
}

func (r *UpdateUserRequest) Validate() error {
	if !logic.IsValidEmail(r.Email) {
		return fmt.Errorf("Validation failed for field name: Email")
	}
	if !logic.StartsWithCapital(r.Name) {
		return fmt.Errorf("Validation failed for field name: Name")
	}
	return nil
}`

	if d := util.Diff(expectedContent, string(actualContent)); d != "" {
		t.Errorf("proto_validate.go content mismatch (-expected +actual):\n%s", d)
	}
}

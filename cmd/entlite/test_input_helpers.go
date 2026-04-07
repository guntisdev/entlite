package main

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTestGoMod(t *testing.T, tmpDir string) {
	t.Helper()

	goModContent := `module github.com/guntisdev/entlite/examples/01-basic-entity

go 1.26.0
`
	goModPath := filepath.Join(tmpDir, "go.mod")
	if err := os.WriteFile(goModPath, []byte(goModContent), 0644); err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}
}

func writeTestUserSchema(t *testing.T, schemaDir string) {
	t.Helper()

	userSchemaContent := `package ent

import (
	"time"

	"github.com/guntisdev/entlite/examples/01-basic-entity/ent/logic"
	"github.com/guntisdev/entlite/pkg/entlite"
	"github.com/guntisdev/entlite/pkg/entlite/field"
	"github.com/guntisdev/entlite/pkg/entlite/permissions"
)

type User struct {
	entlite.Schema
}

func (User) Annotations() []entlite.Annotation {
	return []entlite.Annotation{
		entlite.Message(),
		entlite.GRPC(),
	}
}

func (User) Fields() []entlite.Field {
	return []entlite.Field{
		field.String("email").Unique().ProtoField(2),
		field.String("name").Validate(logic.StartsWithCapital).Comment("First name and surname"),
		field.Int("age").Optional(),
		field.String("password").Permissions(permissions.WriteOnly),
		field.Float("score").Default(4.2),
		field.String("uuid").Immutable().DefaultFunc(logic.GetUuidStr),
		field.Bool("is_admin").ProtoField(5),
		field.Byte("api_key").DefaultFunc(logic.GenerateAPIKey),
		field.Int64("last_login_ms"),
		field.Time("created_at").Permissions(permissions.ReadOnly).DefaultFunc(time.Now).ProtoField(6).Immutable(),
		field.Time("updated_at").Permissions(permissions.ReadOnly).DefaultFunc(time.Now).ProtoField(7),
	}
}`

	userSchemaPath := filepath.Join(schemaDir, "user.go")
	if err := os.WriteFile(userSchemaPath, []byte(userSchemaContent), 0644); err != nil {
		t.Fatalf("Failed to write user schema: %v", err)
	}
}

func writeTestLogic(t *testing.T, logicDir string) {
	t.Helper()

	logicContent := `package logic

import (
	"crypto/rand"
	"fmt"
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
}

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenCommandFunction(t *testing.T) {
	tmpDir := t.TempDir()

	// Create directory structure
	schemaDir := filepath.Join(tmpDir, "ent", "schema")
	logicDir := filepath.Join(tmpDir, "ent", "logic")

	if err := os.MkdirAll(schemaDir, 0755); err != nil {
		t.Fatalf("Failed to create schema directory: %v", err)
	}

	if err := os.MkdirAll(logicDir, 0755); err != nil {
		t.Fatalf("Failed to create logic directory: %v", err)
	}

	// Create schema/user.go
	userSchemaContent := `package ent

import (
	"time"

	"github.com/guntisdev/entlite/examples/01-basic-entity/ent/logic"
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
		field.String("email").Unique().ProtoField(2),
		field.String("name").Validate(logic.StartsWithCapital).Comment("First name and surname"),
		field.Int32("age").Optional(),
		field.String("uuid").Immutable().DefaultFunc(logic.GetUuidStr),
		field.Bool("is_admin").ProtoField(5),
		field.Time("created_at").DefaultFunc(time.Now).ProtoField(6).Immutable(),
		field.Time("updated_at").DefaultFunc(time.Now).ProtoField(7),
	}
}
`

	userSchemaPath := filepath.Join(schemaDir, "user.go")
	if err := os.WriteFile(userSchemaPath, []byte(userSchemaContent), 0644); err != nil {
		t.Fatalf("Failed to write user schema: %v", err)
	}

	// Create logic/logic.go with helper functions
	logicContent := `package logic

import (
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
`

	logicPath := filepath.Join(logicDir, "logic.go")
	if err := os.WriteFile(logicPath, []byte(logicContent), 0644); err != nil {
		t.Fatalf("Failed to write logic file: %v", err)
	}

	genCommand([]string{schemaDir})

	expectedDirs := []string{
		filepath.Join(tmpDir, "ent", "contract", "proto"),
		filepath.Join(tmpDir, "ent", "contract", "sqlc"),
		filepath.Join(tmpDir, "ent", "gen"),
	}
	expectedFiles := []string{
		filepath.Join(tmpDir, "ent", "contract", "proto", "schema.proto"),
		filepath.Join(tmpDir, "ent", "contract", "sqlc", "schema.sql"),
		filepath.Join(tmpDir, "ent", "contract", "sqlc", "queries.sql"),
	}

	for _, dir := range expectedDirs {
		if info, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("Expected directory %s was not created", dir)
		} else if !info.IsDir() {
			t.Errorf("Expected %s to be a directory", dir)
		}
	}

	for _, file := range expectedFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Errorf("Expected file %s was not created", file)
		}
	}

	// Verify schema.proto
	protoPath := filepath.Join(tmpDir, "ent", "contract", "proto", "schema.proto")
	if content, err := os.ReadFile(protoPath); err == nil {
		protoContent := string(content)

		expectedProtoStrings := []string{
			"syntax = \"proto3\"",
			"package entlite",
			"message User",
			"string email = 2",
			"string name = ",
			"int32 age = ",
			"bool is_admin = 5",
			"google.protobuf.Timestamp created_at = 6",
			"google.protobuf.Timestamp updated_at = 7",
			"service UserService",
		}

		for _, expected := range expectedProtoStrings {
			if !strings.Contains(protoContent, expected) {
				t.Errorf("Proto file missing expected content: %s", expected)
			}
		}
	}

	// Verify schema.sql
	sqlSchemaPath := filepath.Join(tmpDir, "ent", "contract", "sqlc", "schema.sql")
	if content, err := os.ReadFile(sqlSchemaPath); err == nil {
		sqlContent := string(content)

		// Check for expected SQL content
		expectedSQLStrings := []string{
			"CREATE TABLE",
			"user",
			"email",
			"name",
			"age",
			"is_admin",
			"created_at",
			"updated_at",
		}

		for _, expected := range expectedSQLStrings {
			if !strings.Contains(sqlContent, expected) {
				t.Errorf("SQL schema file missing expected content: %s", expected)
			}
		}
	}

	// Verify queries.sql
	sqlQueriesPath := filepath.Join(tmpDir, "ent", "contract", "sqlc", "queries.sql")
	if content, err := os.ReadFile(sqlQueriesPath); err == nil {
		queriesContent := string(content)

		expectedQueryStrings := []string{
			"-- name:",
			"User",
		}

		for _, expected := range expectedQueryStrings {
			if !strings.Contains(queriesContent, expected) {
				t.Errorf("SQL queries file missing expected content: %s", expected)
			}
		}
	}
}

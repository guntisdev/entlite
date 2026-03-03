package main

import (
	"os"
	"path/filepath"
	"testing"

	testutil "github.com/guntisdev/entlite/internal/util"
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
	expectedProtoContent := `syntax = "proto3";

package entlite;

option go_package = "./pb";

import "google/protobuf/timestamp.proto";
import "google/protobuf/empty.proto";

// User represents as user entity
message User {
  int32 id = 1;
  string email = 2;
  // First name and surname
  string name = 3;
  optional int32 age = 4;
  string uuid = 8;
  bool is_admin = 5;
  google.protobuf.Timestamp created_at = 6;
  google.protobuf.Timestamp updated_at = 7;
}

message CreateUserRequest {
  string email = 2;
  // First name and surname
  string name = 3;
  optional int32 age = 4;
  bool is_admin = 5;
}
message GetUserRequest {
  int32 id = 1;
}
message UpdateUserRequest {
  int32 id = 1;
  string email = 2;
  // First name and surname
  string name = 3;
  optional int32 age = 4;
  bool is_admin = 5;
}
message DeleteUserRequest {
  int32 id = 1;
}
message ListUserRequest {
  int32 limit = 1;
  int32 offset = 2;
}

message ListUserResponse {
  repeated User users = 1;
}

// UserService provides CRUD opertions for User entities
service UserService {
  rpc Create(CreateUserRequest) returns (User);
  rpc Get(GetUserRequest) returns (User);
  rpc Update(UpdateUserRequest) returns (User);
  rpc Delete(DeleteUserRequest) returns (google.protobuf.Empty);
  rpc List(ListUserRequest) returns (ListUserResponse);
}`

	if content, err := os.ReadFile(protoPath); err != nil {
		t.Fatalf("Failed to read proto file: %v", err)
	} else {
		actualContent := string(content)
		if d := testutil.Diff(expectedProtoContent, actualContent); d != "" {
			t.Errorf("Proto file content mismatch (-expected +actual):\n%s", d)
		}
	}

	// Verify schema.sql
	sqlSchemaPath := filepath.Join(tmpDir, "ent", "contract", "sqlc", "schema.sql")
	expectedSQLSchema := `-- Generated schema.sql
-- This file contains table definitions for all entities

-- user table
CREATE TABLE "user"(
  id SERIAL PRIMARY KEY,
  email TEXT UNIQUE NOT NULL,
  name TEXT NOT NULL,
  age INTEGER,
  uuid TEXT NOT NULL,
  is_admin BOOLEAN NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL
);`

	if content, err := os.ReadFile(sqlSchemaPath); err != nil {
		t.Fatalf("Failed to read SQL schema file: %v", err)
	} else {
		actualContent := string(content)
		if d := testutil.Diff(expectedSQLSchema, actualContent); d != "" {
			t.Errorf("SQL schema content mismatch (-expected +actual):\n%s", d)
		}
	}

	// Verify queries.sql
	sqlQueriesPath := filepath.Join(tmpDir, "ent", "contract", "sqlc", "queries.sql")
	expectedSQLQueries := `-- Generate queries.sql
-- This file contains SQLC-compatible queries definitions

-- User CRUD operations

-- name: CreateUser :one
INSERT INTO "user" (
  email,
  name,
  age,
  uuid,
  is_admin,
  created_at,
  updated_at
) VALUES (
  $1,
  $2,
  $3,
  $4,
  $5,
  $6,
  $7
) RETURNING id;

-- name: GetUser :one
SELECT * FROM "user" WHERE id = $1;

-- name: ListUser :many
SELECT * FROM "user" ORDER BY id;

-- name: UpdateUser :one
UPDATE "user" SET
  email = $1,
  name = $2,
  age = $3,
  is_admin = $4,
  updated_at = $5
WHERE id = $6
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM "user" WHERE id = $1;`

	if content, err := os.ReadFile(sqlQueriesPath); err != nil {
		t.Fatalf("Failed to read SQL queries file: %v", err)
	} else {
		actualContent := string(content)
		if d := testutil.Diff(expectedSQLQueries, actualContent); d != "" {
			t.Errorf("SQL queries content mismatch (-expected +actual):\n%s", d)
		}
	}
}

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

	// Write common test input files
	writeTestGoMod(t, tmpDir)
	writeTestUserSchema(t, schemaDir)
	writeTestLogic(t, logicDir)

	// Create sqlc.yaml
	sqlcYamlContent := `version: "2"
sql:
  - schema: "contract/sqlc/schema.sql"
    queries: "contract/sqlc/queries.sql"    
    engine: "postgresql"
    gen:
      go:
        package: "internal"
        out: "gen/db/internal"
        emit_json_tags: true`

	sqlcYamlPath := filepath.Join(tmpDir, "ent", "sqlc.yaml")
	if err := os.WriteFile(sqlcYamlPath, []byte(sqlcYamlContent), 0644); err != nil {
		t.Fatalf("Failed to write sqlc.yaml: %v", err)
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
import "buf/validate/validate.proto";

// User represents as user entity
message User {
  int32 ID = 1 [(buf.validate.field).required = true];
  string email = 2 [(buf.validate.field).required = true];
  // First name and surname
  string name = 3 [(buf.validate.field).required = true];
  optional int32 age = 4;
  double score = 9 [(buf.validate.field).required = true];
  string uuid = 10 [(buf.validate.field).required = true];
  bool is_admin = 5;
  bytes api_key = 11 [(buf.validate.field).required = true];
  int64 last_login_ms = 12 [(buf.validate.field).required = true];
  google.protobuf.Timestamp created_at = 6 [(buf.validate.field).required = true];
  google.protobuf.Timestamp updated_at = 7 [(buf.validate.field).required = true];
}

message CreateUserRequest {
  string email = 2 [(buf.validate.field).required = true];
  // First name and surname
  string name = 3 [(buf.validate.field).required = true];
  optional int32 age = 4;
  string password = 8 [(buf.validate.field).required = true];
  optional double score = 9;
  optional string uuid = 10;
  bool is_admin = 5;
  optional bytes api_key = 11;
  int64 last_login_ms = 12 [(buf.validate.field).required = true];
}
message GetUserRequest {
  int32 ID = 1 [(buf.validate.field).required = true];
}
message UpdateUserRequest {
  int32 ID = 1 [(buf.validate.field).required = true];
  string email = 2 [(buf.validate.field).required = true];
  // First name and surname
  string name = 3 [(buf.validate.field).required = true];
  optional int32 age = 4;
  optional string password = 8;
  optional double score = 9;
  bool is_admin = 5 [(buf.validate.field).required = true];
  optional bytes api_key = 11;
  int64 last_login_ms = 12 [(buf.validate.field).required = true];
}
message DeleteUserRequest {
  int32 ID = 1 [(buf.validate.field).required = true];
}
message ListUserRequest {
  int32 limit = 1 [(buf.validate.field).required = true];
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
  ID SERIAL PRIMARY KEY,
  email TEXT UNIQUE NOT NULL,
  name TEXT NOT NULL,
  age INT,
  password TEXT NOT NULL,
  score DOUBLE PRECISION DEFAULT 4.2 NOT NULL,
  uuid TEXT NOT NULL,
  is_admin BOOLEAN NOT NULL,
  api_key BYTEA NOT NULL,
  last_login_ms BIGINT NOT NULL,
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
  password,
  score,
  uuid,
  is_admin,
  api_key,
  last_login_ms,
  created_at,
  updated_at
) VALUES (
  $1,
  $2,
  $3,
  $4,
  $5,
  $6,
  $7,
  $8,
  $9,
  $10,
  $11
) RETURNING ID;

-- name: GetUser :one
SELECT * FROM "user" WHERE ID = $1;

-- name: ListUser :many
SELECT * FROM "user" ORDER BY ID;

-- name: UpdateUser :one
UPDATE "user" SET
  email = sqlc.arg('email'),
  name = sqlc.arg('name'),
  age = sqlc.arg('age'),
  password = COALESCE(sqlc.narg('password'), password),
  score = COALESCE(sqlc.narg('score'), score),
  is_admin = sqlc.arg('is_admin'),
  api_key = COALESCE(sqlc.narg('api_key'), api_key),
  last_login_ms = sqlc.arg('last_login_ms'),
  updated_at = sqlc.arg('updated_at')
WHERE ID = sqlc.arg('ID')
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM "user" WHERE ID = $1;`

	if content, err := os.ReadFile(sqlQueriesPath); err != nil {
		t.Fatalf("Failed to read SQL queries file: %v", err)
	} else {
		actualContent := string(content)
		if d := testutil.Diff(expectedSQLQueries, actualContent); d != "" {
			t.Errorf("SQL queries content mismatch (-expected +actual):\n%s", d)
		}
	}
}

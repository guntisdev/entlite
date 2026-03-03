package main

import (
	"os"
	"path/filepath"
	"testing"

	testutil "github.com/guntisdev/entlite/internal/util"
)

func TestConvertCommand(t *testing.T) {
	t.Skip("TODO: Fix test - will return to it later")

	tmpDir := t.TempDir()

	// Create directory structure
	schemaDir := filepath.Join(tmpDir, "ent", "schema")
	logicDir := filepath.Join(tmpDir, "ent", "logic")
	dbDir := filepath.Join(tmpDir, "ent", "gen", "db")
	pbDir := filepath.Join(tmpDir, "ent", "gen", "pb")

	if err := os.MkdirAll(schemaDir, 0755); err != nil {
		t.Fatalf("Failed to create schema directory: %v", err)
	}

	if err := os.MkdirAll(logicDir, 0755); err != nil {
		t.Fatalf("Failed to create logic directory: %v", err)
	}

	if err := os.MkdirAll(dbDir, 0755); err != nil {
		t.Fatalf("Failed to create db directory: %v", err)
	}

	if err := os.MkdirAll(pbDir, 0755); err != nil {
		t.Fatalf("Failed to create pb directory: %v", err)
	}

	// Create go.mod file
	goModContent := `module github.com/guntisdev/entlite/examples/01-basic-entity

go 1.23.0
`
	goModPath := filepath.Join(tmpDir, "go.mod")
	if err := os.WriteFile(goModPath, []byte(goModContent), 0644); err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
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

	// Create dummy db/db.go file - just needs to exist for the import path check
	dbContent := `package db

import (
	"database/sql"
	"time"
)

type User struct {
	ID        int32
	Email     string
	Name      string
	Age       sql.NullInt32
	Uuid      string
	IsAdmin   bool
	CreatedAt time.Time
	UpdatedAt time.Time
}
`
	dbPath := filepath.Join(dbDir, "db.go")
	if err := os.WriteFile(dbPath, []byte(dbContent), 0644); err != nil {
		t.Fatalf("Failed to write db file: %v", err)
	}

	// Create dummy pb/pb.go file - just needs to exist for the import path check
	pbContent := `package pb

import (
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

type User struct {
	Id        int32
	Email     string
	Name      string
	Age       *int32
	Uuid      string
	IsAdmin   bool
	CreatedAt *timestamppb.Timestamp
	UpdatedAt *timestamppb.Timestamp
}
`
	pbPath := filepath.Join(pbDir, "pb.go")
	if err := os.WriteFile(pbPath, []byte(pbContent), 0644); err != nil {
		t.Fatalf("Failed to write pb file: %v", err)
	}

	// Change to tmpDir to simulate working in the project
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	if err := os.Chdir(filepath.Join(tmpDir, "ent")); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}
	defer os.Chdir(originalDir)

	// Run convertCommand
	convertCommand([]string{dbDir, pbDir})

	// Check that convert.go was created
	convertPath := filepath.Join(tmpDir, "ent", "gen", "convert", "convert.go")
	if _, err := os.Stat(convertPath); os.IsNotExist(err) {
		t.Fatalf("Expected convert.go was not created at %s", convertPath)
	}

	// Read the generated content
	actualContent, err := os.ReadFile(convertPath)
	if err != nil {
		t.Fatalf("Failed to read generated convert.go: %v", err)
	}

	// Expected content based on examples/01-basic-entity/ent/gen/convert/convert.go
	expectedContent := `// generate convertion between db and pb types
package convert

import (
	"database/sql"
	"time"
	"google.golang.org/protobuf/types/known/timestamppb"
	"github.com/guntisdev/entlite/examples/01-basic-entity/ent/gen/db"
	"github.com/guntisdev/entlite/examples/01-basic-entity/ent/gen/pb"
)

// +++++ User conversion functions

// UserDBToProto converts a database model to proto message
func UserDBToProto(db *db.User) *pb.User {
	if db == nil {
		return nil
	}

	return &pb.User{
		Id: db.ID,
		Email: db.Email,
		Name: db.Name,
		Age: NullInt32ToPtr(db.Age),
		Uuid: db.Uuid,
		IsAdmin: db.IsAdmin,
		CreatedAt: TimeToProto(db.CreatedAt),
		UpdatedAt: TimeToProto(db.UpdatedAt),
	}
}

// UserProtoToDB converts a proto message to database model
func UserProtoToDB(pb *pb.User) *db.User {
	if pb == nil {
		return nil
	}

	return &db.User{
		ID: pb.Id,
		Email: pb.Email,
		Name: pb.Name,
		Age: PtrToNullInt32(pb.Age),
		Uuid: pb.Uuid,
		IsAdmin: pb.IsAdmin,
		CreatedAt: ProtoToTime(pb.CreatedAt),
		UpdatedAt: ProtoToTime(pb.UpdatedAt),
	}
}
// UserDBSliceToProtoSlice converts db slice to proto array message
func UserDBSliceToProtoSlice(dbSlice []*db.User) []*pb.User {
	if dbSlice == nil {
		return nil
	}

	result := make([]*pb.User, len(dbSlice))
	for i, row := range dbSlice {
		result[i] = UserDBToProto(row)
	}
	return result
}



// ++++++ Helper functions for type conversions

// TimeToProto converts time.Time to timestamppb.Timestamp pointer
func TimeToProto(t time.Time) *timestamppb.Timestamp {
	return timestamppb.New(t)
}

// Note: If the pointer is nil, it returns a zero time.Time{}
func ProtoToTime(t *timestamppb.Timestamp) time.Time {
	if t == nil {
		return time.Time{}
	}
	return t.AsTime()
}

// --- Int32 Converters ---
func NullInt32ToPtr(n sql.NullInt32) *int32 {
	if !n.Valid { return nil }
	return &n.Int32
}

func PtrToNullInt32(i *int32) sql.NullInt32 {
	if i == nil {
		return sql.NullInt32{Valid: false}
	}
	return sql.NullInt32{ Int32: *i, Valid: true }
}

// --- String Converters ---
func NullStringToPtr(n sql.NullString) *string {
	if !n.Valid { return nil }
	return &n.String
}

func PtrToNullString(i *string) sql.NullString {
	if i == nil {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{ String: *i, Valid: true }
}

// --- Bool Converters ---
func NullBoolToPtr(n sql.NullBool) *bool {
	if !n.Valid { return nil }
	return &n.Bool
}

func PtrToNullBool(i *bool) sql.NullBool {
	if i == nil {
		return sql.NullBool{Valid: false}
	}
	return sql.NullBool{ Bool: *i, Valid: true }
}

// --- Time Converters ---
func NullTimeToProto(n sql.NullTime) *timestamppb.Timestamp {
	if !n.Valid {
		return nil
	}
	return timestamppb.New(n.Time)
}

func ProtoToNullTime(t *timestamppb.Timestamp) sql.NullTime {
	if t == nil {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{
		Time:  t.AsTime(),
		Valid: true,
	}
}
`

	// Compare using Diff
	if d := testutil.Diff(expectedContent, string(actualContent)); d != "" {
		t.Errorf("Convert.go content mismatch (-expected +actual):\n%s", d)
	}
}

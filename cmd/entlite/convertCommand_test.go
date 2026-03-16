package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/guntisdev/entlite/internal/util"
)

func TestConvertCommand(t *testing.T) {

	tmpDir := t.TempDir()

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

	// Write common test input files
	writeTestGoMod(t, tmpDir)
	writeTestUserSchema(t, schemaDir)
	writeTestLogic(t, logicDir)

	dbContent := `package db

import (
	"database/sql"
	"time"
)

type User struct {
	ID        int64
	Email     string
	Name      string
	Age       sql.NullInt64
	Score     float64
	Uuid      string
	IsAdmin   bool
	ApiKey    []byte
	CreatedAt time.Time
	UpdatedAt time.Time
}`
	dbPath := filepath.Join(dbDir, "db.go")
	if err := os.WriteFile(dbPath, []byte(dbContent), 0644); err != nil {
		t.Fatalf("Failed to write db file: %v", err)
	}

	pbContent := `package pb

import (
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

type User struct {
	Id        int64
	Email     string
	Name      string
	Age       *int64
	Score     float64
	Uuid      string
	IsAdmin   bool
	ApiKey    []byte
	CreatedAt *timestamppb.Timestamp
	UpdatedAt *timestamppb.Timestamp
}`
	pbPath := filepath.Join(pbDir, "pb.go")
	if err := os.WriteFile(pbPath, []byte(pbContent), 0644); err != nil {
		t.Fatalf("Failed to write pb file: %v", err)
	}

	// Change to tmpDir to simulate working in the project
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	// run convert command from t.TempDir()/ent/
	if err := os.Chdir(filepath.Join(tmpDir, "ent")); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}
	defer os.Chdir(originalDir)

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

	convertCommand([]string{dbDir, pbDir})

	convertPath := filepath.Join(tmpDir, "ent", "gen", "convert", "convert.go")
	if _, err := os.Stat(convertPath); os.IsNotExist(err) {
		t.Fatalf("Expected convert.go was not created at %s", convertPath)
	}

	actualContent, err := os.ReadFile(convertPath)
	if err != nil {
		t.Fatalf("Failed to read generated convert.go: %v", err)
	}

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
		Age: NullInt64ToPtr(db.Age),
		Score: db.Score,
		Uuid: db.Uuid,
		IsAdmin: db.IsAdmin,
		ApiKey: db.ApiKey,
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
		Age: PtrToNullInt64(pb.Age),
		Score: pb.Score,
		Uuid: pb.Uuid,
		IsAdmin: pb.IsAdmin,
		ApiKey: pb.ApiKey,
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

// --- Int64 Converters ---
func NullInt64ToPtr(n sql.NullInt64) *int64 {
	if !n.Valid { return nil }
	return &n.Int64
}

func PtrToNullInt64(i *int64) sql.NullInt64 {
	if i == nil {
		return sql.NullInt64{Valid: false}
	}
	return sql.NullInt64{ Int64: *i, Valid: true }
}

// --- Float64 Converters ---
func NullFloat64ToPtr(n sql.NullFloat64) *float64 {
	if !n.Valid { return nil }
	return &n.Float64
}

func PtrToNullFloat64(i *float64) sql.NullFloat64 {
	if i == nil {
		return sql.NullFloat64{Valid: false}
	}
	return sql.NullFloat64{ Float64: *i, Valid: true }
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

// --- Bytes Converters ---
func NullBytesToPtr(b []byte) *[]byte {
    if b == nil { return nil }
    return &b
}

func PtrToNullBytes(b *[]byte) []byte {
    if b == nil { return nil }
    return *b
}

// --- SQLite bool converters ---
func SQLiteIntToBool(i int64) bool {
    switch i {
    case 0:
        return false
    case 1:
        return true
    default:
        panic("Unable convert sqlite int to bool")
    }
}

func SQLiteBoolToInt(b bool) int64 {
    if b {
        return 1
    } else {
        return 0
    }
}`

	if d := util.Diff(expectedContent, string(actualContent)); d != "" {
		t.Errorf("Convert.go content mismatch (-expected +actual):\n%s", d)
	}
}

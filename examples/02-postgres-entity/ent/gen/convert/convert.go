// generate convertion between db and pb types
package convert

import (
	"database/sql"
	"time"
	"google.golang.org/protobuf/types/known/timestamppb"
	"github.com/guntisdev/entlite/examples/02-postgres-entity/ent/gen/db"
	"github.com/guntisdev/entlite/examples/02-postgres-entity/ent/gen/pb"
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
		Score: db.Score,
		Uuid: db.Uuid,
		IsAdmin: db.IsAdmin,
		ApiKey: db.ApiKey,
		LastLoginMs: db.LastLoginMs,
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
		Score: pb.Score,
		Uuid: pb.Uuid,
		IsAdmin: pb.IsAdmin,
		ApiKey: pb.ApiKey,
		LastLoginMs: pb.LastLoginMs,
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
}

// generate convertion between db and pb types
package convert

import (
	"database/sql"
	"time"
	"google.golang.org/protobuf/types/known/timestamppb"
	"github.com/guntisdev/entlite/examples/02-multiple-entities/ent/gen/db"
	"github.com/guntisdev/entlite/examples/02-multiple-entities/ent/gen/pb"
)

// +++++ Post conversion functions

// PostDBToProto converts a database model to proto message
func PostDBToProto(db *db.Post) *pb.Post {
	if db == nil {
		return nil
	}

	return &pb.Post{
		Id: db.ID,
		Title: db.Title,
		Content: db.Content,
		Published: db.Published,
	}
}

// PostProtoToDB converts a proto message to database model
func PostProtoToDB(pb *pb.Post) *db.Post {
	if pb == nil {
		return nil
	}

	return &db.Post{
		ID: pb.Id,
		Title: pb.Title,
		Content: pb.Content,
		Published: pb.Published,
	}
}
// PostDBSliceToProtoSlice converts db slice to proto array message
func PostDBSliceToProtoSlice(dbSlice []*db.Post) []*pb.Post {
	if dbSlice == nil {
		return nil
	}

	result := make([]*pb.Post, len(dbSlice))
	for i, row := range dbSlice {
		result[i] = PostDBToProto(row)
	}
	return result
}

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

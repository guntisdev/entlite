package db

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"google.golang.org/protobuf/types/known/timestamppb"
	"github.com/guntisdev/entlite/examples/01-sqlite-entity/ent/logic"
	"time"
	internal "github.com/guntisdev/entlite/examples/01-sqlite-entity/ent/gen/db/internal"
)

type CreateUserParams struct {
	Email string `json:"email"`
	Name string `json:"name"`
	Age *int32 `json:"age"`
	Password string `json:"password"`
	Score float64 `json:"score"`
	Uuid *string `json:"uuid"`
	IsAdmin bool `json:"is_admin"`
	ApiKey *[]byte `json:"api_key"`
	LastLoginMs int64 `json:"last_login_ms"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (int32, error) {
	if !logic.StartsWithCapital(arg.Name) {
		return 0, fmt.Errorf("Failed create: incorrect value for 'User' in field 'name', validated by 'logic.StartsWithCapital'")
	}
	internalArg := internal.CreateUserParams{
		Email: arg.Email,
		Name: arg.Name,
		Age: SQLitePtrInt32ToNullInt64(arg.Age),
		Password: arg.Password,
		Score: arg.Score,
		Uuid: OptionalWithFallback(arg.Uuid, logic.GetUuidStr()),
		IsAdmin: SQLiteBoolToInt(arg.IsAdmin),
		ApiKey: OptionalWithFallback(arg.ApiKey, logic.GenerateAPIKey()),
		LastLoginMs: arg.LastLoginMs,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	id, err := (*internal.Queries)(q).CreateUser(ctx, internalArg)
	return SQLiteInt64ToInt32(id), err
}

func (q *Queries) DeleteUser(ctx context.Context, id int32) error {
	return (*internal.Queries)(q).DeleteUser(ctx, SQLiteInt32ToInt64(id))
}

func (q *Queries) GetUser(ctx context.Context, id int32) (*User, error) {
	dbResult, err := (*internal.Queries)(q).GetUser(ctx, SQLiteInt32ToInt64(id))
	if err != nil {
		return nil, err
	}
	return UserFromSQL(&dbResult), nil
}

func (q *Queries) ListUser(ctx context.Context) ([]*User, error) {
	dbResults, err := (*internal.Queries)(q).ListUser(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]*User, len(dbResults))
	for i := range dbResults {
		result[i] = UserFromSQL(&dbResults[i])
	}
	return result, nil
}

type UpdateUserParams struct {
	Email string `json:"email"`
	Name string `json:"name"`
	Age *int32 `json:"age"`
	Password *string `json:"password"`
	Score float64 `json:"score"`
	IsAdmin bool `json:"is_admin"`
	ApiKey *[]byte `json:"api_key"`
	LastLoginMs int64 `json:"last_login_ms"`
	UpdatedAt *time.Time `json:"updated_at"`
	ID int32 `json:"id"`
}

func (q *Queries) UpdateUser(ctx context.Context, arg UpdateUserParams) (*User, error) {
	if !logic.StartsWithCapital(arg.Name) {
		return nil, fmt.Errorf("Failed update: incorrect value for 'User' in field 'name', validated by 'logic.StartsWithCapital'")
	}
	internalArg := internal.UpdateUserParams{
		ID: SQLiteInt32ToInt64(arg.ID),
		Email: arg.Email,
		Name: arg.Name,
		Age: SQLitePtrInt32ToNullInt64(arg.Age),
		Password: PtrToNullString(arg.Password),
		Score: arg.Score,
		IsAdmin: SQLiteBoolToInt(arg.IsAdmin),
		ApiKey: OptionalWithFallback(arg.ApiKey, logic.GenerateAPIKey()),
		LastLoginMs: arg.LastLoginMs,
		UpdatedAt: time.Now(),
	}

	dbUser, err := (*internal.Queries)(q).UpdateUser(ctx, internalArg)
	if err != nil {
		return nil, err
	}
	return UserFromSQL(&dbUser), nil
}


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

// OptionalWithFallback chooses fallback if optional value is nil
func OptionalWithFallback[T any](val *T, fallback T) T {
    if val != nil {
        return *val
    }
    return fallback
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

// --- SQLite int converters int32 - int64 ---
func SQLiteInt64ToInt32(n int64) int32 {
    if n < math.MinInt32 || n > math.MaxInt32 {
		panic("Unable convert sqlite int64 to int32")
	}
	return int32(n)
}

func SQLiteInt32ToInt64(n int32) int64 {
    return int64(n)
}

// --- SQLite null-int converters int32 - int64 ---
func SQLiteNullInt64ToPtrInt32(n sql.NullInt64) *int32 {
	if !n.Valid { return nil }
    v := SQLiteInt64ToInt32(n.Int64)
	return &v
}

func SQLitePtrInt32ToNullInt64(i *int32) sql.NullInt64 {
    if i == nil {
		return sql.NullInt64{Valid: false}
	}
	return sql.NullInt64{ Int64: int64(*i), Valid: true }
}

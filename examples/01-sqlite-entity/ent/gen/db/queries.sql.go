package db

import (
	"context"
	"database/sql"
	"fmt"
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
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (int32, error) {
	if !logic.StartsWithCapital(arg.Name) {
		return 0, fmt.Errorf("Failed create: incorrect value for 'User' in field 'name', validated by 'logic.StartsWithCapital'")
	}
	internalArg := internal.CreateUserParams{
		Email: arg.Email,
		Name: arg.Name,
		Age: IntPtrConvert[int32, int64](arg.Age),
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
	return IntConvert[int64, int32](id), err
}

func (q *Queries) DeleteUser(ctx context.Context, id int32) error {
	return (*internal.Queries)(q).DeleteUser(ctx, IntConvert[int32, int64](id))
}

func (q *Queries) GetUser(ctx context.Context, id int32) (*User, error) {
	dbResult, err := (*internal.Queries)(q).GetUser(ctx, IntConvert[int32, int64](id))
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
	ID int32 `json:"id"`
}

func (q *Queries) UpdateUser(ctx context.Context, arg UpdateUserParams) (*User, error) {
	if !logic.StartsWithCapital(arg.Name) {
		return nil, fmt.Errorf("Failed update: incorrect value for 'User' in field 'name', validated by 'logic.StartsWithCapital'")
	}
	internalArg := internal.UpdateUserParams{
		ID: IntConvert[int32, int64](arg.ID),
		Email: arg.Email,
		Name: arg.Name,
		Age: IntPtrConvert[int32, int64](arg.Age),
		Password: arg.Password,
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

// example: IntPtrConvert[int64, int32](dbRow.Age)
func IntPtrConvert[From, To ~int | ~int32 | ~int64 | ~float32 | ~float64](src *From) *To {
	if src == nil {
		return nil
	}
	val := To(*src)
	return &val
}
	
func IntConvert[From, To ~int | ~int8 | ~int16 | ~int32 | ~int64 | 
    ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | 
    ~float32 | ~float64](src From) To {
    return To(src)
}
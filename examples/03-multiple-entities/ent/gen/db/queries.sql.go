package db

import (
	"context"
	"database/sql"
	"math"
	internal "github.com/guntisdev/entlite/examples/03-multiple-entities/ent/gen/db/internal"
)

type CreatePostParams struct {
	Title string `json:"title"`
	Content string `json:"content"`
	Published bool `json:"published"`
}

func (q *Queries) CreatePost(ctx context.Context, arg CreatePostParams) (int32, error) {
	internalArg := internal.CreatePostParams{
		Title: arg.Title,
		Content: arg.Content,
		Published: arg.Published,
	}
	return (*internal.Queries)(q).CreatePost(ctx, internalArg)
}

type CreateUserParams struct {
	Email string `json:"email"`
	Name string `json:"name"`
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (int32, error) {
	internalArg := internal.CreateUserParams{
		Email: arg.Email,
		Name: arg.Name,
	}
	return (*internal.Queries)(q).CreateUser(ctx, internalArg)
}

func (q *Queries) DeletePost(ctx context.Context, id int32) error {
	return (*internal.Queries)(q).DeletePost(ctx, id)
}

func (q *Queries) DeleteUser(ctx context.Context, id int32) error {
	return (*internal.Queries)(q).DeleteUser(ctx, id)
}

func (q *Queries) GetPost(ctx context.Context, id int32) (*Post, error) {
	dbResult, err := (*internal.Queries)(q).GetPost(ctx, id)
	if err != nil {
		return nil, err
	}
	return PostFromSQL(&dbResult), nil
}

func (q *Queries) GetUser(ctx context.Context, id int32) (*User, error) {
	dbResult, err := (*internal.Queries)(q).GetUser(ctx, id)
	if err != nil {
		return nil, err
	}
	return UserFromSQL(&dbResult), nil
}

func (q *Queries) ListPost(ctx context.Context) ([]*Post, error) {
	dbResults, err := (*internal.Queries)(q).ListPost(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]*Post, len(dbResults))
	for i := range dbResults {
		result[i] = PostFromSQL(&dbResults[i])
	}
	return result, nil
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

type UpdatePostParams struct {
	Title string `json:"title"`
	Content string `json:"content"`
	Published bool `json:"published"`
	ID int32 `json:"id"`
}

func (q *Queries) UpdatePost(ctx context.Context, arg UpdatePostParams) (*Post, error) {
	internalArg := internal.UpdatePostParams{
		Title: arg.Title,
		Content: arg.Content,
		Published: arg.Published,
	}

	dbPost, err := (*internal.Queries)(q).UpdatePost(ctx, internalArg)
	if err != nil {
		return nil, err
	}
	return PostFromSQL(&dbPost), nil
}

type UpdateUserParams struct {
	Email string `json:"email"`
	Name string `json:"name"`
	ID int32 `json:"id"`
}

func (q *Queries) UpdateUser(ctx context.Context, arg UpdateUserParams) (*User, error) {
	internalArg := internal.UpdateUserParams{
		Email: arg.Email,
		Name: arg.Name,
	}

	dbUser, err := (*internal.Queries)(q).UpdateUser(ctx, internalArg)
	if err != nil {
		return nil, err
	}
	return UserFromSQL(&dbUser), nil
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

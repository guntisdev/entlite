package db

import (
	"context"
	"fmt"
	"github.com/guntisdev/entlite/examples/03-mysql-entity/ent/logic"
	"time"
	internal "github.com/guntisdev/entlite/examples/03-mysql-entity/ent/gen/db/internal"
)

type CreateUserParams struct {
	Email string `json:"email"`
	Name string `json:"name"`
	Age *int32 `json:"age"`
	Password string `json:"password"`
	Score *float64 `json:"score"`
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
		Age: PtrToNullInt32(arg.Age),
		Password: arg.Password,
		Score: OptionalWithFallback(arg.Score, 4.2),
		Uuid: OptionalWithFallback(arg.Uuid, logic.GetUuidStr()),
		IsAdmin: arg.IsAdmin,
		ApiKey: OptionalWithFallback(arg.ApiKey, logic.GenerateAPIKey()),
		LastLoginMs: arg.LastLoginMs,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	id, err := (*internal.Queries)(q).CreateUser(ctx, internalArg)
	return IntConvert[int64, int32](id), err
}

func (q *Queries) DeleteUser(ctx context.Context, id int32) error {
	return (*internal.Queries)(q).DeleteUser(ctx, id)
}

func (q *Queries) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	dbResult, err := (*internal.Queries)(q).GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return UserFromSQL(&dbResult), nil
}

func (q *Queries) GetUserByID(ctx context.Context, id int32) (*User, error) {
	dbResult, err := (*internal.Queries)(q).GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return UserFromSQL(&dbResult), nil
}

type GetUserByNameAgeParams = internal.GetUserByNameAgeParams
func (q *Queries) GetUserByNameAge(ctx context.Context, arg GetUserByNameAgeParams) (*User, error) {
	dbResult, err := (*internal.Queries)(q).GetUserByNameAge(ctx, arg)
	if err != nil {
		return nil, err
	}
	return UserFromSQL(&dbResult), nil
}

func (q *Queries) ListUserByName(ctx context.Context, name string) ([]*User, error) {
	dbResults, err := (*internal.Queries)(q).ListUserByName(ctx, name)
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
	Score *float64 `json:"score"`
	IsAdmin bool `json:"is_admin"`
	ApiKey *[]byte `json:"api_key"`
	LastLoginMs int64 `json:"last_login_ms"`
	ID int32 `json:"ID"`
}

func (q *Queries) UpdateUser(ctx context.Context, arg UpdateUserParams) (*User, error) {
	if !logic.StartsWithCapital(arg.Name) {
		return nil, fmt.Errorf("Failed update: incorrect value for 'User' in field 'name', validated by 'logic.StartsWithCapital'")
	}
	internalArg := internal.UpdateUserParams{
		ID: arg.ID,
		Email: arg.Email,
		Name: arg.Name,
		Age: PtrToNullInt32(arg.Age),
		Password: PtrToNullString(arg.Password),
		Score: PtrToNullFloat64(arg.Score),
		IsAdmin: arg.IsAdmin,
		ApiKey: PtrBytesToNullString(arg.ApiKey),
		LastLoginMs: arg.LastLoginMs,
		UpdatedAt: time.Now(),
	}

	err := (*internal.Queries)(q).UpdateUser(ctx, internalArg)
	if err != nil {
		return nil, err
	}
	dbUser, err := (*internal.Queries)(q).GetUserByID(ctx, arg.ID)
	if err != nil {
		return nil, err
	}
	return UserFromSQL(&dbUser), nil
}


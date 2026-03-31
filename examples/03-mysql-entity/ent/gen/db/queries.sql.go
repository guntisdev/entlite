package db

import (
	"context"
	"fmt"
	"github.com/guntisdev/entlite/examples/01-sqlite-entity/ent/logic"
	"time"
	internal "github.com/guntisdev/entlite/examples/03-mysql-entity/ent/gen/db/internal"
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
		Age: PtrToNullInt32(arg.Age),
		Password: arg.Password,
		Score: arg.Score,
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

func (q *Queries) GetUser(ctx context.Context, id int32) (*User, error) {
	dbResult, err := (*internal.Queries)(q).GetUser(ctx, id)
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
		ID: arg.ID,
		Email: arg.Email,
		Name: arg.Name,
		Age: PtrToNullInt32(arg.Age),
		Password: PtrToNullString(arg.Password),
		Score: arg.Score,
		IsAdmin: arg.IsAdmin,
		ApiKey: OptionalWithFallback(arg.ApiKey, logic.GenerateAPIKey()),
		LastLoginMs: arg.LastLoginMs,
		UpdatedAt: time.Now(),
	}

	err := (*internal.Queries)(q).UpdateUser(ctx, internalArg)
	if err != nil {
		return nil, err
	}
	dbUser, err := (*internal.Queries)(q).GetUser(ctx, arg.ID)
	if err != nil {
		return nil, err
	}
	return UserFromSQL(&dbUser), nil
}


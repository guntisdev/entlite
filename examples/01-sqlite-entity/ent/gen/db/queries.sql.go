package db

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/guntisdev/entlite/examples/01-sqlite-entity/ent/logic"
	"time"
	internal "github.com/guntisdev/entlite/examples/01-sqlite-entity/ent/gen/db/internal"
)

type CreateUserParams struct {
	Email string `json:"email"`
	Name string `json:"name"`
	Age *int32 `json:"age"`
	Score float64 `json:"score"`
	IsAdmin bool `json:"is_admin"`
	LastLoginMs int64 `json:"last_login_ms"`
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (int64, error) {
	if !logic.StartsWithCapital(arg.Name) {
		return 0, fmt.Errorf("Failed create: incorrect value for 'User' in field 'name', validated by 'logic.StartsWithCapital'")
	}
	internalArg := internal.CreateUserParams{
		Email: arg.Email,
		Name: arg.Name,
		Age: arg.Age,
		Score: arg.Score,
		Uuid: logic.GetUuidStr(),
		IsAdmin: arg.IsAdmin,
		ApiKey: logic.GenerateAPIKey(),
		LastLoginMs: arg.LastLoginMs,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	return (*internal.Queries)(q).CreateUser(ctx, internalArg)
}

type UpdateUserParams struct {
	Email string `json:"email"`
	Name string `json:"name"`
	Age *int32 `json:"age"`
	Score float64 `json:"score"`
	IsAdmin bool `json:"is_admin"`
	LastLoginMs int64 `json:"last_login_ms"`
	ID int32 `json:"id"`
}

func (q *Queries) UpdateUser(ctx context.Context, arg UpdateUserParams) (User, error) {
	if !logic.StartsWithCapital(arg.Name) {
		return User{}, fmt.Errorf("Failed update: incorrect value for 'User' in field 'name', validated by 'logic.StartsWithCapital'")
	}
	internalArg := internal.UpdateUserParams{
		Email: arg.Email,
		Name: arg.Name,
		Age: arg.Age,
		Score: arg.Score,
		IsAdmin: arg.IsAdmin,
		LastLoginMs: arg.LastLoginMs,
		UpdatedAt: time.Now(),
	}
	return (*internal.Queries)(q).UpdateUser(ctx, internalArg)
}


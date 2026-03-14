package db

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/guntisdev/entlite/examples/01-basic-entity/ent/logic"
	"time"
	internal "github.com/guntisdev/entlite/examples/01-basic-entity/ent/gen/db/internal"
)

type CreateUserParams struct {
	Email string `json:"email"`
	Name string `json:"name"`
	Age sql.NullInt64 `json:"age"`
	IsAdmin bool `json:"is_admin"`
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (int64, error) {
	if !logic.StartsWithCapital(arg.Name) {
		return 0, fmt.Errorf("Failed create: incorrect value for 'User' in field 'name', validated by 'logic.StartsWithCapital'")
	}
	internalArg := internal.CreateUserParams{
		Email: arg.Email,
		Name: arg.Name,
		Age: arg.Age,
		Uuid: logic.GetUuidStr(),
		IsAdmin: arg.IsAdmin,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	return (*internal.Queries)(q).CreateUser(ctx, internalArg)
}

type UpdateUserParams struct {
	Email string `json:"email"`
	Name string `json:"name"`
	Age sql.NullInt64 `json:"age"`
	IsAdmin bool `json:"is_admin"`
	ID int64 `json:"id"`
}

func (q *Queries) UpdateUser(ctx context.Context, arg UpdateUserParams) (User, error) {
	if !logic.StartsWithCapital(arg.Name) {
		return User{}, fmt.Errorf("Failed update: incorrect value for 'User' in field 'name', validated by 'logic.StartsWithCapital'")
	}
	internalArg := internal.UpdateUserParams{
		Email: arg.Email,
		Name: arg.Name,
		Age: arg.Age,
		IsAdmin: arg.IsAdmin,
		UpdatedAt: time.Now(),
	}
	return (*internal.Queries)(q).UpdateUser(ctx, internalArg)
}


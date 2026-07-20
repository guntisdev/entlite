package db

import (
	"context"
	"fmt"
	"github.com/guntisdev/entlite/examples/01-sqlite-entity/ent/logic"
	"time"
	internal "github.com/guntisdev/entlite/examples/01-sqlite-entity/ent/gen/db/internal"
)

type CreateBulkUserParams struct {
	Email string `json:"email"`
	Name string `json:"name"`
	Age *int32 `json:"age"`
	Password string `json:"password"`
	ApiKey *[]byte `json:"api_key"`
	IsActive *bool `json:"is_active"`
	LoginCount *int64 `json:"login_count"`
	Rating *float64 `json:"rating"`
}

func (q *Queries) CreateBulkUser(ctx context.Context, args []CreateBulkUserParams) ([]int32, error) {
	results := make([]int32, 0, len(args))
	for _, item := range args {
		if !logic.StartsWithCapital(item.Name) {
			return nil, fmt.Errorf("Failed create_bulk: incorrect value for 'User' in field 'name', validated by 'logic.StartsWithCapital'")
		}
		internalArg := internal.CreateBulkUserParams{
			Email: item.Email,
			Name: item.Name,
			Age: IntPtrConvert[int32, int64](item.Age),
			Password: item.Password,
			ApiKey: OptionalWithFallback(item.ApiKey, logic.GenerateAPIKey()),
			IsActive: SQLiteBoolToInt(OptionalWithFallback(item.IsActive, true)),
			LoginCount: OptionalWithFallback(item.LoginCount, 0),
			Rating: OptionalWithFallback(item.Rating, 0),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		id, err := (*internal.Queries)(q).CreateBulkUser(ctx, internalArg)
		if err != nil {
			return nil, err
		}
		results = append(results, IntConvert[int64, int32](id))
	}
	return results, nil
}

type CreateUserParams struct {
	Email string `json:"email"`
	Name string `json:"name"`
	Age *int32 `json:"age"`
	Password string `json:"password"`
	ApiKey *[]byte `json:"api_key"`
	IsActive *bool `json:"is_active"`
	LoginCount *int64 `json:"login_count"`
	Rating *float64 `json:"rating"`
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
		ApiKey: OptionalWithFallback(arg.ApiKey, logic.GenerateAPIKey()),
		IsActive: SQLiteBoolToInt(OptionalWithFallback(arg.IsActive, true)),
		LoginCount: OptionalWithFallback(arg.LoginCount, 0),
		Rating: OptionalWithFallback(arg.Rating, 0),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	id, err := (*internal.Queries)(q).CreateUser(ctx, internalArg)
	return IntConvert[int64, int32](id), err
}

func (q *Queries) DeleteAllUser(ctx context.Context) error {
	return (*internal.Queries)(q).DeleteAllUser(ctx)
}

func (q *Queries) DeleteUser(ctx context.Context, id int32) error {
	return (*internal.Queries)(q).DeleteUser(ctx, IntConvert[int32, int64](id))
}

func (q *Queries) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	dbResult, err := (*internal.Queries)(q).GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return UserFromSQL(&dbResult), nil
}

func (q *Queries) GetUserByID(ctx context.Context, id int32) (*User, error) {
	dbResult, err := (*internal.Queries)(q).GetUserByID(ctx, IntConvert[int32, int64](id))
	if err != nil {
		return nil, err
	}
	return UserFromSQL(&dbResult), nil
}

func (q *Queries) ListAllUser(ctx context.Context) ([]*User, error) {
	dbResults, err := (*internal.Queries)(q).ListAllUser(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]*User, len(dbResults))
	for i := range dbResults {
		result[i] = UserFromSQL(&dbResults[i])
	}
	return result, nil
}

func (q *Queries) ListUserByIsActive(ctx context.Context, isActive int64) ([]*User, error) {
	dbResults, err := (*internal.Queries)(q).ListUserByIsActive(ctx, isActive)
	if err != nil {
		return nil, err
	}
	result := make([]*User, len(dbResults))
	for i := range dbResults {
		result[i] = UserFromSQL(&dbResults[i])
	}
	return result, nil
}

type ListUserFilterByAgeNameParams = internal.ListUserFilterByAgeNameParams
func (q *Queries) ListUserFilterByAgeName(ctx context.Context, arg ListUserFilterByAgeNameParams) ([]*User, error) {
	dbResults, err := (*internal.Queries)(q).ListUserFilterByAgeName(ctx, arg)
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
	IsActive *bool `json:"is_active"`
	LoginCount *int64 `json:"login_count"`
	Rating *float64 `json:"rating"`
	ID int32 `json:"ID"`
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
		IsActive: SQLiteBoolPtrToInt64Ptr(arg.IsActive),
		LoginCount: arg.LoginCount,
		Rating: arg.Rating,
		UpdatedAt: time.Now(),
	}

	dbUser, err := (*internal.Queries)(q).UpdateUser(ctx, internalArg)
	if err != nil {
		return nil, err
	}
	return UserFromSQL(&dbUser), nil
}


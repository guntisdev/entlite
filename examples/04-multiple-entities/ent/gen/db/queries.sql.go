package db

import (
	"context"
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
		ID: arg.ID,
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
		ID: arg.ID,
		Email: arg.Email,
		Name: arg.Name,
	}

	dbUser, err := (*internal.Queries)(q).UpdateUser(ctx, internalArg)
	if err != nil {
		return nil, err
	}
	return UserFromSQL(&dbUser), nil
}


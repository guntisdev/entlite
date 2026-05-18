package db

import (
	"context"
	internal "github.com/guntisdev/entlite/examples/04-multiple-entities/ent/gen/db/internal"
)

func (q *Queries) GetPostByID(ctx context.Context, id int32) (*Post, error) {
	dbResult, err := (*internal.Queries)(q).GetPostByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return PostFromSQL(&dbResult), nil
}


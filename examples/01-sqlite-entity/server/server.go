package server

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/guntisdev/entlite/examples/01-sqlite-entity/ent/gen/convert"
	"github.com/guntisdev/entlite/examples/01-sqlite-entity/ent/gen/db"
	"github.com/guntisdev/entlite/examples/01-sqlite-entity/ent/gen/pb"
)

type UserServer struct {
	db *sql.DB
}

func NewUserServiceServer(db *sql.DB) *UserServer {
	return &UserServer{
		db: db,
	}
}

func (s *UserServer) Create(
	ctx context.Context,
	req *connect.Request[pb.CreateUserRequest],
) (*connect.Response[pb.User], error) {
	log.Printf("Create user: %+v", req.Msg)

	queries := db.New(s.db)
	wrappedQueries := (*db.Queries)(queries)

	userID, err := wrappedQueries.CreateUser(ctx, db.CreateUserParams{
		Email:       req.Msg.Email,
		Name:        req.Msg.Name,
		Age:         convert.SQLitePtrInt32ToNullInt64(req.Msg.Age),
		IsAdmin:     convert.SQLiteBoolToInt(req.Msg.IsAdmin),
		LastLoginMs: req.Msg.LastLoginMs,
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to create user: %w", err))
	}

	dbUser, err := queries.GetUser(ctx, userID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get created user: %w", err))
	}

	user := convert.UserDBToProto((*db.User)(&dbUser))
	return connect.NewResponse(user), nil
}

func (s *UserServer) Get(
	ctx context.Context,
	req *connect.Request[pb.GetUserRequest],
) (*connect.Response[pb.User], error) {
	log.Printf("Get user: id=%d", req.Msg.Id)

	queries := db.New(s.db)

	dbUser, err := queries.GetUser(ctx, convert.SQLiteInt32ToInt64(req.Msg.Id))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("user not found"))
		}
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get user: %w", err))
	}

	user := convert.UserDBToProto((*db.User)(&dbUser))
	return connect.NewResponse(user), nil
}

func (s *UserServer) Update(
	ctx context.Context,
	req *connect.Request[pb.UpdateUserRequest],
) (*connect.Response[pb.User], error) {
	log.Printf("Update user: id=%d, %+v", req.Msg.Id, req.Msg)

	queries := db.New(s.db)
	wrappedQueries := (*db.Queries)(queries)

	dbUser, err := wrappedQueries.UpdateUser(ctx, db.UpdateUserParams{
		ID:          convert.SQLiteInt32ToInt64(req.Msg.Id),
		Email:       req.Msg.Email,
		Name:        req.Msg.Name,
		Age:         convert.SQLitePtrInt32ToNullInt64(req.Msg.Age),
		IsAdmin:     convert.SQLiteBoolToInt(req.Msg.IsAdmin),
		LastLoginMs: req.Msg.LastLoginMs,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("user not found"))
		}
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to update user: %w", err))
	}

	user := convert.UserDBToProto((*db.User)(&dbUser))
	return connect.NewResponse(user), nil
}

func (s *UserServer) Delete(
	ctx context.Context,
	req *connect.Request[pb.DeleteUserRequest],
) (*connect.Response[emptypb.Empty], error) {
	log.Printf("Delete user: id=%d", req.Msg.Id)

	queries := db.New(s.db)

	err := queries.DeleteUser(ctx, convert.SQLiteInt32ToInt64(req.Msg.Id))
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to delete user: %w", err))
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (s *UserServer) List(
	ctx context.Context,
	req *connect.Request[pb.ListUserRequest],
) (*connect.Response[pb.ListUserResponse], error) {
	log.Printf("List users: limit=%d, offset=%d", req.Msg.Limit, req.Msg.Offset)

	queries := db.New(s.db)

	dbUsers, err := queries.ListUser(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to list users: %w", err))
	}

	dbUserPtrs := make([]*db.User, len(dbUsers))
	for i := range dbUsers {
		dbUserPtrs[i] = (*db.User)(&dbUsers[i])
	}

	users := convert.UserDBSliceToProtoSlice(dbUserPtrs)

	response := &pb.ListUserResponse{
		Users: users,
	}

	return connect.NewResponse(response), nil
}

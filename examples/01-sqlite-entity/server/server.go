package server

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/emptypb"

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

	userID, err := queries.CreateUser(ctx, db.CreateUserParams{
		Email:       req.Msg.Email,
		Name:        req.Msg.Name,
		Age:         req.Msg.Age,
		IsAdmin:     req.Msg.IsAdmin,
		LastLoginMs: req.Msg.LastLoginMs,
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to create user: %w", err))
	}

	user, err := queries.GetUser(ctx, userID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get created user: %w", err))
	}

	return connect.NewResponse(user), nil
}

func (s *UserServer) Get(
	ctx context.Context,
	req *connect.Request[pb.GetUserRequest],
) (*connect.Response[pb.User], error) {
	log.Printf("Get user: id=%d", req.Msg.ID)

	queries := db.New(s.db)

	user, err := queries.GetUser(ctx, req.Msg.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("user not found"))
		}
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get user: %w", err))
	}

	return connect.NewResponse(user), nil
}

func (s *UserServer) Update(
	ctx context.Context,
	req *connect.Request[pb.UpdateUserRequest],
) (*connect.Response[pb.User], error) {
	log.Printf("Update user: id=%d, %+v", req.Msg.ID, req.Msg)

	queries := db.New(s.db)

	dbUser, err := queries.UpdateUser(ctx, db.UpdateUserParams{
		ID:          req.Msg.ID,
		Email:       req.Msg.Email,
		Name:        req.Msg.Name,
		Age:         req.Msg.Age,
		IsAdmin:     req.Msg.IsAdmin,
		LastLoginMs: req.Msg.LastLoginMs,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("user not found"))
		}
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to update user: %w", err))
	}

	// TODO actually db user should be already converted to go types (no sql types exposed)
	pbUser := db.UserDBToProto(&dbUser)
	return connect.NewResponse(pbUser), nil
}

func (s *UserServer) Delete(
	ctx context.Context,
	req *connect.Request[pb.DeleteUserRequest],
) (*connect.Response[emptypb.Empty], error) {
	log.Printf("Delete user: id=%d", req.Msg.ID)

	queries := db.New(s.db)

	err := queries.DeleteUser(ctx, req.Msg.ID)
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

	response := &pb.ListUserResponse{
		Users: dbUsers,
	}

	return connect.NewResponse(response), nil
}

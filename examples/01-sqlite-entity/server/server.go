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
		Password:    req.Msg.Password,
		Score:       req.Msg.Score,
		Uuid:        req.Msg.Uuid,
		IsAdmin:     req.Msg.IsAdmin,
		ApiKey:      &req.Msg.ApiKey,
		LastLoginMs: req.Msg.LastLoginMs,
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to create user: %w", err))
	}

	user, err := queries.GetUserByID(ctx, userID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get created user: %w", err))
	}

	return connect.NewResponse(user.ToProto()), nil
}

func (s *UserServer) GetByID(
	ctx context.Context,
	req *connect.Request[pb.GetUserByIDRequest],
) (*connect.Response[pb.User], error) {
	log.Printf("Get user: id=%d", req.Msg.ID)

	queries := db.New(s.db)

	user, err := queries.GetUserByID(ctx, req.Msg.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("user not found"))
		}
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get user: %w", err))
	}

	return connect.NewResponse(user.ToProto()), nil
}

func (s *UserServer) GetByEmail(
	ctx context.Context,
	req *connect.Request[pb.GetUserByEmailRequest],
) (*connect.Response[pb.User], error) {
	log.Printf("Get user by email: email=%s", req.Msg.Email)

	queries := db.New(s.db)

	user, err := queries.GetUserByEmail(ctx, req.Msg.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("user not found"))
		}
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get user by email: %w", err))
	}

	return connect.NewResponse(user.ToProto()), nil
}

func (s *UserServer) GetByNameAge(
	ctx context.Context,
	req *connect.Request[pb.GetUserByNameAgeRequest],
) (*connect.Response[pb.User], error) {
	log.Printf("Get user by name and age: name=%s, age=%d", req.Msg.Name, req.Msg.Age)

	queries := db.New(s.db)

	user, err := queries.GetUserByNameAge(ctx, db.GetUserByNameAgeParams{
		Name: req.Msg.Name,
		Age:  db.IntPtrConvert[int32, int64](&req.Msg.Age),
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("user not found"))
		}
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get user by name and age: %w", err))
	}

	return connect.NewResponse(user.ToProto()), nil
}

func (s *UserServer) Update(
	ctx context.Context,
	req *connect.Request[pb.UpdateUserRequest],
) (*connect.Response[pb.User], error) {
	log.Printf("Update user: ID=%d, %+v", req.Msg.ID, req.Msg)

	queries := db.New(s.db)

	dbUser, err := queries.UpdateUser(ctx, db.UpdateUserParams{
		ID:          req.Msg.ID,
		Email:       req.Msg.Email,
		Name:        req.Msg.Name,
		Age:         req.Msg.Age,
		Password:    req.Msg.Password,
		Score:       req.Msg.Score,
		IsAdmin:     req.Msg.IsAdmin,
		ApiKey:      &req.Msg.ApiKey,
		LastLoginMs: req.Msg.LastLoginMs,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("user not found"))
		}
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to update user: %w", err))
	}

	return connect.NewResponse(dbUser.ToProto()), nil
}

func (s *UserServer) Delete(
	ctx context.Context,
	req *connect.Request[pb.DeleteUserRequest],
) (*connect.Response[emptypb.Empty], error) {
	log.Printf("Delete user: ID=%d", req.Msg.ID)

	queries := db.New(s.db)

	err := queries.DeleteUser(ctx, req.Msg.ID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to delete user: %w", err))
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (s *UserServer) FilterByAgeNameIsAdmin(
	ctx context.Context,
	req *connect.Request[pb.ListUserFilterByAgeNameIsAdminRequest],
) (*connect.Response[pb.ListUserFilterByAgeNameIsAdminResponse], error) {
	log.Printf("Filter users: min_age=%d, max_age=%d, name=%s, is_admin=%t",
		req.Msg.GetMinAge(), req.Msg.GetMaxAge(), req.Msg.GetName(), req.Msg.GetIsAdmin())

	queries := db.New(s.db)

	minAge := req.Msg.GetMinAge()
	maxAge := req.Msg.GetMaxAge()
	isAdmin := req.Msg.GetIsAdmin()
	dbUsers, err := queries.ListUserFilterByAgeNameIsAdmin(ctx, db.ListUserFilterByAgeNameIsAdminParams{
		MinAge:  db.IntPtrConvert[int32, int64](&minAge),
		MaxAge:  db.IntPtrConvert[int32, int64](&maxAge),
		Name:    req.Msg.GetName(),
		IsAdmin: db.SQLiteBoolToInt(isAdmin),
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to list users: %w", err))
	}

	pbUsers := make([]*pb.User, len(dbUsers))
	for i, dbUser := range dbUsers {
		pbUsers[i] = dbUser.ToProto()
	}

	response := &pb.ListUserFilterByAgeNameIsAdminResponse{
		Users: pbUsers,
	}

	return connect.NewResponse(response), nil
}

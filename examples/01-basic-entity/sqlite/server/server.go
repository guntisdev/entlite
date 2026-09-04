package server

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/guntisdev/entlite/examples/01-basic-entity/sqlite/ent/gen/db"
	"github.com/guntisdev/entlite/examples/01-basic-entity/sqlite/ent/gen/pb"
)

type UserServer struct {
	db *sql.DB
}

// enforces implementation of proto methods
var _ pb.UserServiceHandler = (*UserServer)(nil)

func NewUserServiceServer(db *sql.DB) *UserServer {
	return &UserServer{
		db: db,
	}
}

func (s *UserServer) CreateUser(
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
		ApiKey:      &req.Msg.ApiKey,
		IsActive:    req.Msg.IsActive,
		LoginCount:  req.Msg.LoginCount,
		Rating:      req.Msg.Rating,
		Preferences: req.Msg.Preferences,
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

func (s *UserServer) CreateBulkUser(
	ctx context.Context,
	req *connect.Request[pb.CreateBulkUserRequest],
) (*connect.Response[pb.CreateBulkUserResponse], error) {
	log.Printf("Create bulk users: %d rows", len(req.Msg.Rows))

	queries := db.New(s.db)

	params := make([]db.CreateBulkUserParams, 0, len(req.Msg.Rows))
	for _, row := range req.Msg.Rows {
		params = append(params, db.CreateBulkUserParams{
			Email:       row.Email,
			Name:        row.Name,
			Age:         row.Age,
			Password:    row.Password,
			ApiKey:      &row.ApiKey,
			IsActive:    row.IsActive,
			LoginCount:  row.LoginCount,
			Rating:      row.Rating,
			Preferences: row.Preferences,
		})
	}

	userIDs, err := queries.CreateBulkUser(ctx, params)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to create users: %w", err))
	}

	users := make([]*pb.User, 0, len(userIDs))
	for _, userID := range userIDs {
		user, err := queries.GetUserByID(ctx, userID)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get created user: %w", err))
		}
		users = append(users, user.ToProto())
	}

	return connect.NewResponse(&pb.CreateBulkUserResponse{Rows: users}), nil
}

func (s *UserServer) GetUserByID(
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

func (s *UserServer) GetUserByEmail(
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

func (s *UserServer) UpdateUser(
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
		IsActive:    req.Msg.IsActive,
		LoginCount:  req.Msg.LoginCount,
		Rating:      req.Msg.Rating,
		Preferences: req.Msg.Preferences,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("user not found"))
		}
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to update user: %w", err))
	}

	return connect.NewResponse(dbUser.ToProto()), nil
}

func (s *UserServer) DeleteUser(
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

func (s *UserServer) DeleteAllUser(
	ctx context.Context,
	req *connect.Request[pb.DeleteAllUserRequest],
) (*connect.Response[emptypb.Empty], error) {
	log.Printf("Delete all users")

	queries := db.New(s.db)

	err := queries.DeleteAllUser(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to delete all users: %w", err))
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (s *UserServer) ListAllUser(
	ctx context.Context,
	req *connect.Request[pb.ListAllUserRequest],
) (*connect.Response[pb.ListAllUserResponse], error) {
	log.Printf("List all users")

	queries := db.New(s.db)

	dbUsers, err := queries.ListAllUser(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to list users: %w", err))
	}

	pbUsers := make([]*pb.User, len(dbUsers))
	for i, dbUser := range dbUsers {
		pbUsers[i] = dbUser.ToProto()
	}

	response := &pb.ListAllUserResponse{
		Rows: pbUsers,
	}

	return connect.NewResponse(response), nil
}

func (s *UserServer) ListActive(
	ctx context.Context,
	req *connect.Request[pb.ListActiveRequest],
) (*connect.Response[pb.ListActiveResponse], error) {
	log.Printf("List users by is_active: is_active=%t", req.Msg.GetIsActive())

	queries := db.New(s.db)

	dbUsers, err := queries.ListActive(ctx, db.ListActiveParams{
		IsActive: req.Msg.GetIsActive(),
		Limit:    req.Msg.GetLimit(),
		Offset:   req.Msg.GetOffset(),
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to list users: %w", err))
	}

	pbUsers := make([]*pb.User, len(dbUsers))
	for i, dbUser := range dbUsers {
		pbUsers[i] = dbUser.ToProto()
	}

	response := &pb.ListActiveResponse{
		Rows: pbUsers,
	}

	return connect.NewResponse(response), nil
}

func (s *UserServer) ListUserFilterByAgeName(
	ctx context.Context,
	req *connect.Request[pb.ListUserFilterByAgeNameRequest],
) (*connect.Response[pb.ListUserFilterByAgeNameResponse], error) {
	log.Printf("Filter users: min_age=%d, max_age=%d, name=%s",
		req.Msg.GetMinAge(), req.Msg.GetMaxAge(), req.Msg.GetName())

	queries := db.New(s.db)

	minAge := req.Msg.GetMinAge()
	maxAge := req.Msg.GetMaxAge()
	dbUsers, err := queries.ListUserFilterByAgeName(ctx, db.ListUserFilterByAgeNameParams{
		MinAge: &minAge,
		MaxAge: &maxAge,
		Name:   req.Msg.GetName(),
		Limit:  req.Msg.GetLimit(),
		Offset: req.Msg.GetOffset(),
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to list users: %w", err))
	}

	pbUsers := make([]*pb.User, len(dbUsers))
	for i, dbUser := range dbUsers {
		pbUsers[i] = dbUser.ToProto()
	}

	response := &pb.ListUserFilterByAgeNameResponse{
		Rows: pbUsers,
	}

	return connect.NewResponse(response), nil
}

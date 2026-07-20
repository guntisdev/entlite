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

// enforces implementation of proto methods
var _ pb.UserServiceHandler = (*UserServer)(nil)

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
		Email:      req.Msg.Email,
		Name:       req.Msg.Name,
		Age:        req.Msg.Age,
		Password:   req.Msg.Password,
		ApiKey:     &req.Msg.ApiKey,
		IsActive:   req.Msg.IsActive,
		LoginCount: req.Msg.LoginCount,
		Rating:     req.Msg.Rating,
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

func (s *UserServer) CreateBulk(
	ctx context.Context,
	req *connect.Request[pb.CreateBulkUserRequest],
) (*connect.Response[pb.CreateBulkUserResponse], error) {
	log.Printf("Create bulk users: %d items", len(req.Msg.Items))

	queries := db.New(s.db)

	params := make([]db.CreateBulkUserParams, 0, len(req.Msg.Items))
	for _, item := range req.Msg.Items {
		params = append(params, db.CreateBulkUserParams{
			Email:      item.Email,
			Name:       item.Name,
			Age:        item.Age,
			Password:   item.Password,
			ApiKey:     &item.ApiKey,
			IsActive:   item.IsActive,
			LoginCount: item.LoginCount,
			Rating:     item.Rating,
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

	return connect.NewResponse(&pb.CreateBulkUserResponse{Users: users}), nil
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

func (s *UserServer) Update(
	ctx context.Context,
	req *connect.Request[pb.UpdateUserRequest],
) (*connect.Response[pb.User], error) {
	log.Printf("Update user: ID=%d, %+v", req.Msg.ID, req.Msg)

	queries := db.New(s.db)

	dbUser, err := queries.UpdateUser(ctx, db.UpdateUserParams{
		ID:         req.Msg.ID,
		Email:      req.Msg.Email,
		Name:       req.Msg.Name,
		Age:        req.Msg.Age,
		Password:   req.Msg.Password,
		IsActive:   req.Msg.IsActive,
		LoginCount: req.Msg.LoginCount,
		Rating:     req.Msg.Rating,
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

func (s *UserServer) DeleteAll(
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

func (s *UserServer) ListAll(
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
		Users: pbUsers,
	}

	return connect.NewResponse(response), nil
}

func (s *UserServer) ListByIsActive(
	ctx context.Context,
	req *connect.Request[pb.ListUserByIsActiveRequest],
) (*connect.Response[pb.ListUserByIsActiveResponse], error) {
	log.Printf("List users by is_active: is_active=%t", req.Msg.GetIsActive())

	queries := db.New(s.db)

	dbUsers, err := queries.ListUserByIsActive(ctx, db.SQLiteBoolToInt(req.Msg.GetIsActive()))
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to list users: %w", err))
	}

	pbUsers := make([]*pb.User, len(dbUsers))
	for i, dbUser := range dbUsers {
		pbUsers[i] = dbUser.ToProto()
	}

	response := &pb.ListUserByIsActiveResponse{
		Users: pbUsers,
	}

	return connect.NewResponse(response), nil
}

func (s *UserServer) FilterByAgeName(
	ctx context.Context,
	req *connect.Request[pb.ListUserFilterByAgeNameRequest],
) (*connect.Response[pb.ListUserFilterByAgeNameResponse], error) {
	log.Printf("Filter users: min_age=%d, max_age=%d, name=%s",
		req.Msg.GetMinAge(), req.Msg.GetMaxAge(), req.Msg.GetName())

	queries := db.New(s.db)

	minAge := req.Msg.GetMinAge()
	maxAge := req.Msg.GetMaxAge()
	dbUsers, err := queries.ListUserFilterByAgeName(ctx, db.ListUserFilterByAgeNameParams{
		MinAge: db.IntPtrConvert[int32, int64](&minAge),
		MaxAge: db.IntPtrConvert[int32, int64](&maxAge),
		Name:   req.Msg.GetName(),
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to list users: %w", err))
	}

	pbUsers := make([]*pb.User, len(dbUsers))
	for i, dbUser := range dbUsers {
		pbUsers[i] = dbUser.ToProto()
	}

	response := &pb.ListUserFilterByAgeNameResponse{
		Users: pbUsers,
	}

	return connect.NewResponse(response), nil
}

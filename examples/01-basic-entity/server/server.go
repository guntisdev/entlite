package server

import (
	"context"
	"database/sql"
	"log"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/guntisdev/entlite/examples/01-basic-entity/ent/gen/pb"
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

	// TODO: Implement user creation logic
	user := &pb.User{
		Id:      1, // This should come from database
		Email:   req.Msg.Email,
		Name:    req.Msg.Name,
		Age:     req.Msg.Age,
		IsAdmin: req.Msg.IsAdmin,
	}

	return connect.NewResponse(user), nil
}

func (s *UserServer) Get(
	ctx context.Context,
	req *connect.Request[pb.GetUserRequest],
) (*connect.Response[pb.User], error) {
	log.Printf("Get user: id=%d", req.Msg.Id)

	// TODO: Implement user retrieval logic
	user := &pb.User{
		Id:      req.Msg.Id,
		Email:   "user@example.com",
		Name:    "Example User",
		IsAdmin: false,
	}

	return connect.NewResponse(user), nil
}

func (s *UserServer) Update(
	ctx context.Context,
	req *connect.Request[pb.UpdateUserRequest],
) (*connect.Response[pb.User], error) {
	log.Printf("Update user: id=%d, %+v", req.Msg.Id, req.Msg)

	// TODO: Implement user update logic
	user := &pb.User{
		Id:      req.Msg.Id,
		Email:   req.Msg.Email,
		Name:    req.Msg.Name,
		Age:     req.Msg.Age,
		IsAdmin: req.Msg.IsAdmin,
	}

	return connect.NewResponse(user), nil
}

func (s *UserServer) Delete(
	ctx context.Context,
	req *connect.Request[pb.DeleteUserRequest],
) (*connect.Response[emptypb.Empty], error) {
	log.Printf("Delete user: id=%d", req.Msg.Id)

	// TODO: Implement user deletion logic

	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (s *UserServer) List(
	ctx context.Context,
	req *connect.Request[pb.ListUserRequest],
) (*connect.Response[pb.ListUserResponse], error) {
	log.Printf("List users: limit=%d, offset=%d", req.Msg.Limit, req.Msg.Offset)

	// TODO: Implement user listing logic
	users := []*pb.User{
		{
			Id:      1,
			Email:   "user1@example.com",
			Name:    "User One",
			IsAdmin: false,
		},
		{
			Id:      2,
			Email:   "user2@example.com",
			Name:    "User Two",
			IsAdmin: true,
		},
	}

	response := &pb.ListUserResponse{
		Users: users,
	}

	return connect.NewResponse(response), nil
}

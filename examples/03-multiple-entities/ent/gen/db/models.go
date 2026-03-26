package db

import (
	"google.golang.org/protobuf/types/known/timestamppb"
	pb "github.com/guntisdev/entlite/examples/03-multiple-entities/ent/gen/pb"
	internal "github.com/guntisdev/entlite/examples/03-multiple-entities/ent/gen/db/internal"
)


type Post struct {
	ID int32 `json:"ID"`
	Title string `json:"title"`
	Content string `json:"content"`
	Published bool `json:"published"`
}

type User struct {
	ID int32 `json:"ID"`
	Email string `json:"email"`
	Name string `json:"name"`
}

func (m *Post) PostToSQL() *internal.Post {
	if m == nil {
		return nil
	}

	return &internal.Post{
		ID: m.ID,
		Title: m.Title,
		Content: m.Content,
		Published: m.Published,
	}
}

func PostFromSQL(db *internal.Post) *Post {
	if db == nil {
		return nil
	}

	return &Post{
		ID: db.ID,
		Title: db.Title,
		Content: db.Content,
		Published: db.Published,
	}
}

func (m *User) UserToSQL() *internal.User {
	if m == nil {
		return nil
	}

	return &internal.User{
		ID: m.ID,
		Email: m.Email,
		Name: m.Name,
	}
}

func UserFromSQL(db *internal.User) *User {
	if db == nil {
		return nil
	}

	return &User{
		ID: db.ID,
		Email: db.Email,
		Name: db.Name,
	}
}

// ToProto converts Post to proto format
func (m *Post) ToProto() *pb.Post {
	if m == nil {
		return nil
	}

	return &pb.Post{
		ID: m.ID,
		Title: m.Title,
		Content: m.Content,
		Published: m.Published,
	}
}

// ToProto converts User to proto format
func (m *User) ToProto() *pb.User {
	if m == nil {
		return nil
	}

	return &pb.User{
		ID: m.ID,
		Email: m.Email,
		Name: m.Name,
	}
}


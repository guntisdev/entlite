package db

import (
	"time"
	"google.golang.org/protobuf/types/known/timestamppb"
	pb "github.com/guntisdev/entlite/examples/03-mysql-entity/ent/gen/pb"
	internal "github.com/guntisdev/entlite/examples/03-mysql-entity/ent/gen/db/internal"
)


type User struct {
	ID int32 `json:"ID"`
	Email string `json:"email"`
	Name string `json:"name"`
	Age *int32 `json:"age"`
	Password string `json:"password"`
	Score float64 `json:"score"`
	Uuid string `json:"uuid"`
	IsAdmin bool `json:"is_admin"`
	ApiKey []byte `json:"api_key"`
	LastLoginMs int64 `json:"last_login_ms"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (m *User) UserToSQL() *internal.User {
	if m == nil {
		return nil
	}

	return &internal.User{
		ID: m.ID,
		Email: m.Email,
		Name: m.Name,
		Age: PtrToNullInt32(m.Age),
		Password: m.Password,
		Score: m.Score,
		Uuid: m.Uuid,
		IsAdmin: m.IsAdmin,
		ApiKey: m.ApiKey,
		LastLoginMs: m.LastLoginMs,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
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
		Age: NullInt32ToPtr(db.Age),
		Password: db.Password,
		Score: db.Score,
		Uuid: db.Uuid,
		IsAdmin: db.IsAdmin,
		ApiKey: db.ApiKey,
		LastLoginMs: db.LastLoginMs,
		CreatedAt: db.CreatedAt,
		UpdatedAt: db.UpdatedAt,
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
		Age: m.Age,
		Score: m.Score,
		Uuid: m.Uuid,
		IsAdmin: m.IsAdmin,
		ApiKey: m.ApiKey,
		LastLoginMs: m.LastLoginMs,
		CreatedAt: timestamppb.New(m.CreatedAt),
		UpdatedAt: timestamppb.New(m.UpdatedAt),
	}
}


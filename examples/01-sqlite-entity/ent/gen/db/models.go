package db

import (
	pb "github.com/guntisdev/entlite/examples/01-sqlite-entity/ent/gen/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
	internal "github.com/guntisdev/entlite/examples/01-sqlite-entity/ent/gen/db/internal"
)


type User struct {
	ID int32 `json:"ID"`
	Email string `json:"email"`
	Name string `json:"name"`
	Age *int32 `json:"age"`
	Password string `json:"password"`
	ApiKey []byte `json:"api_key"`
	IsActive bool `json:"is_active"`
	LoginCount int64 `json:"login_count"`
	Rating float64 `json:"rating"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (m *User) UserToSQL() *internal.User {
	if m == nil {
		return nil
	}

	return &internal.User{
		ID: IntConvert[int32, int64](m.ID),
		Email: m.Email,
		Name: m.Name,
		Age: IntPtrConvert[int32, int64](m.Age),
		Password: m.Password,
		ApiKey: m.ApiKey,
		IsActive: SQLiteBoolToInt(m.IsActive),
		LoginCount: m.LoginCount,
		Rating: m.Rating,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func UserFromSQL(db *internal.User) *User {
	if db == nil {
		return nil
	}

	return &User{
		ID: IntConvert[int64, int32](db.ID),
		Email: db.Email,
		Name: db.Name,
		Age: IntPtrConvert[int64, int32](db.Age),
		Password: db.Password,
		ApiKey: db.ApiKey,
		IsActive: SQLiteIntToBool(db.IsActive),
		LoginCount: db.LoginCount,
		Rating: db.Rating,
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
		ApiKey: m.ApiKey,
		IsActive: m.IsActive,
		LoginCount: m.LoginCount,
		Rating: m.Rating,
		CreatedAt: timestamppb.New(m.CreatedAt),
		UpdatedAt: timestamppb.New(m.UpdatedAt),
	}
}


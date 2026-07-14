package schema

import (
	"time"

	"github.com/guntisdev/entlite/examples/01-sqlite-entity/ent/logic"
	"github.com/guntisdev/entlite/pkg/entlite"
	"github.com/guntisdev/entlite/pkg/entlite/field"
	"github.com/guntisdev/entlite/pkg/entlite/filter"
	"github.com/guntisdev/entlite/pkg/entlite/permissions"
	"github.com/guntisdev/entlite/pkg/entlite/query"
)

type User struct {
	entlite.Schema
}

func (User) Annotations() []entlite.Annotation {
	return []entlite.Annotation{
		entlite.Message(),
		entlite.GRPC(),
	}
}

func (User) Fields() []entlite.Field {
	return []entlite.Field{
		field.String("email").Unique(),
		field.String("name").Comment("Full name, e.g. \"Jane Doe\"").Validate(logic.StartsWithCapital),
		field.Int("age").Optional(),
		field.String("password").Permissions(permissions.WriteOnly),
		field.Byte("api_key").Immutable().DefaultFunc(logic.GenerateAPIKey),
		field.Bool("is_active").Default(true),
		field.Int64("login_count").Default(0),
		field.Float("rating").Default(0),
		field.Time("created_at").DefaultFunc(time.Now).Immutable().Permissions(permissions.ReadOnly),
		field.Time("updated_at").DefaultFunc(time.Now).Permissions(permissions.ReadOnly),
	}
}

func (User) Queries() []entlite.Query {
	return []entlite.Query{
		// DefaultCRUD expands to Create / GetByID / Update / Delete / List.
		query.DefaultCRUD(),
		query.GetBy("email"),
		query.ListBy("is_active"),
		query.ListBy(
			filter.Range("age"),   // age BETWEEN :min_age AND :max_age
			filter.Search("name"), // name LIKE :name
		).OrderBy("created_at").Count(),
	}
}

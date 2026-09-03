package schema

import (
	"time"

	"github.com/guntisdev/entlite/examples/01-basic-entity/postgresql/ent/logic"
	"github.com/guntisdev/entlite/pkg/entlite"
	"github.com/guntisdev/entlite/pkg/entlite/field"
	"github.com/guntisdev/entlite/pkg/entlite/filter"
	"github.com/guntisdev/entlite/pkg/entlite/index"
	"github.com/guntisdev/entlite/pkg/entlite/query"
)

type User struct {
	entlite.Schema
}

func (User) Contracts() []entlite.Contract {
	return []entlite.Contract{
		entlite.SQLC(),
		entlite.PROTO(),
	}
}

func (User) Fields() []entlite.Field {
	return []entlite.Field{
		field.String("email").Unique(),
		// Full name, e.g. "Jane Doe"
		field.String("name").Validate(logic.StartsWithCapital),
		field.Int("age").Optional(),
		field.String("password").Contracts(entlite.SQLC(), entlite.PROTO().WriteOnly()),
		field.Byte("api_key").Immutable().DefaultFunc(logic.GenerateAPIKey),
		field.Bool("is_active").Default(true),
		field.Int64("login_count").Default(0),
		field.Float("rating").Default(0),
		// UI preferences, e.g. {"theme":"dark"}
		field.JSON("preferences").Default("{}"),
		field.Time("created_at").DefaultFunc(time.Now).Immutable().Contracts(entlite.SQLC(), entlite.PROTO().ReadOnly()),
		field.Time("updated_at").DefaultFunc(time.Now).Contracts(entlite.SQLC(), entlite.PROTO().ReadOnly()),
	}
}

func (User) Queries() []entlite.Query {
	return []entlite.Query{
		query.DefaultCRUD(),
		query.CreateBulk(),
		// Look up a user by email address
		query.GetBy("email"),
		query.ListAll(),
		query.DeleteAll(),
		query.ListBy("is_active").Name("ListActive"),
		query.ListBy(
			filter.Range("age"),   // age BETWEEN :min_age AND :max_age
			filter.Search("name"), // name LIKE :name
		).OrderBy("created_at").Count(),
	}
}

func (User) Indexes() []entlite.Index {
	return []entlite.Index{
		// compound primary key, it replaces the generated id column
		// index.Primary("email", "created_at"),
		// index on two columns
		index.Fields("age", "is_active"),
		// descending sort
		index.Fields("is_active").
			Desc("created_at"),
		// unique across two columns
		index.Fields("name", "email").Unique(),
		// explicit index name
		index.Fields("login_count", "rating").
			Name("idx_users_stats"),
	}
}

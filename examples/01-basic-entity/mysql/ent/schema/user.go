package schema

import (
	"time"

	"github.com/guntisdev/entlite/examples/01-basic-entity/mysql/ent/logic"
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
		// re-importing the same users overwrites the row that shares the email
		query.CreateBulk().Upsert("email"),
		// Look up a user by email address
		query.GetBy("email"),
		query.ListAll(),
		query.DeleteAll(),
		query.ListBy("is_active").Name("ListActive").Limit().Offset(),
		query.ListBy(
			filter.Range("age"),   // age BETWEEN :min_age AND :max_age
			filter.Search("name"), // name LIKE :name
		).Asc("created_at").Count().Limit().Offset(),
	}
}

func (User) Indexes() []entlite.Index {
	return []entlite.Index{
		// index on two columns
		index.Asc("age", "is_active"),
		// created_at is sorted the other way round
		index.Asc("is_active").
			Desc("created_at"),
		// newest first, id breaks the tie
		index.Desc("created_at").
			Asc("id"),
		// unique across two columns
		index.Asc("name", "email").Unique(),
		// explicit index name
		index.Asc("login_count", "rating").
			Name("idx_users_stats"),
	}
}

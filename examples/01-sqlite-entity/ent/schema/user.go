package schema

import (
	"time"

	"github.com/guntisdev/entlite/examples/01-sqlite-entity/ent/logic"
	"github.com/guntisdev/entlite/pkg/entlite"
	"github.com/guntisdev/entlite/pkg/entlite/field"
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
		field.String("email").Unique().ProtoField(2),
		field.String("name").Validate(logic.StartsWithCapital).Comment("First name and surname"),
		field.Int("age").Optional(),
		field.String("password").Permissions(permissions.WriteOnly),
		field.Float("score").Default(4.2),
		field.String("uuid").Immutable().DefaultFunc(logic.GetUuidStr),
		field.Bool("is_admin").ProtoField(5),
		field.Byte("api_key").DefaultFunc(logic.GenerateAPIKey),
		field.Int64("last_login_ms"),
		field.Time("created_at").Permissions(permissions.ReadOnly).DefaultFunc(time.Now).ProtoField(6).Immutable(),
		field.Time("updated_at").Permissions(permissions.ReadOnly).DefaultFunc(time.Now).ProtoField(7),
	}
}

func (User) Queries() []entlite.Query {
	return []entlite.Query{
		// query.DefaultCRUD(),
		// query.Create(),
		// query.Get(),
		query.Update(),
		query.Delete(),
		query.List(),
		query.GetBy("email"),
		query.ListBy("age"),
	}
}

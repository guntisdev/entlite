package ent

import (
	"github.com/guntisdev/entlite/pkg/entlite"
	"github.com/guntisdev/entlite/pkg/entlite/field"
)

type User struct {
	entlite.Schema
}

func (User) Annotations() []entlite.Annotation {
	return []entlite.Annotation{
		entlite.Message(),
		entlite.Service(),
	}
}

func (User) Fields() []entlite.Field {
	return []entlite.Field{
		field.String("email").Unique().ProtoField(2),
		field.String("name").Comment("First name and surname"),
		field.Int32("age").Optional(),
		field.Bool("is_admin").ProtoField(5),
		field.Time("created_at").DefaultNow().ProtoField(6).Immutable(),
	}
}

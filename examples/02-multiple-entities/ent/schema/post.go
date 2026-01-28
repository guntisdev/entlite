package ent

import (
	"github.com/guntisdev/entlite/pkg/entlite"
	"github.com/guntisdev/entlite/pkg/entlite/field"
)

type Post struct {
	entlite.Schema
}

func (Post) Annotations() []entlite.Annotation {
	return []entlite.Annotation{
		entlite.Message(),
		entlite.Service(
			entlite.Methods(entlite.MethodGet | entlite.MethodList),
		),
	}
}

func (Post) Fields() []entlite.Field {
	return []entlite.Field{
		field.String("title").ProtoField(2),
		field.String("content").ProtoField(3),
		field.Bool("published").Default(false).ProtoField(4),
		// field.Time("created_at").Default("NOW()").ProtoField(5),
	}
}

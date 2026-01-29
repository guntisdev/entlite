package ent

import (
	"github.com/guntisdev/entlite/pkg/entlite"
	"github.com/guntisdev/entlite/pkg/entlite/field"
	"github.com/guntisdev/entlite/pkg/entlite/service"
)

type Post struct {
	entlite.Schema
}

func (Post) Annotations() []entlite.Annotation {
	return []entlite.Annotation{
		entlite.Message(),
		entlite.Service(
			service.Methods(service.MethodGet | service.MethodList),
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

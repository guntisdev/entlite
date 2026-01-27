package ent

import "github.com/guntisdev/entlite/pkg/entlite"

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
		entlite.String("title").ProtoField(2),
		entlite.String("content").ProtoField(3),
		entlite.Bool("published").Default(false).ProtoField(4),
		// entlite.Time("created_at").Default("NOW()").ProtoField(5),
	}
}

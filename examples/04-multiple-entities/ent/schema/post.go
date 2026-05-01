package schema

import (
	"github.com/guntisdev/entlite/pkg/entlite"
	"github.com/guntisdev/entlite/pkg/entlite/field"
	"github.com/guntisdev/entlite/pkg/entlite/query"
)

type Post struct {
	entlite.Schema
}

func (Post) Annotations() []entlite.Annotation {
	return []entlite.Annotation{
		entlite.Message(),
		entlite.GRPC(),
	}
}

func (Post) Fields() []entlite.Field {
	return []entlite.Field{
		field.String("title").ProtoField(2),
		field.String("content").ProtoField(3),
		field.Bool("published").Default(false).ProtoField(4),
	}
}

func (Post) Queries() []entlite.Query {
	return []entlite.Query{
		query.Get(),
		query.List(),
	}
}

package schema

import (
	"github.com/guntisdev/entlite/pkg/entlite"
	"github.com/guntisdev/entlite/pkg/entlite/field"
	"github.com/guntisdev/entlite/pkg/entlite/grpc"
)

type Post struct {
	entlite.Schema
}

func (Post) Annotations() []entlite.Annotation {
	return []entlite.Annotation{
		entlite.Message(),
		entlite.GRPC(
			grpc.Methods(grpc.MethodGet | grpc.MethodList),
		),
	}
}

func (Post) Fields() []entlite.Field {
	return []entlite.Field{
		field.String("title").ProtoField(2),
		field.String("content").ProtoField(3),
		field.Bool("published").Default(false).ProtoField(4),
	}
}

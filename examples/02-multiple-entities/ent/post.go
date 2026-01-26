package ent

import "github.com/guntisdev/entlite/pkg/entlite"

type Post struct {
	entlite.Schema
}

func (Post) Annotations() []entlite.Annotation {
	return []entlite.Annotation{
		entlite.Message(),
	}
}

func (Post) Field() []entlite.Field {
	return []entlite.Field{
		entlite.String("title").ProtoField(2).Default("helloworld"),
		entlite.String("email"),
		entlite.Bool("published"),
		entlite.Int32("age"),
	}
}

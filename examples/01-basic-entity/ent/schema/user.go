package ent

import "github.com/guntisdev/entlite/pkg/entlite"

// User entity definition - demonstrates basic field types and annotations
type User struct {
	entlite.Schema
}

func (User) Annotations() []entlite.Annotation {
	return []entlite.Annotation{
		entlite.Message(), // Generate proto message
		entlite.Service(), // Generate CRUD service
	}
}

func (User) Fields() []entlite.Field {
	return []entlite.Field{
		entlite.String("email").Unique().ProtoField(2),
		entlite.String("name").ProtoField(3),
		entlite.Int("age").ProtoField(4),
		entlite.Bool("is_admin").ProtoField(5),
		entlite.Time("created_at").Default("NOW()").ProtoField(6),
	}
}


package schema

import (
	"github.com/guntisdev/entlite/pkg/entlite"
	"github.com/guntisdev/entlite/pkg/entlite/field"
	"github.com/guntisdev/entlite/pkg/entlite/query"
)

type Standing struct {
	entlite.Schema
}

// proto only: counted from matches on every request, there is no standings table
func (Standing) Contracts() []entlite.Contract {
	return []entlite.Contract{
		entlite.PROTO(),
	}
}

func (Standing) Fields() []entlite.Field {
	return []entlite.Field{
		// entlite always adds an id field, here it carries the rank

		// Place in the table, 1 is best
		field.Int("id"),
		field.String("player"),
		field.Int("played"),
		field.Int("wins"),
		field.Int("draws"),
		field.Int("losses"),
		// Win 1, draw 0.5
		field.Float("points"),
	}
}

func (Standing) Queries() []entlite.Query {
	return []entlite.Query{
		query.ListAll(),
	}
}

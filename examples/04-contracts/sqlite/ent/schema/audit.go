package schema

import (
	"time"

	"github.com/guntisdev/entlite/pkg/entlite"
	"github.com/guntisdev/entlite/pkg/entlite/field"
	"github.com/guntisdev/entlite/pkg/entlite/query"
)

type Audit struct {
	entlite.Schema
}

// sqlc only: the club keeps this table for itself, no proto message, no rpc
func (Audit) Contracts() []entlite.Contract {
	return []entlite.Contract{
		entlite.SQLC(),
	}
}

func (Audit) Fields() []entlite.Field {
	return []entlite.Field{
		// created | deleted
		field.String("action"),
		field.Int("match_id"),
		field.String("detail"),
		field.Time("created_at").DefaultFunc(time.Now),
	}
}

// queries are called by the server, they never become rpc
func (Audit) Queries() []entlite.Query {
	return []entlite.Query{
		query.Create(),
		query.ListAll(),
	}
}

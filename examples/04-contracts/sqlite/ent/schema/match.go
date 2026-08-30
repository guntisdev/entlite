package schema

import (
	"time"

	"github.com/guntisdev/entlite/examples/04-contracts/sqlite/ent/logic"
	"github.com/guntisdev/entlite/pkg/entlite"
	"github.com/guntisdev/entlite/pkg/entlite/field"
	"github.com/guntisdev/entlite/pkg/entlite/query"
)

type Match struct {
	entlite.Schema
}

// both contracts: stored in sqlite and served to clients
func (Match) Contracts() []entlite.Contract {
	return []entlite.Contract{
		entlite.SQLC(),
		entlite.PROTO(),
	}
}

func (Match) Fields() []entlite.Field {
	return []entlite.Field{
		field.String("white").Validate(logic.NotBlank),
		field.String("black").Validate(logic.NotBlank),
		// 1-0 | 0-1 | 1/2-1/2
		field.String("result").Validate(logic.IsKnownResult),
		// e.g. Sicilian Defence
		field.String("opening").Optional(),
		field.Int("moves"),
		field.Time("played_at").DefaultFunc(time.Now),
		field.Time("created_at").Contracts(entlite.SQLC(), entlite.PROTO().ReadOnly()).DefaultFunc(time.Now).Immutable(),
	}
}

func (Match) Queries() []entlite.Query {
	return []entlite.Query{
		query.Create(),
		query.Get(),
		query.Delete(),
		query.ListAll(),
		query.DeleteAll().Contracts(entlite.SQLC()), // the server runs it, no rpc
	}
}

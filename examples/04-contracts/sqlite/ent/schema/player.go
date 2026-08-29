package schema

import (
	"time"

	"github.com/guntisdev/entlite/pkg/entlite"
	"github.com/guntisdev/entlite/pkg/entlite/field"
	"github.com/guntisdev/entlite/pkg/entlite/query"
)

// Player is the club roster. The secretary maintains it on the server,
// members only read it.
type Player struct {
	entlite.Schema
}

// both contracts, but proto is read only: the roster gets no write rpc
func (Player) Contracts() []entlite.Contract {
	return []entlite.Contract{
		entlite.SQLC(),
		entlite.PROTO().ReadOnly(),
	}
}

func (Player) Fields() []entlite.Field {
	return []entlite.Field{
		field.String("name").Unique(),
		// Elo rating, kept by the secretary
		field.Int("rating"),
		// e.g. GM, IM, FM
		field.String("title").Optional(),
		field.Time("joined_at").Contracts(entlite.SQLC(), entlite.PROTO().ReadOnly()).DefaultFunc(time.Now).Immutable(),
	}
}

func (Player) Queries() []entlite.Query {
	return []entlite.Query{
		query.Create(), // stays a db query, the read only proto contract keeps it out of the service
		query.Get(),
		query.ListAll(),
	}
}

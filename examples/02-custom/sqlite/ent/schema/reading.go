package schema

import (
	"time"

	"github.com/guntisdev/entlite/examples/02-custom/sqlite/ent/logic"
	"github.com/guntisdev/entlite/pkg/entlite"
	"github.com/guntisdev/entlite/pkg/entlite/field"
	"github.com/guntisdev/entlite/pkg/entlite/filter"
	"github.com/guntisdev/entlite/pkg/entlite/query"
)

// Reading is a single measurement captured by a Sensor.
type Reading struct {
	entlite.Schema
}

func (Reading) Contracts() []entlite.Contract {
	return []entlite.Contract{
		entlite.SQLC(),
		entlite.PROTO(),
	}
}

func (Reading) Fields() []entlite.Field {
	return []entlite.Field{
		field.Int64("ID"),
		// References sensor.ID
		field.Int("sensor_id"),
		field.Float("value"),
		// Signal quality 0-100
		field.Int("quality").Validate(logic.IsPercentage),
		// Marked as anomalous by ingestion
		field.Bool("flagged").Default(false),
		// Device measurement time (client-supplied)
		field.Time("recorded_at"),
		field.Time("created_at").Contracts(entlite.SQLC(), entlite.PROTO().ReadOnly()).DefaultFunc(time.Now).Immutable(),
	}
}

func (Reading) Queries() []entlite.Query {
	return []entlite.Query{
		query.Create(),
		query.Get(),
		query.Update().Contracts(entlite.SQLC()), // a recorded fact, clients never edit it, no proto rpc
		query.Delete(),
		query.ListBy("sensor_id"),
		query.ListBy(
			filter.Eq("sensor_id"),
			filter.Range("recorded_at"),
			filter.Eq("flagged"),
		).Count().OrderBy("recorded_at"),
	}
}

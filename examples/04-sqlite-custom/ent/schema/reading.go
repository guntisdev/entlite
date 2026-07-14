package schema

import (
	"time"

	"github.com/guntisdev/entlite/examples/04-sqlite-custom/ent/logic"
	"github.com/guntisdev/entlite/pkg/entlite"
	"github.com/guntisdev/entlite/pkg/entlite/field"
	"github.com/guntisdev/entlite/pkg/entlite/filter"
	"github.com/guntisdev/entlite/pkg/entlite/permissions"
	"github.com/guntisdev/entlite/pkg/entlite/query"
)

// Reading is a single measurement captured by a Sensor.
type Reading struct {
	entlite.Schema
}

func (Reading) Annotations() []entlite.Annotation {
	return []entlite.Annotation{
		entlite.Message(),
		entlite.GRPC(),
	}
}

func (Reading) Fields() []entlite.Field {
	return []entlite.Field{
		field.Int("sensor_id").Comment("References sensor.ID"),
		field.Float("value"),
		field.Int("quality").Validate(logic.IsPercentage).Comment("Signal quality 0-100"),
		field.Bool("flagged").Default(false).Comment("Marked as anomalous by ingestion"),
		field.Time("recorded_at").Comment("Device measurement time (client-supplied)"),
		field.Time("created_at").Permissions(permissions.ReadOnly).DefaultFunc(time.Now).Immutable(),
	}
}

func (Reading) Queries() []entlite.Query {
	return []entlite.Query{
		query.DefaultCRUD(),
		query.ListBy("sensor_id"),
		query.ListBy(
			filter.Eq("sensor_id"),
			filter.Range("recorded_at"),
			filter.Eq("flagged"),
		).Count().OrderBy("recorded_at"),
	}
}

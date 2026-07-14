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

// Sensor is a physical device deployed in the field that emits Readings.
type Sensor struct {
	entlite.Schema
}

func (Sensor) Annotations() []entlite.Annotation {
	return []entlite.Annotation{
		entlite.Message(),
		entlite.GRPC(),
	}
}

func (Sensor) Fields() []entlite.Field {
	return []entlite.Field{
		field.String("code").Unique().Comment("External hardware identifier, e.g. TEMP-A1"),
		field.String("label").Comment("Human friendly name"),
		field.String("kind").Validate(logic.IsKnownSensorKind).Comment("temperature | humidity | pressure | motion"),
		field.String("unit").Comment("Measurement unit, e.g. celsius"),
		field.String("location").Optional(),
		field.Bool("active").Default(true),
		field.String("firmware").Default("1.0.0"),
		field.Int("sample_rate_ms").Default(1000).Comment("Sampling interval in milliseconds"),
		field.Time("installed_at").Immutable().Comment("When the device was physically installed (client-supplied)"),
		field.Time("created_at").Permissions(permissions.ReadOnly).DefaultFunc(time.Now).Immutable(),
		field.Time("updated_at").Permissions(permissions.ReadOnly).DefaultFunc(time.Now),
	}
}

func (Sensor) Queries() []entlite.Query {
	return []entlite.Query{
		query.DefaultCRUD(),
		query.GetBy("code"),
		query.ListBy(
			filter.Search("label"),
			filter.Eq("kind"),
			filter.Eq("active").Optional(),
		).Count().OrderBy("installed_at"),
	}
}

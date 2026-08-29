package schema

import (
	"time"

	"github.com/guntisdev/entlite/examples/03-optional/sqlite/ent/logic"
	"github.com/guntisdev/entlite/pkg/entlite"
	"github.com/guntisdev/entlite/pkg/entlite/field"
	"github.com/guntisdev/entlite/pkg/entlite/filter"
	"github.com/guntisdev/entlite/pkg/entlite/query"
)

type Article struct {
	entlite.Schema
}

func (Article) Contracts() []entlite.Contract {
	return []entlite.Contract{
		entlite.SQLC(),
		entlite.PROTO(),
	}
}

func (Article) Fields() []entlite.Field {
	return []entlite.Field{
		// uuid primary key, generated on the server and not part of requests
		field.String("id").Contracts(entlite.SQLC(), entlite.PROTO().ReadOnly()).Immutable().DefaultFunc(logic.NewUUID),

		// --- required fields ---

		// Human/URL identifier, e.g. hello-world
		field.String("slug").Unique(),
		field.String("title").Validate(logic.NotBlank),
		field.String("author"),

		// --- optional fields, one per optional type ---

		field.String("subtitle").Optional(),      // string
		field.Int("reading_minutes").Optional(),  // int32, estimated read time
		field.Int64("last_viewed_ms").Optional(), // int64, epoch millis
		field.Float("rating").Optional(),         // float, null until rated
		field.Byte("cover_image").Optional(),     // bytes, raw image
		field.Time("published_at").Optional(),    // time, null means draft

		// Free-form metadata, e.g. {"og_image":"/cover.png"}
		field.JSON("metadata").Optional(),

		// --- required flag and server managed timestamps ---

		field.Bool("is_featured").Default(false), // bool cannot be optional, use a default
		field.Time("created_at").Contracts(entlite.SQLC(), entlite.PROTO().ReadOnly()).DefaultFunc(time.Now).Immutable(),
		field.Time("updated_at").Contracts(entlite.SQLC(), entlite.PROTO().ReadOnly()).DefaultFunc(time.Now),
	}
}

func (Article) Queries() []entlite.Query {
	return []entlite.Query{
		query.DefaultCRUD(),
		query.GetBy("slug"),
		query.ListBy("author"),
		query.ListAll(),

		// is_featured is a required field, but an optional filter
		query.ListBy(
			filter.Eq("author"),
			filter.Eq("is_featured").Optional(),
			filter.Range("published_at").Optional(),
			filter.Search("title").Optional(),
		).OrderBy("published_at").Count(),
	}
}

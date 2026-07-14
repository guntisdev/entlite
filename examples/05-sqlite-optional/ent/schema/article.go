package schema

import (
	"time"

	"github.com/guntisdev/entlite/examples/05-sqlite-optional/ent/logic"
	"github.com/guntisdev/entlite/pkg/entlite"
	"github.com/guntisdev/entlite/pkg/entlite/field"
	"github.com/guntisdev/entlite/pkg/entlite/filter"
	"github.com/guntisdev/entlite/pkg/entlite/permissions"
	"github.com/guntisdev/entlite/pkg/entlite/query"
)

// Article is a single piece of content in a CMS. Most descriptive attributes
// are genuinely optional, which makes it a natural showcase for optional
// fields and optional query filters.
type Article struct {
	entlite.Schema
}

func (Article) Annotations() []entlite.Annotation {
	return []entlite.Annotation{
		entlite.Message(),
		entlite.GRPC(),
	}
}

func (Article) Fields() []entlite.Field {
	return []entlite.Field{
		// UUID primary key. A field named "id" is treated as the primary key;
		// the String type makes it a TEXT PRIMARY KEY, and a server-side
		// DefaultFunc means the value is always generated. ReadOnly keeps it out
		// of the Create/Update request bodies - clients never set the id.
		field.String("id").Permissions(permissions.ReadOnly).Immutable().DefaultFunc(logic.NewUUID),

		// --- required core ---
		field.String("slug").Unique().Comment("Human/URL identifier, e.g. hello-world"),
		field.String("title").Validate(logic.NotBlank),
		field.String("author"),

		// --- optional fields: one per optional-capable type ---
		field.String("subtitle").Optional(),     // optional string
		field.Int("reading_minutes").Optional(),  // optional int32 - estimated read time
		field.Int64("last_viewed_ms").Optional(), // optional int64 - epoch millis, null until first view
		field.Float("rating").Optional(),         // optional float - null until first rated
		field.Byte("cover_image").Optional(),     // optional bytes - raw image, may be absent
		field.Time("published_at").Optional(),    // optional time  - null means still a draft

		// --- non-optional flag + server-managed timestamps ---
		field.Bool("is_featured").Default(false), // bool cannot be optional; use a default instead
		field.Time("created_at").Permissions(permissions.ReadOnly).DefaultFunc(time.Now).Immutable(),
		field.Time("updated_at").Permissions(permissions.ReadOnly).DefaultFunc(time.Now),
	}
}

func (Article) Queries() []entlite.Query {
	return []entlite.Query{
		query.DefaultCRUD(),
		query.GetBy("slug"),

		// Simple list: a single required equality filter (author = ?).
		query.ListBy("author"),

		// Faceted search: one required filter plus several OPTIONAL filters.
		// Optional filters become `optional` proto fields backed by NULL-aware
		// SQL, so a caller only sends the facets it actually wants to filter by.
		// Note that `is_featured` is a required column but an optional *filter* -
		// the two are independent choices.
		query.ListBy(
			filter.Eq("author"),
			filter.Eq("is_featured").Optional(),
			filter.Range("published_at").Optional(),
			filter.Search("title").Optional(),
		).OrderBy("published_at").Count(),
	}
}

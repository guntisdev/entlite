package db

import (
	internal "github.com/guntisdev/entlite/examples/04-sqlite-custom/ent/gen/db/internal"
)

type DBTX = internal.DBTX
func New(db DBTX) *Queries { return (*Queries)(internal.New(db)) }
type Queries internal.Queries

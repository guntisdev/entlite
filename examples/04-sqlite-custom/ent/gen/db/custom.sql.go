package db

import (
	"context"
	"fmt"
	"github.com/guntisdev/entlite/examples/04-sqlite-custom/ent/logic"
	"time"
	internal "github.com/guntisdev/entlite/examples/04-sqlite-custom/ent/gen/db/internal"
)

type GetSensorReadingStatsRow = internal.GetSensorReadingStatsRow
type ListSensorsWithLatestReadingParams = internal.ListSensorsWithLatestReadingParams
type ListSensorsWithLatestReadingRow = internal.ListSensorsWithLatestReadingRow

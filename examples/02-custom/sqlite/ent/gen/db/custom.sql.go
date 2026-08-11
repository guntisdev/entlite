package db

import (
	"context"
	"time"
	internal "github.com/guntisdev/entlite/examples/02-custom/sqlite/ent/gen/db/internal"
)

type GetSensorReadingStatsParams = internal.GetSensorReadingStatsParams
type GetSensorReadingStatsRow = internal.GetSensorReadingStatsRow
func (q *Queries) GetSensorReadingStats(ctx context.Context, arg internal.GetSensorReadingStatsParams) (internal.GetSensorReadingStatsRow, error) {
	return (*internal.Queries)(q).GetSensorReadingStats(ctx, arg)
}

type ListSensorsWithLatestReadingParams = internal.ListSensorsWithLatestReadingParams
type ListSensorsWithLatestReadingRow = internal.ListSensorsWithLatestReadingRow
func (q *Queries) ListSensorsWithLatestReading(ctx context.Context, arg internal.ListSensorsWithLatestReadingParams) ([]internal.ListSensorsWithLatestReadingRow, error) {
	return (*internal.Queries)(q).ListSensorsWithLatestReading(ctx, arg)
}

func (q *Queries) PruneReadingsOlderThan(ctx context.Context, cutoff time.Time) (int64, error) {
	return (*internal.Queries)(q).PruneReadingsOlderThan(ctx, cutoff)
}


package db

import (
	"context"
	"time"
	internal "github.com/guntisdev/entlite/examples/04-sqlite-custom/ent/gen/db/internal"
)

type GetSensorReadingStatsRow = internal.GetSensorReadingStatsRow
func (q *Queries) GetSensorReadingStats(ctx context.Context, sensorID int64) (internal.GetSensorReadingStatsRow, error) {
	return (*internal.Queries)(q).GetSensorReadingStats(ctx, sensorID)
}

type ListSensorsWithLatestReadingParams = internal.ListSensorsWithLatestReadingParams
type ListSensorsWithLatestReadingRow = internal.ListSensorsWithLatestReadingRow
func (q *Queries) ListSensorsWithLatestReading(ctx context.Context, arg internal.ListSensorsWithLatestReadingParams) ([]internal.ListSensorsWithLatestReadingRow, error) {
	return (*internal.Queries)(q).ListSensorsWithLatestReading(ctx, arg)
}

func (q *Queries) PruneReadingsOlderThan(ctx context.Context, cutoff time.Time) (int64, error) {
	return (*internal.Queries)(q).PruneReadingsOlderThan(ctx, cutoff)
}


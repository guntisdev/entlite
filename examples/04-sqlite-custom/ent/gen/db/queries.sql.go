package db

import (
	"context"
	"fmt"
	"github.com/guntisdev/entlite/examples/04-sqlite-custom/ent/logic"
	"time"
	internal "github.com/guntisdev/entlite/examples/04-sqlite-custom/ent/gen/db/internal"
)

type CreateReadingParams struct {
	Value float64 `json:"value"`
	Quality int32 `json:"quality"`
	Flagged *bool `json:"flagged"`
	RecordedAt *time.Time `json:"recorded_at"`
}

func (q *Queries) CreateReading(ctx context.Context, arg CreateReadingParams) (int32, error) {
	if !logic.IsPercentage(arg.Quality) {
		return 0, fmt.Errorf("Failed create: incorrect value for 'Reading' in field 'quality', validated by 'logic.IsPercentage'")
	}
	internalArg := internal.CreateReadingParams{
		SensorId: IntConvert[int32, int64](arg.SensorId),
		Value: arg.Value,
		Quality: IntConvert[int32, int64](arg.Quality),
		Flagged: OptionalWithFallback(SQLiteBoolToInt(arg.Flagged), false),
		RecordedAt: OptionalWithFallback(arg.RecordedAt, time.Now()),
		CreatedAt: time.Now(),
	}
	id, err := (*internal.Queries)(q).CreateReading(ctx, internalArg)
	return IntConvert[int64, int32](id), err
}

type CreateSensorParams struct {
	Code string `json:"code"`
	Label string `json:"label"`
	Kind string `json:"kind"`
	Unit string `json:"unit"`
	Location *string `json:"location"`
	Active *bool `json:"active"`
	Firmware *string `json:"firmware"`
	SampleRateMs *int32 `json:"sample_rate_ms"`
}

func (q *Queries) CreateSensor(ctx context.Context, arg CreateSensorParams) (int32, error) {
	if !logic.IsKnownSensorKind(arg.Kind) {
		return 0, fmt.Errorf("Failed create: incorrect value for 'Sensor' in field 'kind', validated by 'logic.IsKnownSensorKind'")
	}
	internalArg := internal.CreateSensorParams{
		Code: arg.Code,
		Label: arg.Label,
		Kind: arg.Kind,
		Unit: arg.Unit,
		Location: arg.Location,
		Active: OptionalWithFallback(SQLiteBoolToInt(arg.Active), true),
		Firmware: OptionalWithFallback(arg.Firmware, "1.0.0"),
		SampleRateMs: OptionalWithFallback(IntConvert[int32, int64](arg.SampleRateMs), 1000),
		InstalledAt: time.Now(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	id, err := (*internal.Queries)(q).CreateSensor(ctx, internalArg)
	return IntConvert[int64, int32](id), err
}

func (q *Queries) DeleteReading(ctx context.Context, id int32) error {
	return (*internal.Queries)(q).DeleteReading(ctx, IntConvert[int32, int64](id))
}

func (q *Queries) DeleteSensor(ctx context.Context, id int32) error {
	return (*internal.Queries)(q).DeleteSensor(ctx, IntConvert[int32, int64](id))
}

func (q *Queries) GetReadingByID(ctx context.Context, id int32) (*Reading, error) {
	dbResult, err := (*internal.Queries)(q).GetReadingByID(ctx, IntConvert[int32, int64](id))
	if err != nil {
		return nil, err
	}
	return ReadingFromSQL(&dbResult), nil
}

func (q *Queries) GetSensorByCode(ctx context.Context, code string) (*Sensor, error) {
	dbResult, err := (*internal.Queries)(q).GetSensorByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	return SensorFromSQL(&dbResult), nil
}

func (q *Queries) GetSensorByID(ctx context.Context, id int32) (*Sensor, error) {
	dbResult, err := (*internal.Queries)(q).GetSensorByID(ctx, IntConvert[int32, int64](id))
	if err != nil {
		return nil, err
	}
	return SensorFromSQL(&dbResult), nil
}

func (q *Queries) ListReadingBySensorId(ctx context.Context, sensorID int64) ([]*Reading, error) {
	dbResults, err := (*internal.Queries)(q).ListReadingBySensorId(ctx, sensorID)
	if err != nil {
		return nil, err
	}
	result := make([]*Reading, len(dbResults))
	for i := range dbResults {
		result[i] = ReadingFromSQL(&dbResults[i])
	}
	return result, nil
}

type ListReadingFilterBySensorIdRecordedAtFlaggedParams = internal.ListReadingFilterBySensorIdRecordedAtFlaggedParams
func (q *Queries) ListReadingFilterBySensorIdRecordedAtFlagged(ctx context.Context, arg ListReadingFilterBySensorIdRecordedAtFlaggedParams) ([]*Reading, error) {
	dbResults, err := (*internal.Queries)(q).ListReadingFilterBySensorIdRecordedAtFlagged(ctx, arg)
	if err != nil {
		return nil, err
	}
	result := make([]*Reading, len(dbResults))
	for i := range dbResults {
		result[i] = ReadingFromSQL(&dbResults[i])
	}
	return result, nil
}

type ListSensorFilterByLabelKindActiveParams = internal.ListSensorFilterByLabelKindActiveParams
func (q *Queries) ListSensorFilterByLabelKindActive(ctx context.Context, arg ListSensorFilterByLabelKindActiveParams) ([]*Sensor, error) {
	dbResults, err := (*internal.Queries)(q).ListSensorFilterByLabelKindActive(ctx, arg)
	if err != nil {
		return nil, err
	}
	result := make([]*Sensor, len(dbResults))
	for i := range dbResults {
		result[i] = SensorFromSQL(&dbResults[i])
	}
	return result, nil
}

type UpdateReadingParams struct {
	Value float64 `json:"value"`
	Quality int32 `json:"quality"`
	Flagged *bool `json:"flagged"`
	RecordedAt *time.Time `json:"recorded_at"`
	ID int32 `json:"ID"`
}

func (q *Queries) UpdateReading(ctx context.Context, arg UpdateReadingParams) (*Reading, error) {
	if !logic.IsPercentage(arg.Quality) {
		return nil, fmt.Errorf("Failed update: incorrect value for 'Reading' in field 'quality', validated by 'logic.IsPercentage'")
	}
	internalArg := internal.UpdateReadingParams{
		ID: IntConvert[int32, int64](arg.ID),
		SensorId: IntConvert[int32, int64](arg.SensorId),
		Value: arg.Value,
		Quality: IntConvert[int32, int64](arg.Quality),
		Flagged: SQLiteBoolToInt(arg.Flagged),
		RecordedAt: arg.RecordedAt,
	}

	dbReading, err := (*internal.Queries)(q).UpdateReading(ctx, internalArg)
	if err != nil {
		return nil, err
	}
	return ReadingFromSQL(&dbReading), nil
}

type UpdateSensorParams struct {
	Code string `json:"code"`
	Label string `json:"label"`
	Kind string `json:"kind"`
	Unit string `json:"unit"`
	Location *string `json:"location"`
	Active *bool `json:"active"`
	Firmware *string `json:"firmware"`
	SampleRateMs *int32 `json:"sample_rate_ms"`
	ID int32 `json:"ID"`
}

func (q *Queries) UpdateSensor(ctx context.Context, arg UpdateSensorParams) (*Sensor, error) {
	if !logic.IsKnownSensorKind(arg.Kind) {
		return nil, fmt.Errorf("Failed update: incorrect value for 'Sensor' in field 'kind', validated by 'logic.IsKnownSensorKind'")
	}
	internalArg := internal.UpdateSensorParams{
		ID: IntConvert[int32, int64](arg.ID),
		Code: arg.Code,
		Label: arg.Label,
		Kind: arg.Kind,
		Unit: arg.Unit,
		Location: arg.Location,
		Active: SQLiteBoolToInt(arg.Active),
		Firmware: arg.Firmware,
		SampleRateMs: IntPtrConvert[int32, int64](arg.SampleRateMs),
		UpdatedAt: time.Now(),
	}

	dbSensor, err := (*internal.Queries)(q).UpdateSensor(ctx, internalArg)
	if err != nil {
		return nil, err
	}
	return SensorFromSQL(&dbSensor), nil
}


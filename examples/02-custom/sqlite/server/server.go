package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/guntisdev/entlite/examples/02-custom/sqlite/ent/gen/db"
	"github.com/guntisdev/entlite/examples/02-custom/sqlite/ent/gen/pb"
)

// --- Sensor: generated CRUD -------------------------------------------------
type SensorServer struct {
	db *sql.DB
}

// enforces implementation of proto methods
var _ pb.SensorServiceHandler = (*SensorServer)(nil)

func NewSensorServiceServer(db *sql.DB) *SensorServer {
	return &SensorServer{
		db: db,
	}
}

func (s *SensorServer) Create(
	ctx context.Context,
	req *connect.Request[pb.CreateSensorRequest],
) (*connect.Response[pb.Sensor], error) {
	log.Printf("Create sensor: %+v", req.Msg)

	queries := db.New(s.db)

	sensorID, err := queries.CreateSensor(ctx, db.CreateSensorParams{
		Code:         req.Msg.Code,
		Label:        req.Msg.Label,
		Kind:         req.Msg.Kind,
		Unit:         req.Msg.Unit,
		Location:     req.Msg.Location,
		Active:       req.Msg.Active,
		Firmware:     req.Msg.Firmware,
		SampleRateMs: req.Msg.SampleRateMs,
		InstalledAt:  req.Msg.InstalledAt.AsTime(),
		// latest_value is permissions.Virtual: it exists on the proto message
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to create sensor: %w", err))
	}

	sensor, err := queries.GetSensorByID(ctx, sensorID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get created sensor: %w", err))
	}

	return connect.NewResponse(sensor.ToProto()), nil
}

func (s *SensorServer) GetByID(
	ctx context.Context,
	req *connect.Request[pb.GetSensorByIDRequest],
) (*connect.Response[pb.Sensor], error) {
	log.Printf("Get sensor: ID=%d", req.Msg.ID)

	queries := db.New(s.db)

	sensor, err := queries.GetSensorByID(ctx, req.Msg.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("sensor not found"))
		}
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get sensor: %w", err))
	}

	return connect.NewResponse(sensor.ToProto()), nil
}

func (s *SensorServer) GetByCode(
	ctx context.Context,
	req *connect.Request[pb.GetSensorByCodeRequest],
) (*connect.Response[pb.Sensor], error) {
	log.Printf("Get sensor by code: code=%s", req.Msg.Code)

	queries := db.New(s.db)

	sensor, err := queries.GetSensorByCode(ctx, req.Msg.Code)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("sensor not found"))
		}
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get sensor by code: %w", err))
	}

	return connect.NewResponse(sensor.ToProto()), nil
}

func (s *SensorServer) Update(
	ctx context.Context,
	req *connect.Request[pb.UpdateSensorRequest],
) (*connect.Response[pb.Sensor], error) {
	log.Printf("Update sensor: ID=%d, %+v", req.Msg.ID, req.Msg)

	queries := db.New(s.db)

	sensor, err := queries.UpdateSensor(ctx, db.UpdateSensorParams{
		ID:           req.Msg.ID,
		Code:         req.Msg.Code,
		Label:        req.Msg.Label,
		Kind:         req.Msg.Kind,
		Unit:         req.Msg.Unit,
		Location:     req.Msg.Location,
		Active:       req.Msg.Active,
		Firmware:     req.Msg.Firmware,
		SampleRateMs: req.Msg.SampleRateMs,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("sensor not found"))
		}
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to update sensor: %w", err))
	}

	return connect.NewResponse(sensor.ToProto()), nil
}

func (s *SensorServer) Delete(
	ctx context.Context,
	req *connect.Request[pb.DeleteSensorRequest],
) (*connect.Response[emptypb.Empty], error) {
	log.Printf("Delete sensor: ID=%d", req.Msg.ID)

	queries := db.New(s.db)

	if err := queries.DeleteSensor(ctx, req.Msg.ID); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to delete sensor: %w", err))
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (s *SensorServer) FilterByLabelKindActive(
	ctx context.Context,
	req *connect.Request[pb.ListSensorFilterByLabelKindActiveRequest],
) (*connect.Response[pb.ListSensorFilterByLabelKindActiveResponse], error) {
	log.Printf("Filter sensors: label=%s, kind=%s, active=%t",
		req.Msg.Label, req.Msg.Kind, req.Msg.GetActive())

	queries := db.New(s.db)

	dbSensors, err := queries.ListSensorFilterByLabelKindActive(ctx, db.ListSensorFilterByLabelKindActiveParams{
		Label:  req.Msg.Label, // filter.Search: compared with LIKE, so the caller supplies the wildcards
		Kind:   req.Msg.Kind,
		Active: req.Msg.GetActive(),
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to list sensors: %w", err))
	}

	pbSensors := make([]*pb.Sensor, len(dbSensors))
	for i, dbSensor := range dbSensors {
		pbSensors[i] = dbSensor.ToProto()
	}

	return connect.NewResponse(&pb.ListSensorFilterByLabelKindActiveResponse{
		Sensors: pbSensors,
	}), nil
}

// --- Reading: generated CRUD ------------------------------------------------
type ReadingServer struct {
	db *sql.DB
}

var _ pb.ReadingServiceHandler = (*ReadingServer)(nil)

func NewReadingServiceServer(db *sql.DB) *ReadingServer {
	return &ReadingServer{
		db: db,
	}
}

func (s *ReadingServer) Create(
	ctx context.Context,
	req *connect.Request[pb.CreateReadingRequest],
) (*connect.Response[pb.Reading], error) {
	log.Printf("Create reading: %+v", req.Msg)

	queries := db.New(s.db)

	readingID, err := queries.CreateReading(ctx, db.CreateReadingParams{
		SensorID:   req.Msg.SensorId,
		Value:      req.Msg.Value,
		Quality:    req.Msg.Quality,
		Flagged:    req.Msg.Flagged,
		RecordedAt: req.Msg.RecordedAt.AsTime(),
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to create reading: %w", err))
	}

	reading, err := queries.GetReadingByID(ctx, readingID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get created reading: %w", err))
	}

	return connect.NewResponse(reading.ToProto()), nil
}

func (s *ReadingServer) GetByID(
	ctx context.Context,
	req *connect.Request[pb.GetReadingByIDRequest],
) (*connect.Response[pb.Reading], error) {
	log.Printf("Get reading: ID=%d", req.Msg.ID)

	queries := db.New(s.db)

	reading, err := queries.GetReadingByID(ctx, req.Msg.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("reading not found"))
		}
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get reading: %w", err))
	}

	return connect.NewResponse(reading.ToProto()), nil
}

func (s *ReadingServer) Update(
	ctx context.Context,
	req *connect.Request[pb.UpdateReadingRequest],
) (*connect.Response[pb.Reading], error) {
	log.Printf("Update reading: ID=%d, %+v", req.Msg.ID, req.Msg)

	queries := db.New(s.db)

	reading, err := queries.UpdateReading(ctx, db.UpdateReadingParams{
		ID:         req.Msg.ID,
		SensorID:   req.Msg.SensorId,
		Value:      req.Msg.Value,
		Quality:    req.Msg.Quality,
		Flagged:    req.Msg.Flagged,
		RecordedAt: req.Msg.RecordedAt.AsTime(),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("reading not found"))
		}
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to update reading: %w", err))
	}

	return connect.NewResponse(reading.ToProto()), nil
}

func (s *ReadingServer) Delete(
	ctx context.Context,
	req *connect.Request[pb.DeleteReadingRequest],
) (*connect.Response[emptypb.Empty], error) {
	log.Printf("Delete reading: ID=%d", req.Msg.ID)

	queries := db.New(s.db)

	if err := queries.DeleteReading(ctx, req.Msg.ID); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to delete reading: %w", err))
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (s *ReadingServer) ListBySensorId(
	ctx context.Context,
	req *connect.Request[pb.ListReadingBySensorIdRequest],
) (*connect.Response[pb.ListReadingBySensorIdResponse], error) {
	log.Printf("List readings: sensor_id=%d", req.Msg.SensorId)

	queries := db.New(s.db)

	dbReadings, err := queries.ListReadingBySensorId(ctx, req.Msg.SensorId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to list readings: %w", err))
	}

	return connect.NewResponse(&pb.ListReadingBySensorIdResponse{
		Readings: toProtoReadings(dbReadings),
	}), nil
}

func (s *ReadingServer) FilterBySensorIdRecordedAtFlagged(
	ctx context.Context,
	req *connect.Request[pb.ListReadingFilterBySensorIdRecordedAtFlaggedRequest],
) (*connect.Response[pb.ListReadingFilterBySensorIdRecordedAtFlaggedResponse], error) {
	log.Printf("Filter readings: sensor_id=%d, recorded_at=%s..%s, flagged=%t",
		req.Msg.SensorId, req.Msg.MinRecordedAt.AsTime(), req.Msg.MaxRecordedAt.AsTime(), req.Msg.Flagged)

	queries := db.New(s.db)

	dbReadings, err := queries.ListReadingFilterBySensorIdRecordedAtFlagged(ctx, db.ListReadingFilterBySensorIdRecordedAtFlaggedParams{
		SensorID: req.Msg.SensorId,
		Flagged:  req.Msg.Flagged,
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to list readings: %w", err))
	}

	return connect.NewResponse(&pb.ListReadingFilterBySensorIdRecordedAtFlaggedResponse{
		Readings: toProtoReadings(dbReadings),
	}), nil
}

func toProtoReadings(dbReadings []*db.Reading) []*pb.Reading {
	pbReadings := make([]*pb.Reading, len(dbReadings))
	for i, dbReading := range dbReadings {
		pbReadings[i] = dbReading.ToProto()
	}
	return pbReadings
}

// --- Sensor analytics: hand-written custom.sql + custom.proto ---------------
type AnalyticsServer struct {
	db *sql.DB
}

var _ pb.SensorAnalyticsServiceHandler = (*AnalyticsServer)(nil)

func NewSensorAnalyticsServiceServer(db *sql.DB) *AnalyticsServer {
	return &AnalyticsServer{
		db: db,
	}
}

func (s *AnalyticsServer) GetReadingStats(
	ctx context.Context,
	req *connect.Request[pb.GetSensorReadingStatsRequest],
) (*connect.Response[pb.SensorReadingStats], error) {
	log.Printf("Reading stats: sensor_id=%d, %s..%s",
		req.Msg.SensorId, req.Msg.FromTs.AsTime(), req.Msg.ToTs.AsTime())

	queries := db.New(s.db)

	row, err := queries.GetSensorReadingStats(ctx, db.GetSensorReadingStatsParams{
		SensorID: int64(req.Msg.SensorId),
		FromTs:   req.Msg.FromTs.AsTime(),
		ToTs:     req.Msg.ToTs.AsTime(),
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get reading stats: %w", err))
	}

	// AVG/MIN/MAX are NULL when the window holds no rows
	stats := &pb.SensorReadingStats{
		ReadingCount: row.ReadingCount,
		MinValue:     anyToFloat(row.MinValue),
		MaxValue:     anyToFloat(row.MaxValue),
	}
	if row.AvgValue != nil {
		stats.AvgValue = *row.AvgValue
	}

	return connect.NewResponse(stats), nil
}

func (s *AnalyticsServer) ListWithLatestReading(
	ctx context.Context,
	req *connect.Request[pb.ListSensorsWithLatestReadingRequest],
) (*connect.Response[pb.ListSensorsWithLatestReadingResponse], error) {
	log.Printf("List sensors with latest reading: limit=%d, offset=%d", req.Msg.Limit, req.Msg.Offset)

	queries := db.New(s.db)

	rows, err := queries.ListSensorsWithLatestReading(ctx, db.ListSensorsWithLatestReadingParams{
		Limit:  int64(req.Msg.Limit),
		Offset: int64(req.Msg.Offset),
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to list sensors with latest reading: %w", err))
	}

	items := make([]*pb.SensorWithLatestReading, len(rows))
	for i, row := range rows {
		sensor := db.SensorFromSQL(&row.Sensor).ToProto()
		sensor.LatestValue = row.LatestValue

		item := &pb.SensorWithLatestReading{
			Sensor:      sensor,
			LatestValue: row.LatestValue,
		}
		if row.LatestRecordedAt != nil {
			item.LatestRecordedAt = timestamppb.New(*row.LatestRecordedAt)
		}
		items[i] = item
	}

	return connect.NewResponse(&pb.ListSensorsWithLatestReadingResponse{Items: items}), nil
}

func (s *AnalyticsServer) PruneReadings(
	ctx context.Context,
	req *connect.Request[pb.PruneReadingsRequest],
) (*connect.Response[pb.PruneReadingsResponse], error) {
	log.Printf("Prune readings older than %s", req.Msg.Cutoff.AsTime())

	queries := db.New(s.db)

	deleted, err := queries.PruneReadingsOlderThan(ctx, req.Msg.Cutoff.AsTime())
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to prune readings: %w", err))
	}

	return connect.NewResponse(&pb.PruneReadingsResponse{Deleted: deleted}), nil
}

func anyToFloat(v any) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case int64:
		return float64(n)
	default:
		return 0
	}
}

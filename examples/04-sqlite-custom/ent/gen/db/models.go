package db

import (
	"time"
	"google.golang.org/protobuf/types/known/timestamppb"
	pb "github.com/guntisdev/entlite/examples/04-sqlite-custom/ent/gen/pb"
	internal "github.com/guntisdev/entlite/examples/04-sqlite-custom/ent/gen/db/internal"
)


type Reading struct {
	ID int32 `json:"ID"`
	SensorId int32 `json:"sensor_id"`
	Value float64 `json:"value"`
	Quality *int32 `json:"quality"`
	Flagged bool `json:"flagged"`
	RecordedAt time.Time `json:"recorded_at"`
	CreatedAt time.Time `json:"created_at"`
}

type Sensor struct {
	ID int32 `json:"ID"`
	Code string `json:"code"`
	Label string `json:"label"`
	Kind string `json:"kind"`
	Unit string `json:"unit"`
	Location *string `json:"location"`
	Active bool `json:"active"`
	Firmware string `json:"firmware"`
	SampleRateMs int32 `json:"sample_rate_ms"`
	InstalledAt time.Time `json:"installed_at"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (m *Reading) ReadingToSQL() *internal.Reading {
	if m == nil {
		return nil
	}

	return &internal.Reading{
		ID: IntConvert[int32, int64](m.ID),
		SensorId: IntConvert[int32, int64](m.SensorId),
		Value: m.Value,
		Quality: IntPtrConvert[int32, int64](m.Quality),
		Flagged: SQLiteBoolToInt(m.Flagged),
		RecordedAt: m.RecordedAt,
		CreatedAt: m.CreatedAt,
	}
}

func ReadingFromSQL(db *internal.Reading) *Reading {
	if db == nil {
		return nil
	}

	return &Reading{
		ID: IntConvert[int64, int32](db.ID),
		SensorId: IntConvert[int64, int32](db.SensorId),
		Value: db.Value,
		Quality: IntPtrConvert[int64, int32](db.Quality),
		Flagged: SQLiteIntToBool(db.Flagged),
		RecordedAt: db.RecordedAt,
		CreatedAt: db.CreatedAt,
	}
}

func (m *Sensor) SensorToSQL() *internal.Sensor {
	if m == nil {
		return nil
	}

	return &internal.Sensor{
		ID: IntConvert[int32, int64](m.ID),
		Code: m.Code,
		Label: m.Label,
		Kind: m.Kind,
		Unit: m.Unit,
		Location: m.Location,
		Active: SQLiteBoolToInt(m.Active),
		Firmware: m.Firmware,
		SampleRateMs: IntConvert[int32, int64](m.SampleRateMs),
		InstalledAt: m.InstalledAt,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func SensorFromSQL(db *internal.Sensor) *Sensor {
	if db == nil {
		return nil
	}

	return &Sensor{
		ID: IntConvert[int64, int32](db.ID),
		Code: db.Code,
		Label: db.Label,
		Kind: db.Kind,
		Unit: db.Unit,
		Location: db.Location,
		Active: SQLiteIntToBool(db.Active),
		Firmware: db.Firmware,
		SampleRateMs: IntConvert[int64, int32](db.SampleRateMs),
		InstalledAt: db.InstalledAt,
		CreatedAt: db.CreatedAt,
		UpdatedAt: db.UpdatedAt,
	}
}

// ToProto converts Reading to proto format
func (m *Reading) ToProto() *pb.Reading {
	if m == nil {
		return nil
	}

	return &pb.Reading{
		ID: m.ID,
		SensorId: m.SensorId,
		Value: m.Value,
		Quality: m.Quality,
		Flagged: m.Flagged,
		RecordedAt: timestamppb.New(m.RecordedAt),
		CreatedAt: timestamppb.New(m.CreatedAt),
	}
}

// ToProto converts Sensor to proto format
func (m *Sensor) ToProto() *pb.Sensor {
	if m == nil {
		return nil
	}

	return &pb.Sensor{
		ID: m.ID,
		Code: m.Code,
		Label: m.Label,
		Kind: m.Kind,
		Unit: m.Unit,
		Location: m.Location,
		Active: m.Active,
		Firmware: m.Firmware,
		SampleRateMs: m.SampleRateMs,
		InstalledAt: timestamppb.New(m.InstalledAt),
		CreatedAt: timestamppb.New(m.CreatedAt),
		UpdatedAt: timestamppb.New(m.UpdatedAt),
	}
}


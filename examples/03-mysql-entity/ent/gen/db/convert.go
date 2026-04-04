package db

import (
	"database/sql"
	"reflect"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// TimeToProto converts time.Time to timestamppb.Timestamp pointer
func TimeToProto(t time.Time) *timestamppb.Timestamp {
	return timestamppb.New(t)
}

// Note: If the pointer is nil, it returns a zero time.Time{}
func ProtoToTime(t *timestamppb.Timestamp) time.Time {
	if t == nil {
		return time.Time{}
	}
	return t.AsTime()
}

// --- Time Converters ---
func NullTimeToProto(n sql.NullTime) *timestamppb.Timestamp {
	if !n.Valid {
		return nil
	}
	return timestamppb.New(n.Time)
}

func ProtoToNullTime(t *timestamppb.Timestamp) sql.NullTime {
	if t == nil {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{
		Time:  t.AsTime(),
		Valid: true,
	}
}

// OptionalWithFallback chooses fallback if optional value is nil
func OptionalWithFallback[T any](val *T, fallback T) T {
	if val == nil {
		return fallback
	}

	// For nil-able types like []byte, check if the dereferenced value is nil
	if reflect.ValueOf(any(*val)).IsNil() {
		return fallback
	}

	return *val
}
// --- Bytes Converters ---
func NullBytesToPtr(b []byte) *[]byte {
    if b == nil { return nil }
    return &b
}

func PtrToNullBytes(b *[]byte) []byte {
    if b == nil { return nil }
    return *b
}

// --- Nullable SQL Type Converters ---
func NullInt32ToPtr(n sql.NullInt32) *int32 {
	if !n.Valid {
		return nil
	}
	return &n.Int32
}

func PtrToNullInt32(p *int32) sql.NullInt32 {
	if p == nil {
		return sql.NullInt32{Valid: false}
	}
	return sql.NullInt32{
		Int32: *p,
		Valid: true,
	}
}

func NullInt64ToPtr(n sql.NullInt64) *int64 {
	if !n.Valid {
		return nil
	}
	return &n.Int64
}

func PtrToNullInt64(p *int64) sql.NullInt64 {
	if p == nil {
		return sql.NullInt64{Valid: false}
	}
	return sql.NullInt64{
		Int64: *p,
		Valid: true,
	}
}

func NullFloat64ToPtr(n sql.NullFloat64) *float64 {
	if !n.Valid {
		return nil
	}
	return &n.Float64
}

func PtrToNullFloat64(p *float64) sql.NullFloat64 {
	if p == nil {
		return sql.NullFloat64{Valid: false}
	}
	return sql.NullFloat64{
		Float64: *p,
		Valid:   true,
	}
}

func NullStringToPtr(n sql.NullString) *string {
	if !n.Valid {
		return nil
	}
	return &n.String
}

func PtrToNullString(p *string) sql.NullString {
	if p == nil {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{
		String: *p,
		Valid:  true,
	}
}

// --- SQLite bool converters ---
func SQLiteIntToBool(i int64) bool {
    switch i {
    case 0:
        return false
    case 1:
        return true
    default:
        panic("Unable convert sqlite int to bool")
    }
}

func SQLiteBoolToInt(b bool) int64 {
    if b {
        return 1
    } else {
        return 0
    }
}

// example: IntPtrConvert[int64, int32](dbRow.Age)
func IntPtrConvert[From, To ~int | ~int32 | ~int64 | ~float32 | ~float64](src *From) *To {
	if src == nil {
		return nil
	}
	val := To(*src)
	return &val
}
	
func IntConvert[From, To ~int | ~int8 | ~int16 | ~int32 | ~int64 | 
    ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | 
    ~float32 | ~float64](src From) To {
    return To(src)
}
// PtrBytesToNullString converts *[]byte to sql.NullString for MySQL compatibility
func PtrBytesToNullString(p *[]byte) sql.NullString {
    if p == nil || *p == nil {
        return sql.NullString{Valid: false}
    }
    return sql.NullString{
        String: string(*p),
        Valid:  true,
    }
}
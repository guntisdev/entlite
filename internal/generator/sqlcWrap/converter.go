package sqlcwrap

import (
	"strings"
)

func GenerateConvertFile(packageName string, hasTimeField bool) string {
	var content strings.Builder

	content.WriteString("package ")
	content.WriteString(packageName)
	content.WriteString("\n\n")

	content.WriteString("import (\n")
	content.WriteString("\t\"database/sql\"\n")
	if hasTimeField {
		content.WriteString("\t\"time\"\n\n")
		content.WriteString("\t\"google.golang.org/protobuf/types/known/timestamppb\"\n")
	}
	content.WriteString(")\n")

	content.WriteString(generateConverterFunctions(hasTimeField))

	return content.String()
}

func generateConverterFunctions(hasTimeField bool) string {
	var content strings.Builder

	if hasTimeField {
		content.WriteString(timeToproto)
	}
	if hasTimeField {
		content.WriteString(nullableTime)
	}
	content.WriteString(optionalWithFallback)
	content.WriteString(nullableBytes)
	content.WriteString(sqliteBools)
	content.WriteString(sqlLiteInts)

	return content.String()
}

const optionalWithFallback = `
// OptionalWithFallback chooses fallback if optional value is nil
func OptionalWithFallback[T any](val *T, fallback T) T {
    if val != nil {
        return *val
    }
    return fallback
}`

const timeToproto = `
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
`

const nullableTime = `
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
`

const nullableBytes = `
// --- Bytes Converters ---
func NullBytesToPtr(b []byte) *[]byte {
    if b == nil { return nil }
    return &b
}

func PtrToNullBytes(b *[]byte) []byte {
    if b == nil { return nil }
    return *b
}
`

const sqliteBools = `
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
`

const sqlLiteInts = `
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
}`

package convert

import (
	"bytes"
	"strings"
	"text/template"
)

func generateHelperFunctions() string {
	var content strings.Builder

	content.WriteString(timeToproto)
	content.WriteString(nullableConverters())
	content.WriteString(nullableTime)
	content.WriteString(nullableBytes)
	content.WriteString(sqliteBools)
	content.WriteString(sqlLiteInts)

	return content.String()
}

func nullableConverters() string {
	var buf bytes.Buffer

	type Config struct {
		Type      string
		Primitive string
	}

	fields := []Config{
		{Type: "Int32", Primitive: "int32"},
		{Type: "Int64", Primitive: "int64"},
		{Type: "Float64", Primitive: "float64"},
		{Type: "String", Primitive: "string"},
		{Type: "Bool", Primitive: "bool"},
	}

	for _, field := range fields {
		_ = nullableConvertersTemplate.Execute(&buf, field)
	}

	return buf.String()
}

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

var nullableConvertersTemplate = template.Must(template.New("converters").Parse(`
// --- {{.Type}} Converters ---
func Null{{.Type}}ToPtr(n sql.Null{{.Type}}) *{{.Primitive}} {
	if !n.Valid { return nil }
	return &n.{{.Type}}
}

func PtrToNull{{.Type}}(i *{{.Primitive}}) sql.Null{{.Type}} {
	if i == nil {
		return sql.Null{{.Type}}{Valid: false}
	}
	return sql.Null{{.Type}}{ {{.Type}}: *i, Valid: true }
}
`))

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
// --- SQLite int converters int32 - int64 ---
func SQLiteInt64ToInt32(n int64) int32 {
    if n < math.MinInt32 || n > math.MaxInt32 {
		panic("Unable convert sqlite int64 to int32")
	}
	return int32(n)
}

func SQLiteInt32ToInt64(n int32) int64 {
    return int64(n)
}

// --- SQLite null-int converters int32 - int64 ---
func SQLiteNullInt64ToPtrInt32(n sql.NullInt64) *int32 {
	if !n.Valid { return nil }
    v := SQLiteInt64ToInt32(n.Int64)
	return &v
}

func SQLitePtrInt32ToNullInt64(i *int32) sql.NullInt64 {
    if i == nil {
		return sql.NullInt64{Valid: false}
	}
	return sql.NullInt64{ Int64: int64(*i), Valid: true }
}
`

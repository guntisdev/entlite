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

package sqlcwrap

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/guntisdev/entlite/internal/schema"
	"github.com/guntisdev/entlite/internal/util"
)

// EntliteAccessFileName exposes the connection handle sqlc keeps unexported. sqlc
// only rewrites its own files, so this one survives regeneration.
const EntliteAccessFileName = "entlite_access.go"

// GenerateAccessFile emits EntliteAccessFileName. The wrapper needs the handle
// to open a transaction around multi-row operations.
func GenerateAccessFile(packageName string) string {
	return util.GeneratedGo + fmt.Sprintf(`package %s

// DB returns the handle this Queries was constructed with.
func (q *Queries) DB() DBTX { return q.db }
`, packageName)
}

// PackageNameOf reads the package clause of the first Go file in dir. sqlc takes
// the name from sqlc.yaml, so it cannot be guessed from the directory.
func PackageNameOf(dir string) (string, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}

	fset := token.NewFileSet()
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".go") {
			continue
		}
		node, err := parser.ParseFile(fset, filepath.Join(dir, file.Name()), nil, parser.PackageClauseOnly)
		if err != nil {
			return "", err
		}
		return node.Name.Name, nil
	}

	return "", fmt.Errorf("no Go files found in %s", dir)
}

// convertFileNeeds reports which optional parts of convert.go the entities call for
func convertFileNeeds(entities []schema.Entity) (hasTimeField bool, hasCreateBulk bool, hasJSONField bool) {
	for _, entity := range entities {
		for _, field := range entity.Fields {
			if field.Type == schema.FieldTypeTime {
				hasTimeField = true
			}
			if field.Type == schema.FieldTypeJSON && !field.IsVirtual() {
				hasJSONField = true
			}
		}
		for _, query := range entity.SQLCQueries() {
			if query.Type == schema.QueryCreateBulk {
				hasCreateBulk = true
			}
		}
	}

	return hasTimeField, hasCreateBulk, hasJSONField
}

func GenerateConvertFile(packageName string, entities []schema.Entity, sqlDialect schema.SQLDialect) string {
	hasTimeField, hasCreateBulk, hasJSONField := convertFileNeeds(entities)
	// sqlite keeps json as TEXT, only the native JSONB/JSON columns need converters
	needsJSON := hasJSONField && sqlDialect != schema.SQLite

	var content strings.Builder

	content.WriteString(util.GeneratedGo)
	content.WriteString("package ")
	content.WriteString(packageName)
	content.WriteString("\n\n")

	content.WriteString("import (\n")
	if hasCreateBulk {
		content.WriteString("\t\"context\"\n")
	}
	content.WriteString("\t\"database/sql\"\n")
	if needsJSON {
		content.WriteString("\t\"encoding/json\"\n")
	}
	content.WriteString("\t\"reflect\"\n")
	if hasTimeField {
		content.WriteString("\t\"time\"\n\n")
		content.WriteString("\t\"google.golang.org/protobuf/types/known/timestamppb\"\n")
	}
	if needsJSON && sqlDialect == schema.PostgreSQL {
		content.WriteString("\n\t\"github.com/sqlc-dev/pqtype\"\n")
	}
	content.WriteString(")\n")

	content.WriteString(generateConverterFunctions(hasTimeField, hasCreateBulk, needsJSON, sqlDialect))

	return content.String()
}

func generateConverterFunctions(hasTimeField bool, hasCreateBulk bool, needsJSON bool, sqlDialect schema.SQLDialect) string {
	var content strings.Builder

	if hasTimeField {
		content.WriteString(timeToproto)
	}
	if hasTimeField {
		content.WriteString(nullableTime)
	}
	if hasCreateBulk {
		content.WriteString(txBeginner)
	}
	content.WriteString(optionalWithFallback)
	content.WriteString(nullableBytes)
	content.WriteString(nullableTypes)
	content.WriteString(sqliteBools)
	content.WriteString(sqlLiteInts)
	content.WriteString(mysqlBytes)
	if needsJSON {
		content.WriteString(rawMessageTypes)
		if sqlDialect == schema.PostgreSQL {
			content.WriteString(nullRawMessageTypes)
		} else {
			content.WriteString(mysqlRawMessage)
		}
	}

	return content.String()
}

const txBeginner = `
// txBeginner is satisfied by *sql.DB and *sql.Conn but not *sql.Tx: a Queries already
// in a transaction runs inside the caller's one instead of nesting.
type txBeginner interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}
`

const optionalWithFallback = `
// OptionalWithFallback chooses fallback if optional value is nil
func OptionalWithFallback[T any](val *T, fallback T) T {
	if val == nil {
		return fallback
	}

	// nil-able types like []byte need a nil check, and IsNil panics on other kinds
	switch value := reflect.ValueOf(any(*val)); value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		if value.IsNil() {
			return fallback
		}
	}

	return *val
}`

const timeToproto = `
// TimeToProto converts time.Time to timestamppb.Timestamp pointer
func TimeToProto(t time.Time) *timestamppb.Timestamp {
	return timestamppb.New(t)
}

// A nil pointer returns a zero time.Time{}
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

const nullableTypes = `
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

func NullBoolToPtr(n sql.NullBool) *bool {
	if !n.Valid {
		return nil
	}
	return &n.Bool
}

func PtrToNullBool(p *bool) sql.NullBool {
	if p == nil {
		return sql.NullBool{Valid: false}
	}
	return sql.NullBool{
		Bool:  *p,
		Valid: true,
	}
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

// SQLiteBoolPtrToInt64Ptr converts *bool to *int64 for nullable columns via sqlc.narg
func SQLiteBoolPtrToInt64Ptr(b *bool) *int64 {
    if b == nil {
        return nil
    }
    v := SQLiteBoolToInt(*b)
    return &v
}

// SQLiteInt64PtrToBoolPtr is the inverse of SQLiteBoolPtrToInt64Ptr.
func SQLiteInt64PtrToBoolPtr(i *int64) *bool {
    if i == nil {
        return nil
    }
    v := SQLiteIntToBool(*i)
    return &v
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

const mysqlBytes = `
// PtrBytesToNullString converts *[]byte to sql.NullString for MySQL compatibility
func PtrBytesToNullString(p *[]byte) sql.NullString {
    if p == nil || *p == nil {
        return sql.NullString{Valid: false}
    }
    return sql.NullString{
        String: string(*p),
        Valid:  true,
    }
}`

const rawMessageTypes = `
// --- JSON Converters ---
// postgres JSONB and mysql JSON come out of sqlc as json.RawMessage
func StringToRawMessage(s string) json.RawMessage {
    return json.RawMessage(s)
}

func RawMessageToString(r json.RawMessage) string {
    return string(r)
}
`

const mysqlRawMessage = `
// mysql keeps a nullable JSON column as json.RawMessage, nil means NULL
func PtrToRawMessage(p *string) json.RawMessage {
    if p == nil {
        return nil
    }
    return json.RawMessage(*p)
}

func RawMessageToPtr(r json.RawMessage) *string {
    if r == nil {
        return nil
    }
    s := string(r)
    return &s
}
`

const nullRawMessageTypes = `
// postgres keeps a nullable JSONB column as pqtype.NullRawMessage
func PtrToNullRawMessage(p *string) pqtype.NullRawMessage {
    if p == nil {
        return pqtype.NullRawMessage{Valid: false}
    }
    return pqtype.NullRawMessage{
        RawMessage: json.RawMessage(*p),
        Valid:      true,
    }
}

func NullRawMessageToPtr(n pqtype.NullRawMessage) *string {
    if !n.Valid {
        return nil
    }
    s := string(n.RawMessage)
    return &s
}
`

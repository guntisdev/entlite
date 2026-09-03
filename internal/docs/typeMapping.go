package docs

import (
	"github.com/guntisdev/entlite/internal/generator/proto"
	"github.com/guntisdev/entlite/internal/generator/sqlc"
	sqlcwrap "github.com/guntisdev/entlite/internal/generator/sqlcWrap"
	"github.com/guntisdev/entlite/internal/schema"
)

var fieldTypes = []struct {
	Builder string
	Type    schema.FieldType
}{
	{Builder: "field.String", Type: schema.FieldTypeString},
	{Builder: "field.Bool", Type: schema.FieldTypeBool},
	{Builder: "field.Int", Type: schema.FieldTypeInt},
	{Builder: "field.Int64", Type: schema.FieldTypeInt64},
	{Builder: "field.Float", Type: schema.FieldTypeFloat},
	{Builder: "field.Time", Type: schema.FieldTypeTime},
	{Builder: "field.Byte", Type: schema.FieldTypeByte},
	{Builder: "field.JSON", Type: schema.FieldTypeJSON},
}

var tsTypes = map[schema.FieldType]string{
	schema.FieldTypeString: "string",
	schema.FieldTypeBool:   "boolean",
	schema.FieldTypeInt:    "number",
	schema.FieldTypeInt64:  "bigint",
	schema.FieldTypeFloat:  "number",
	schema.FieldTypeTime:   "Timestamp",
	schema.FieldTypeByte:   "Uint8Array",
	schema.FieldTypeJSON:   "string",
}

func typeMappingPage() []byte {
	page := newPage("Type mapping", "One field type, five outputs. Every column except TypeScript is read "+
		"from the generator itself, so the table cannot drift from the code.")

	rows := make([][]string, 0, len(fieldTypes))
	for _, ft := range fieldTypes {
		rows = append(rows, []string{
			code(ft.Builder),
			code(sqlcwrap.GoTypeFor(ft.Type)),
			code(sqlc.SQLTypeFor(schema.SQLite, ft.Type)),
			code(sqlc.SQLTypeFor(schema.PostgreSQL, ft.Type)),
			code(sqlc.SQLTypeFor(schema.MySQL, ft.Type)),
			code(proto.ProtoTypeFor(ft.Type)),
			code(tsTypes[ft.Type]),
		})
	}
	page.Table([]string{"Builder", "Go", "SQLite", "PostgreSQL", "MySQL", "Proto", "TypeScript"}, rows)

	page.Heading(2, "Notes")
	page.Table([]string{"Type", "Note"}, [][]string{
		{code("field.Int"), "int32, not the Go `int`, so it fits a JavaScript number."},
		{code("field.Int64"), "JSON encodes it as a string, in TypeScript it is a `bigint`."},
		{code("field.Bool"), "SQLite has no boolean, the column is an INTEGER."},
		{code("field.Time"), "Proto uses `google.protobuf.Timestamp`, the wrapper converts it."},
		{code("field.JSON"), "Raw json text, it stays a Go `string` on every layer. SQLite adds a `CHECK (json_valid(col))`."},
		{code("field.String"), "MySQL uses VARCHAR, TEXT cannot be indexed without a key length."},
	})

	page.Heading(2, "Optional fields")
	page.Text("`Optional()` makes the column nullable, the Go type becomes a pointer and the proto field " +
		"is marked `optional`.")

	return page.Bytes()
}

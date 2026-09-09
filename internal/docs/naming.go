package docs

import (
	"fmt"

	"github.com/guntisdev/entlite/internal/naming"
)

const namingEntity = "MyUser"

var namingFields = []struct {
	Name    string
	Written string // what the schema author types, the id field is added for them
	TS      string // protobuf-es lowerCamelCase, the one column entlite does not generate
}{
	{Name: "id", Written: "*(auto)*", TS: "id"},
	{Name: "is_active", TS: "isActive"},
	{Name: "display_name", TS: "displayName"},
}

func namingPage() []byte {
	page := newPage("Naming", "One canonical name per concept in the schema. Every generator derives its "+
		"own spelling from it, nothing is passed through verbatim. Every Go column below is read from "+
		"`internal/naming`, the same code the generators call, so the tables cannot drift from it.")

	page.Heading(2, "The rule")
	page.Table([]string{"Concept", "You write", "Why"}, [][]string{
		{"Entity", code(namingEntity), "a Go type, it has to be exported, so PascalCase"},
		{"Field", code(namingFields[1].Name), "underscores mark the word boundaries every other layer needs"},
	})
	page.Text(fmt.Sprintf("Word boundaries are the only thing that survives the trip through five languages. "+
		"snake_case states them, camelCase leaves them to be guessed, and sqlc does not guess: it "+
		"lowercases the whole name and splits on underscores. So `initCount` reaches Go as `%s`, "+
		"not `%s`. The parser rejects anything that is not the spelling above.",
		naming.SqlcGoName("initCount"), naming.ProtocGoName("initCount")))

	page.Heading(2, "One entity, every layer")
	page.Code("go", namingExample)

	page.Heading(3, "Types")
	page.Table([]string{"Layer", "Name"}, [][]string{
		{"SQL table", code(naming.TableName(namingEntity))},
		{"Proto message", code(namingEntity)},
		// the round trip: sqlc has to read the entity name back out of the table name
		{"Go, both layers", code(naming.SqlcGoName(naming.TableName(namingEntity)))},
		{"TypeScript", code(namingEntity)},
	})
	page.Text("The table is snake_case and singular. Everything else keeps the entity name as written.")

	page.Heading(3, "Fields")
	fieldRows := make([][]string, 0, len(namingFields))
	for _, f := range namingFields {
		written := f.Written
		if written == "" {
			written = code(f.Name)
		}
		fieldRows = append(fieldRows, []string{
			written,
			code(f.Name), // the sql column is the name itself
			code(f.Name), // and so is the proto field
			code(naming.SqlcGoName(f.Name)),
			code(naming.ProtocGoName(f.Name)),
			code(f.TS),
		})
	}
	page.Table([]string{"You write", "SQL", "Proto", "Go via sqlc", "Go via protoc", "TypeScript"}, fieldRows)

	page.Heading(2, "Why the two Go layers differ")
	page.Text(fmt.Sprintf("sqlc applies one initialism, a path segment equal to `id` becomes `%s`. protoc "+
		"applies none. So `%s` is `%s` in `gen/db` and `%s` in `gen/pb`, and it stays that way: each is "+
		"idiomatic for the generator that wrote it. The converter bridges the two.",
		naming.SqlcGoName("id"), namingForeignKey,
		naming.SqlcGoName(namingForeignKey), naming.ProtocGoName(namingForeignKey)))
	page.Code("go", fmt.Sprintf("return &pb.SensorReading{\n\t%s: m.%s,\n}",
		naming.ProtocGoName(namingForeignKey), naming.SqlcGoName(namingForeignKey)))

	page.Heading(2, "What holds it in place")
	page.Table([]string{"Check", "Catches"}, [][]string{
		{"Parser", "an entity that is not exported PascalCase, a field that is not lower snake_case"},
		{"Round trip test", "a generated Go name that does not match what sqlc or protoc would write"},
	})

	return page.Bytes()
}

const namingForeignKey = "sensor_id"

const namingExample = `type MyUser struct {
	entlite.Schema
}

func (MyUser) Fields() []entlite.Field {
	return []entlite.Field{
		field.Bool("is_active"),
		field.String("display_name"),
	}
}`

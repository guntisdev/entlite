package docs

import (
	"fmt"

	"github.com/guntisdev/entlite/internal/naming"
)

const namingEntity = "MyUser"

var namingFields = []struct {
	Name    string
	Written string // what you type. entlite adds the id field
	TS      string // protobuf-es name. entlite does not make it, so it is written here
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
		// round trip: sqlc must read the entity name back from the table name
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
			code(f.Name), // sql column is the name itself
			code(f.Name), // proto field too
			code(naming.SqlcGoName(f.Name)),
			code(naming.ProtocGoName(f.Name)),
			code(f.TS),
		})
	}
	page.Table([]string{"You write", "SQL", "Proto", "Go via sqlc", "Go via protoc", "TypeScript"}, fieldRows)

	page.Heading(3, "Queries")
	page.Text("A query name is built from the entity, and the layers hang their own suffixes off it. " +
		"sqlc reads the name out of the `-- name:` line as its Go method, so the name is a Go " +
		"identifier and not a sql one.")

	createQuery := naming.CreateQueryName(namingEntity)
	bulkQuery := naming.CreateBulkQueryName(namingEntity)
	listQuery := naming.ListByQueryName(namingEntity, nil, []string{namingFields[1].Name}, nil)

	page.Table([]string{"You write", "Query name"}, [][]string{
		{code("query.Create()"), code(createQuery)},
		{code("query.CreateBulk()"), code(bulkQuery)},
		{code(`query.ListBy("` + namingFields[1].Name + `")`), code(listQuery)},
		{code("query.Get()"), code(naming.GetByQueryName(namingEntity, []string{namingFields[0].Name}))},
		{code("query.Delete()"), code(naming.DeleteQueryName(namingEntity))},
	})

	page.Heading(3, "What each layer adds")
	page.Table([]string{"Layer", "Adds", "Gives"}, [][]string{
		{"sqlc", code(naming.SuffixParams), code(naming.ParamsName(createQuery)) + ", the argument struct, only when the query takes more than one"},
		{"proto", code(naming.SuffixRequest), code(naming.RequestName(createQuery)) + ", the rpc input"},
		{"proto", code(naming.SuffixResponse), code(naming.ResponseName(listQuery)) + ", for a list, a create returns the entity itself"},
		{"proto", code(naming.SuffixRow), code(naming.RowName(bulkQuery)) + ", one row of a bulk insert"},
		{"proto", code(naming.SuffixService), code(naming.ServiceName(namingEntity)) + ", one service per entity, off the entity and not the query"},
	})
	page.Text("A custom `Name()` may not end with any of those suffixes, or the generated message would " +
		"come out as `ListActiveRequestRequest`.")

	page.Heading(3, "Aggregate columns")
	page.Text("`Sum()`, `Avg()`, `Min()` and `Max()` select a column of their own, named after the " +
		"function and the column it folds. The name is fixed, only the query itself is renamed, with `Name()`.")

	aggregates := []struct{ method, fn string }{{"Sum", "sum"}, {"Avg", "avg"}, {"Min", "min"}, {"Max", "max"}}
	aggregateRows := make([][]string, 0, len(aggregates))
	for _, aggregate := range aggregates {
		column := naming.AggregateColumn(aggregate.fn, namingAggregateField)
		aggregateRows = append(aggregateRows, []string{
			code(fmt.Sprintf("%s(%q)", aggregate.method, namingAggregateField)),
			code(column),
			code(naming.SqlcGoName(column)),
			code(column),
			code(naming.ProtocGoName(column)),
		})
	}
	page.Table([]string{"You write", "SQL", "Go via sqlc", "Proto", "Go via protoc"}, aggregateRows)
	page.Text("A grouped query selects the grouped columns first and the aggregates after them, in chain " +
		"order, and both the row struct and the row message follow that order.")

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

const namingAggregateField = "duration_ms"

const namingExample = `type MyUser struct {
	entlite.Schema
}

func (MyUser) Fields() []entlite.Field {
	return []entlite.Field{
		field.Bool("is_active"),
		field.String("display_name"),
	}
}`

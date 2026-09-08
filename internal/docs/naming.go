package docs

// TODO derive these tables from /internal/naming once it exists, the way
// typeMappingPage reads the generators, so the page cannot drift from the code.

func namingPage() []byte {
	page := newPage("Naming", "One canonical name per concept in the schema. Every generator derives its "+
		"own spelling from it, nothing is passed through verbatim.")

	page.Text("> Status: the convention entlite is moving to. The generators do not follow all of it yet.")

	page.Heading(2, "The rule")
	page.Table([]string{"Concept", "You write", "Why"}, [][]string{
		{"Entity", code("MyUser"), "a Go type, it has to be exported, so PascalCase"},
		{"Field", code("is_active"), "underscores mark the word boundaries every other layer needs"},
	})
	page.Text("Word boundaries are the only thing that survives the trip through five languages. " +
		"snake_case states them, camelCase leaves them to be guessed, and sqlc does not guess: it " +
		"lowercases the whole name and splits on underscores. So `initCount` reaches Go as `Initcount`, " +
		"not `InitCount`. The parser rejects anything that is not the spelling above.")

	page.Heading(2, "One entity, every layer")
	page.Code("go", namingExample)

	page.Heading(3, "Types")
	page.Table([]string{"Layer", "Name"}, [][]string{
		{"SQL table", code("my_user")},
		{"Proto message", code("MyUser")},
		{"Go, both layers", code("MyUser")},
		{"TypeScript", code("MyUser")},
	})
	page.Text("The table is snake_case and singular. Everything else keeps the entity name as written.")

	page.Heading(3, "Fields")
	page.Table([]string{"You write", "SQL", "Proto", "Go via sqlc", "Go via protoc", "TypeScript"}, [][]string{
		{"*(auto)*", code("id"), code("id"), code("ID"), code("Id"), code("id")},
		{code("is_active"), code("is_active"), code("is_active"), code("IsActive"), code("IsActive"), code("isActive")},
		{code("display_name"), code("display_name"), code("display_name"), code("DisplayName"), code("DisplayName"), code("displayName")},
	})

	page.Heading(3, "Queries")
	page.Table([]string{"You write", "SQL name", "Go method", "rpc"}, [][]string{
		{code(`Create()`), code("CreateMyUser"), code("CreateMyUser(ctx, arg CreateMyUserParams)"), code("CreateMyUser")},
		{code(`Get()`), code("GetMyUser"), code("GetMyUser(ctx, id int32)"), code("GetMyUser")},
		{code(`GetBy("display_name")`), code("GetMyUserByDisplayName"), code("GetMyUserByDisplayName(ctx, displayName string)"), code("GetMyUserByDisplayName")},
		{code(`ListBy("is_active")`), code("ListMyUsersByIsActive"), code("ListMyUsersByIsActive(ctx, isActive bool)"), code("ListMyUsersByIsActive")},
	})
	page.Text("An unsuffixed `Get`, `Update` or `Delete` keys on the primary key, a suffix names the " +
		"fields it looks up instead. Lists and batches say the resource in plural. The `-- name:` in " +
		"the sql file is PascalCase because sqlc uses it verbatim as the Go method name, it is a Go " +
		"identifier and not a sql one.")

	page.Heading(3, "Messages")
	page.Table([]string{"Message", "Holds"}, [][]string{
		{code("MyUser"), "`id`, `is_active`, `display_name`"},
		{code("CreateMyUserRequest"), "`is_active`, `display_name`, the database assigns the id"},
		{code("ListMyUsersByIsActiveRequest"), "`is_active`"},
		{code("ListMyUsersByIsActiveResponse"), code("repeated MyUser my_users = 1")},
	})
	page.Text("A list response names its field after the resource, not `rows`. `Get`, `Create` and " +
		"`Update` return the resource itself, `Delete` returns `google.protobuf.Empty`.")

	page.Heading(2, "Why the two Go layers differ")
	page.Text("sqlc applies one initialism, a path segment equal to `id` becomes `ID`. protoc applies " +
		"none. So `sensor_id` is `SensorID` in `gen/db` and `SensorId` in `gen/pb`, and it stays that " +
		"way: each is idiomatic for the generator that wrote it. The converter bridges the two.")
	page.Code("go", "return &pb.Reading{\n\tSensorId: m.SensorID,\n}")

	page.Heading(2, "What holds it in place")
	page.Table([]string{"Check", "Catches"}, [][]string{
		{"Parser", "an entity that is not exported PascalCase, a field that is not lower snake_case"},
		{"Round trip test", "a generated Go name that does not match what sqlc or protoc would write"},
		{code("buf lint"), "a proto message or field that drifts from the convention"},
	})

	return page.Bytes()
}

const namingExample = `type MyUser struct {
	entlite.Schema
}

func (MyUser) Fields() []entlite.Field {
	return []entlite.Field{
		field.Bool("is_active"),
		field.String("display_name"),
	}
}

func (MyUser) Queries() []entlite.Query {
	return []entlite.Query{
		query.Create(),
		query.Get(),
		query.GetBy("display_name"),
		query.ListBy("is_active"),
	}
}`

package docs

// dslPackages are the packages the reference pages are built from.
var dslPackages = []struct{ Name, Dir string }{
	{Name: "entlite", Dir: "pkg/entlite"},
	{Name: "field", Dir: "pkg/entlite/field"},
	{Name: "query", Dir: "pkg/entlite/query"},
	{Name: "filter", Dir: "pkg/entlite/filter"},
	{Name: "index", Dir: "pkg/entlite/index"},
}

// reference writes the api reference pages, built from the entlite packages.
func reference(p *Pipeline) error {
	api := map[string]apiPackage{}
	for _, pkg := range dslPackages {
		parsed, err := parseAPI(p.Abs(pkg.Dir))
		if err != nil {
			return err
		}
		api[pkg.Name] = parsed
	}

	cli, err := cliPage(p)
	if err != nil {
		return err
	}

	p.Write(p.Out("reference", "entity.md"), entityPage(api["entlite"]))
	p.Write(p.Out("reference", "fields.md"), fieldsPage(api["field"]))
	p.Write(p.Out("reference", "queries.md"), queriesPage(api["query"]))
	p.Write(p.Out("reference", "filters.md"), filtersPage(api["filter"]))
	p.Write(p.Out("reference", "indexes.md"), indexesPage(api["index"]))
	p.Write(p.Out("reference", "contracts.md"), contractsPage(api["entlite"]))
	p.Write(p.Out("reference", "type-mapping.md"), typeMappingPage())
	p.Write(p.Out("reference", "cli.md"), cli)

	return nil
}

var entityMethods = []struct{ Name, Result, Required, Doc string }{
	{Name: "Contracts", Result: "[]entlite.Contract", Required: "yes", Doc: "The layers the entity is generated for, sqlc, proto or both."},
	{Name: "Fields", Result: "[]entlite.Field", Required: "no", Doc: "The columns of the entity, see the fields page."},
	{Name: "Queries", Result: "[]entlite.Query", Required: "no", Doc: "The queries generated for the entity, see the queries page."},
	{Name: "Indexes", Result: "[]entlite.Index", Required: "no", Doc: "The indexes of the table, see the indexes page."},
}

func entityPage(pkg apiPackage) []byte {
	page := newPage("Entity", "An entity is a struct in `ent/schema` that embeds `entlite.Schema`. "+
		"The methods below are read from the source, they are never executed.")

	page.Code("go", `package schema

type User struct {
	entlite.Schema
}

func (User) Contracts() []entlite.Contract { ... }
func (User) Fields() []entlite.Field       { ... }
func (User) Queries() []entlite.Query      { ... }
func (User) Indexes() []entlite.Index      { ... }`)

	rows := make([][]string, 0, len(entityMethods))
	for _, m := range entityMethods {
		rows = append(rows, []string{code(m.Name + "() " + m.Result), m.Required, m.Doc})
	}
	page.Heading(2, "Methods")
	page.Table([]string{"Method", "Required", "Description"}, rows)
	page.Text("Only `Contracts()` is enforced, generation stops when it is missing or empty. " +
		"An entity without `Fields()` still gets its `id` column.")

	page.Heading(2, "Marker interfaces")
	page.Text("Each list holds one of these, the builder packages return them.")
	markers := []string{"Field", "Query", "Index", "Contract"}
	markerRows := make([][]string, 0, len(markers))
	for _, name := range markers {
		if t, ok := pkg.Type(name); ok {
			markerRows = append(markerRows, []string{code("entlite." + t.Name), t.Doc})
		}
	}
	page.Table([]string{"Interface", "Description"}, markerRows)

	page.Heading(2, "The id field")
	page.Text("Every entity gets an `id` field, it is added by the parser when the schema does not declare one. " +
		"An `index.Primary(...)` declares the primary key itself, so the generated `id` column is dropped and " +
		"`query.Get()`, `query.Update()` and `query.Delete()` key on the index columns instead. " +
		"A declared `id` field is kept, as a plain column.")

	return page.Bytes()
}

func fieldsPage(pkg apiPackage) []byte {
	page := newPage("Fields", pkg.Doc)

	builders := fieldBuilders(pkg)
	options := optionColumns(builders)

	page.Heading(2, "Matrix")
	page.Text("Which option each field type accepts.")

	headers := append([]string{"Field"}, options...)
	rows := make([][]string, 0, len(builders))
	for _, b := range builders {
		row := []string{code("field." + b.ctor.Name)}
		for _, option := range options {
			row = append(row, mark(b.typ.HasMethod(option)))
		}
		rows = append(rows, row)
	}
	page.Table(headers, rows)

	for _, b := range builders {
		page.Heading(2, b.ctor.Name)
		page.Text(code("field." + b.ctor.Signature))
		page.Text(b.ctor.Doc)
		page.MethodTable(b.typ.Methods)
	}

	return page.Bytes()
}

type builder struct {
	ctor apiFunc
	typ  apiType
}

func fieldBuilders(pkg apiPackage) []builder {
	var builders []builder
	for _, ctor := range pkg.FuncsReturning("FieldBuilder") {
		typ, ok := pkg.Type(ctor.Result)
		if !ok {
			continue
		}
		builders = append(builders, builder{ctor: ctor, typ: typ})
	}

	return builders
}

func optionColumns(builders []builder) []string {
	var options []string
	seen := map[string]bool{}

	for _, b := range builders {
		for _, name := range b.typ.MethodNames() {
			if seen[name] {
				continue
			}
			seen[name] = true
			options = append(options, name)
		}
	}

	return options
}

// mark renders a matrix cell.
func mark(on bool) string {
	if on {
		return "yes"
	}

	return "–"
}

func queriesPage(pkg apiPackage) []byte {
	page := newPage("Queries", pkg.Doc)

	page.Heading(2, "Builders")
	page.FuncTable("Builder", "query", pkg.Funcs)

	for _, name := range []string{"QueryOperations", "ListByOperations"} {
		typ, ok := pkg.Type(name)
		if !ok {
			continue
		}
		page.Heading(2, name)
		page.Text(typ.Doc)
		page.MethodTable(typ.Methods)
	}

	page.Heading(2, "Query types")
	page.Text("The kind stored on a parsed query, it shows up in the generated method names.")
	page.ConstTable("query", pkg.Consts)

	return page.Bytes()
}

func filtersPage(pkg apiPackage) []byte {
	page := newPage("Filters", pkg.Doc)

	page.Heading(2, "Builders")
	page.FuncTable("Builder", "filter", pkg.Funcs)
	page.Text("A plain field name given to `query.ListBy` becomes `filter.Eq`, so " +
		"`query.ListBy(\"org_id\")` and `query.ListBy(filter.Eq(\"org_id\"))` are the same.")

	for _, ctor := range pkg.Funcs {
		typ, ok := pkg.Type(ctor.Result)
		if !ok {
			continue
		}
		page.Heading(2, ctor.Name)
		page.Text(code("filter." + ctor.Signature))
		page.Text(ctor.Doc)
		page.MethodTable(typ.Methods)
	}

	return page.Bytes()
}

func indexesPage(pkg apiPackage) []byte {
	page := newPage("Indexes", pkg.Doc)

	page.Heading(2, "Builders")
	page.FuncTable("Builder", "index", pkg.Funcs)

	for _, name := range []string{"IndexOperations"} {
		typ, ok := pkg.Type(name)
		if !ok {
			continue
		}
		page.Heading(2, name)
		page.Text(typ.Doc)
		page.MethodTable(typ.Methods)
	}

	page.Heading(2, "Index types")
	page.ConstTable("index", pkg.Consts)

	return page.Bytes()
}

func contractsPage(pkg apiPackage) []byte {
	page := newPage("Contracts", "A contract says which layer a field, a query or a whole entity belongs to. "+
		"`SQLC()` is the database side, `PROTO()` is the api side.")

	page.Heading(2, "Layers")
	page.FuncTable("Layer", "entlite", pkg.FuncsReturning("Layer"))

	if typ, ok := pkg.Type("Layer"); ok {
		page.Heading(2, "Direction")
		page.Text("Without a direction the field or query is readable and writable.")
		page.MethodTable(typ.Methods)
	}

	page.Heading(2, "Where they are used")
	page.Table([]string{"Place", "Effect"}, [][]string{
		{code("Contracts() []entlite.Contract"), "On the entity, it picks the layers the entity is generated for."},
		{code(".Contracts(...)") + " on a field", "Keeps the field out of the layers that are not listed."},
		{code(".Contracts(...)") + " on a query", "Keeps the query out of the layers that are not listed."},
	})
	page.Text("A field with no `Contracts(...)` follows the entity. A field that lists only `entlite.PROTO()` " +
		"is virtual, it has no column and it is filled by hand.")

	return page.Bytes()
}

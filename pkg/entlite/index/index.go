package index

type Type string

const (
	// TypePrimary marks a (compound) primary key.
	TypePrimary Type = "primary"
	// TypeIndex is a regular secondary index (optionally unique).
	TypeIndex Type = "index"
)

// Column is a single indexed column together with its sort direction.
type Column struct {
	name string
	desc bool // false = ASC (default), true = DESC
}

func (c Column) GetName() string { return c.name }
func (c Column) IsDesc() bool    { return c.desc }

type IndexBuilder interface {
	Index()
}

// IndexOperations exposes the fluent modifiers available on a Fields() index.
type IndexOperations interface {
	IndexBuilder
	Unique() IndexOperations
	// Name overrides the auto-generated index name
	Name(name string) IndexOperations
	Asc(field string) IndexOperations
	Desc(field string) IndexOperations
}

type Index struct {
	typeName Type
	columns  []Column
	unique   bool
	name     string
}

// marker method for sealed interface
func (Index) Index() {}

func Primary(fields ...string) IndexBuilder {
	return Index{typeName: TypePrimary, columns: columnsFromFields(fields)}
}

func Fields(fields ...string) IndexOperations {
	return Index{typeName: TypeIndex, columns: columnsFromFields(fields)}
}

func columnsFromFields(fields []string) []Column {
	cols := make([]Column, len(fields))
	for i, f := range fields {
		cols[i] = Column{name: f}
	}
	return cols
}

// Unique turns the index into a unique constraint.
func (i Index) Unique() IndexOperations {
	i.unique = true
	return i
}

// Name overrides the auto-generated index name.
func (i Index) Name(name string) IndexOperations {
	i.name = name
	return i
}

// Asc appends a column sorted ascending.
func (i Index) Asc(field string) IndexOperations {
	i.columns = append(i.columns, Column{name: field, desc: false})
	return i
}

// Desc appends a column sorted descending.
func (i Index) Desc(field string) IndexOperations {
	i.columns = append(i.columns, Column{name: field, desc: true})
	return i
}

func (i Index) GetType() Type {
	return i.typeName
}

func (i Index) GetColumns() []Column {
	return i.columns
}

// GetFields returns just the column names, in order.
func (i Index) GetFields() []string {
	fields := make([]string, len(i.columns))
	for idx, c := range i.columns {
		fields[idx] = c.name
	}
	return fields
}

func (i Index) IsUnique() bool {
	return i.unique
}

func (i Index) GetName() string {
	return i.name
}

// Package index holds the index builders used in an entity schema.
package index

// Type tells which kind of index is described.
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

// GetName returns the column name.
func (c Column) GetName() string { return c.name }

// IsDesc reports if the column is sorted descending.
func (c Column) IsDesc() bool { return c.desc }

// IndexBuilder is implemented by every index.
type IndexBuilder interface {
	Index()
}

// IndexOperations exposes the fluent modifiers available on a Fields() index.
type IndexOperations interface {
	IndexBuilder
	// Unique turns the index into a unique constraint.
	Unique() IndexOperations
	// Name overrides the auto-generated index name
	Name(name string) IndexOperations
	// Asc appends a column sorted ascending.
	Asc(field string) IndexOperations
	// Desc appends a column sorted descending.
	Desc(field string) IndexOperations
}

// Index holds the state of one index.
type Index struct {
	typeName Type
	columns  []Column
	unique   bool
	name     string
}

// marker method for sealed interface
func (Index) Index() {}

// Primary declares the primary key over the given fields. It replaces the generated
// id column, which is left out of the table.
func Primary(fields ...string) IndexBuilder {
	return Index{typeName: TypePrimary, columns: columnsFromFields(fields)}
}

// Fields declares a secondary index over the given fields.
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

// GetType returns the index kind, primary or index.
func (i Index) GetType() Type {
	return i.typeName
}

// GetColumns returns the columns with their sort direction, in order.
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

// IsUnique reports if the index is a unique constraint.
func (i Index) IsUnique() bool {
	return i.unique
}

// GetName returns the custom index name, or "" when auto-generated.
func (i Index) GetName() string {
	return i.name
}

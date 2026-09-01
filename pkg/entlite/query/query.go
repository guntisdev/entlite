// Package query holds the query builders used in an entity schema.
package query

import (
	"github.com/guntisdev/entlite/pkg/entlite"
	"github.com/guntisdev/entlite/pkg/entlite/filter"
)

// Type tells which kind of query is described.
type Type string

const (
	// TypeDefaultCRUD expands to the common create, read, update and delete queries.
	TypeDefaultCRUD Type = "default_crud"
	// TypeCreate inserts one row.
	TypeCreate Type = "create"
	// TypeCreateBulk inserts many rows in one call.
	TypeCreateBulk Type = "create_bulk"
	// TypeGet reads one row by primary key.
	TypeGet Type = "get"
	// TypeUpdate updates one row by primary key.
	TypeUpdate Type = "update"
	// TypeDelete deletes one row by primary key.
	TypeDelete Type = "delete"
	// TypeDeleteAll deletes every row of the table.
	TypeDeleteAll Type = "delete_all"
	// TypeListAll reads every row of the table.
	TypeListAll Type = "list_all"
	// TypeGetBy reads one row by the given fields.
	TypeGetBy Type = "get_by"
	// TypeListBy reads many rows by the given filters.
	TypeListBy Type = "list_by"
)

// QueryBuilder is implemented by every query.
type QueryBuilder interface {
	Query()
}

// QueryOperations exposes the modifiers available on a plain query.
type QueryOperations interface {
	QueryBuilder
	// Name overrides the auto-generated query/method name
	Name(name string) QueryOperations
	// Contracts limits the query to the given layers, sqlc or proto.
	Contracts(contracts ...entlite.Layer) QueryOperations
}

// ListByOperations exposes the modifiers available on a ListBy query.
type ListByOperations interface {
	QueryBuilder
	// Count also returns the number of matching rows.
	Count() ListByOperations
	// OrderBy sorts the result by the given field.
	OrderBy(field string) ListByOperations
	// Name overrides the auto-generated query/method name
	Name(name string) ListByOperations
	// Contracts limits the query to the given layers, sqlc or proto.
	Contracts(contracts ...entlite.Layer) ListByOperations
}

// Query holds the state of one query.
type Query struct {
	typeName  Type
	fields    []string        // For GetBy: list of field name strings
	filters   []filter.Filter // For ListBy: list of filters
	count     bool            // For ListBy: whether to count
	orderBy   string          // For ListBy: order by field
	name      string          // Custom query name
	contracts []entlite.Layer
}

// marker method for sealed interface
func (Query) Query() {}

// Name overrides the auto-generated query/method name
func (q Query) Name(name string) QueryOperations {
	q.name = name
	return q
}

// Contracts limits the query to the given layers, sqlc or proto.
func (q Query) Contracts(contracts ...entlite.Layer) QueryOperations {
	q.contracts = contracts
	return q
}

type listByQuery struct {
	base Query
}

// marker method for sealed interface
func (listByQuery) Query() {}

// Name overrides the auto-generated query/method name
func (q listByQuery) Name(name string) ListByOperations {
	q.base.name = name
	return q
}

// Contracts limits the query to the given layers, sqlc or proto.
func (q listByQuery) Contracts(contracts ...entlite.Layer) ListByOperations {
	q.base.contracts = contracts
	return q
}

// Count adds a COUNT operation to the ListBy query
func (q listByQuery) Count() ListByOperations {
	q.base.count = true
	return q
}

// OrderBy adds ordering to the ListBy query
func (q listByQuery) OrderBy(field string) ListByOperations {
	q.base.orderBy = field
	return q
}

// GetBy gets a record by one or more fields, e.g. GetBy("id") or GetBy("org_id", "email")
func GetBy(fields ...string) QueryOperations {
	return Query{typeName: TypeGetBy, fields: fields}
}

// DefaultCRUD expands to several queries, so it cannot be named
func DefaultCRUD() QueryBuilder {
	return Query{typeName: TypeDefaultCRUD}
}

// Create inserts one record.
func Create() QueryOperations {
	return Query{typeName: TypeCreate}
}

// CreateBulk inserts many records in one call.
func CreateBulk() QueryOperations {
	return Query{typeName: TypeCreateBulk}
}

// Get reads one record by primary key.
func Get() QueryOperations {
	return Query{typeName: TypeGet}
}

// Update updates one record by primary key.
func Update() QueryOperations {
	return Query{typeName: TypeUpdate}
}

// Delete deletes one record by primary key.
func Delete() QueryOperations {
	return Query{typeName: TypeDelete}
}

// DeleteAll deletes every record of the table.
func DeleteAll() QueryOperations {
	return Query{typeName: TypeDeleteAll}
}

// ListAll reads every record of the table.
func ListAll() QueryOperations {
	return Query{typeName: TypeListAll}
}

// ListBy lists records with filters. Takes field names, which default to Eq, or Filter
// values, e.g. ListBy("org_id") or ListBy(filter.Range("age"), filter.Search("name"))
func ListBy(args ...interface{}) ListByOperations {
	q := Query{typeName: TypeListBy}

	for _, arg := range args {
		switch v := arg.(type) {
		case string:
			// String field names default to Eq filter
			q.filters = append(q.filters, filter.Eq(v))
		case filter.Filter:
			q.filters = append(q.filters, v)
		}
	}

	return listByQuery{base: q}
}

// GetType returns the query kind.
func (q Query) GetType() Type {
	return q.typeName
}

// GetFields returns the fields of a GetBy query.
func (q Query) GetFields() []string {
	return q.fields
}

// GetFilters returns the filters of a ListBy query.
func (q Query) GetFilters() []filter.Filter {
	return q.filters
}

// HasCount reports if a ListBy query also returns a count.
func (q Query) HasCount() bool {
	return q.count
}

// GetOrderBy returns the order by field, or "" when there is none.
func (q Query) GetOrderBy() string {
	return q.orderBy
}

// GetName returns the custom query name, or "" when auto-generated.
func (q Query) GetName() string {
	return q.name
}

// GetContracts returns the layers the query belongs to.
func (q Query) GetContracts() []entlite.Layer {
	return q.contracts
}

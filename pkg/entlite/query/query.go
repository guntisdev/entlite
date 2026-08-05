package query

import "github.com/guntisdev/entlite/pkg/entlite/filter"

type Type string

const (
	TypeDefaultCRUD Type = "default_crud"
	TypeCreate      Type = "create"
	TypeCreateBulk  Type = "create_bulk"
	TypeGet         Type = "get"
	TypeUpdate      Type = "update"
	TypeDelete      Type = "delete"
	TypeDeleteAll   Type = "delete_all"
	TypeListAll     Type = "list_all"
	TypeGetBy       Type = "get_by"
	TypeListBy      Type = "list_by"
)

type QueryBuilder interface {
	Query()
}

type QueryOperations interface {
	QueryBuilder
	// Name overrides the auto-generated query/method name
	Name(name string) QueryOperations
}

type ListByOperations interface {
	QueryBuilder
	Count() ListByOperations
	OrderBy(field string) ListByOperations
	// Name overrides the auto-generated query/method name
	Name(name string) ListByOperations
}

type Query struct {
	typeName Type
	fields   []string        // For GetBy: list of field name strings
	filters  []filter.Filter // For ListBy: list of filters
	count    bool            // For ListBy: whether to count
	orderBy  string          // For ListBy: order by field
	name     string          // Custom query name
}

// marker method for sealed interface
func (Query) Query() {}

func (q Query) Name(name string) QueryOperations {
	q.name = name
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

// GetBy creates a query to get a record by one or more fields
// Example: GetBy("id") or GetBy("org_id", "email")
func GetBy(fields ...string) QueryOperations {
	return Query{typeName: TypeGetBy, fields: fields}
}

// DefaultCRUD expands to several queries, so it cannot be named
func DefaultCRUD() QueryBuilder {
	return Query{typeName: TypeDefaultCRUD}
}

func Create() QueryOperations {
	return Query{typeName: TypeCreate}
}

func CreateBulk() QueryOperations {
	return Query{typeName: TypeCreateBulk}
}

func Get() QueryOperations {
	return Query{typeName: TypeGet}
}

func Update() QueryOperations {
	return Query{typeName: TypeUpdate}
}

func Delete() QueryOperations {
	return Query{typeName: TypeDelete}
}

func DeleteAll() QueryOperations {
	return Query{typeName: TypeDeleteAll}
}

func ListAll() QueryOperations {
	return Query{typeName: TypeListAll}
}

// ListBy creates a query to list records with filters
// Can accept either string field names (defaulting to Eq filter) or Filter objects
// Example: ListBy("org_id") or ListBy(filter.Range("age"), filter.Search("name"))
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

func (q Query) GetType() Type {
	return q.typeName
}

func (q Query) GetFields() []string {
	return q.fields
}

func (q Query) GetFilters() []filter.Filter {
	return q.filters
}

func (q Query) HasCount() bool {
	return q.count
}

func (q Query) GetOrderBy() string {
	return q.orderBy
}

// GetName returns the custom query name, or "" when auto-generated.
func (q Query) GetName() string {
	return q.name
}

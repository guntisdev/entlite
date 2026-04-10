package query

import "github.com/guntisdev/entlite/pkg/entlite/filter"

type Type string

const (
	TypeDefaultCRUD Type = "default_crud"
	TypeCreate      Type = "create"
	TypeGet         Type = "get"
	TypeUpdate      Type = "update"
	TypeDelete      Type = "delete"
	TypeList        Type = "list"
	TypeGetBy       Type = "get_by"
	TypeListBy      Type = "list_by"
)

type QueryBuilder interface {
	Query()
}

type ListByOperations interface {
	Query()
	Count() ListByOperations
	OrderBy(field string) ListByOperations
}

type Query struct {
	typeName Type
	fields   []string        // For GetBy: list of field name strings
	filters  []filter.Filter // For ListBy: list of filters
	count    bool            // For ListBy: whether to count
	orderBy  string          // For ListBy: order by field
}

// marker method for sealed interface
func (Query) Query() {}

// Count adds a COUNT operation to the ListBy query
func (q Query) Count() ListByOperations {
	q.count = true
	return q
}

// OrderBy adds ordering to the ListBy query
func (q Query) OrderBy(field string) ListByOperations {
	q.orderBy = field
	return q
}

// GetBy creates a query to get a record by one or more fields
// Example: GetBy("id") or GetBy("org_id", "email")
func GetBy(fields ...string) QueryBuilder {
	return Query{typeName: TypeGetBy, fields: fields}
}

func DefaultCRUD() QueryBuilder {
	return Query{typeName: TypeDefaultCRUD}
}

func Create() QueryBuilder {
	return Query{typeName: TypeCreate}
}

func Get() QueryBuilder {
	return Query{typeName: TypeGet}
}

func Update() QueryBuilder {
	return Query{typeName: TypeUpdate}
}

func Delete() QueryBuilder {
	return Query{typeName: TypeDelete}
}

func List() QueryBuilder {
	return Query{typeName: TypeList}
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

	return q
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

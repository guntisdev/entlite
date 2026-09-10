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

// CreateOperations exposes the modifiers available on a Create or CreateBulk query.
type CreateOperations interface {
	QueryBuilder
	// Upsert updates the existing row when the given fields collide. The fields must be
	// a primary key or a unique constraint, no fields means the primary key.
	Upsert(fields ...string) CreateOperations
	// Ignore keeps the existing row instead of updating it. Needs an Upsert.
	Ignore() CreateOperations
	// Name overrides the auto-generated query/method name
	Name(name string) CreateOperations
	// Contracts limits the query to the given layers, sqlc or proto.
	Contracts(contracts ...entlite.Layer) CreateOperations
}

// ListAllOperations exposes the modifiers available on a ListAll query.
type ListAllOperations interface {
	QueryBuilder
	// Count also returns how many rows match, counted before Limit. An empty page reports 0.
	Count() ListAllOperations
	// Distinct returns the deduplicated values of the given columns instead of whole
	// rows. Every column is part of the key, so sorting is limited to them.
	Distinct(fields ...string) ListAllOperations
	// GroupBy returns one row per distinct combination of the given columns, an
	// aggregate folds each group. Sorting is limited to the grouped columns.
	GroupBy(fields ...string) ListAllOperations
	// Sum adds up the given column, per group or over the whole table without a GroupBy.
	Sum(field string) ListAllOperations
	// Avg averages the given column, per group or over the whole table without a GroupBy.
	Avg(field string) ListAllOperations
	// Min takes the smallest value of the given column, per group or over the whole table.
	Min(field string) ListAllOperations
	// Max takes the largest value of the given column, per group or over the whole table.
	Max(field string) ListAllOperations
	// Asc appends a sort column, ascending.
	Asc(field string) ListAllOperations
	// Desc appends a sort column, descending.
	Desc(field string) ListAllOperations
	// Limit takes the row count from the caller, Limit(rows) sets it in the query.
	Limit(rows ...int) ListAllOperations
	// Offset asks the caller how many rows to skip. Needs a Limit.
	Offset() ListAllOperations
	// Name overrides the auto-generated query/method name
	Name(name string) ListAllOperations
	// Contracts limits the query to the given layers, sqlc or proto.
	Contracts(contracts ...entlite.Layer) ListAllOperations
}

// ListByOperations exposes the modifiers available on a ListBy query.
type ListByOperations interface {
	QueryBuilder
	// Count also returns how many rows match, counted before Limit. An empty page reports 0.
	Count() ListByOperations
	// Distinct returns the deduplicated values of the given columns instead of whole
	// rows. Every column is part of the key, so sorting is limited to them.
	Distinct(fields ...string) ListByOperations
	// GroupBy returns one row per distinct combination of the given columns, an
	// aggregate folds each group. Sorting is limited to the grouped columns.
	GroupBy(fields ...string) ListByOperations
	// Sum adds up the given column, per group or over the whole table without a GroupBy.
	Sum(field string) ListByOperations
	// Avg averages the given column, per group or over the whole table without a GroupBy.
	Avg(field string) ListByOperations
	// Min takes the smallest value of the given column, per group or over the whole table.
	Min(field string) ListByOperations
	// Max takes the largest value of the given column, per group or over the whole table.
	Max(field string) ListByOperations
	// Asc appends a sort column, ascending.
	Asc(field string) ListByOperations
	// Desc appends a sort column, descending.
	Desc(field string) ListByOperations
	// Limit takes the row count from the caller, Limit(rows) sets it in the query.
	Limit(rows ...int) ListByOperations
	// Offset asks the caller how many rows to skip. Needs a Limit.
	Offset() ListByOperations
	// Name overrides the auto-generated query/method name
	Name(name string) ListByOperations
	// Contracts limits the query to the given layers, sqlc or proto.
	Contracts(contracts ...entlite.Layer) ListByOperations
}

// Func tells which sql aggregate folds a column.
type Func string

const (
	// FuncSum adds up the column values, SUM().
	FuncSum Func = "sum"
	// FuncAvg averages the column values, AVG().
	FuncAvg Func = "avg"
	// FuncMin takes the smallest column value, MIN().
	FuncMin Func = "min"
	// FuncMax takes the largest column value, MAX().
	FuncMax Func = "max"
)

// Aggregate is one aggregate function over one column.
type Aggregate struct {
	fn    Func
	field string
}

// GetFunc returns the aggregate function.
func (a Aggregate) GetFunc() Func { return a.fn }

// GetField returns the aggregated column name.
func (a Aggregate) GetField() string { return a.field }

// OrderColumn is a single sort column together with its direction.
type OrderColumn struct {
	name string
	desc bool // false = ASC (default), true = DESC
}

// GetName returns the column name.
func (c OrderColumn) GetName() string { return c.name }

// IsDesc reports if the column is sorted descending.
func (c OrderColumn) IsDesc() bool { return c.desc }

// Query holds the state of one query.
type Query struct {
	typeName     Type
	fields       []string        // For GetBy: list of field name strings
	filters      []filter.Filter // For ListBy: list of filters
	count        bool            // For list queries: whether to count matching rows
	distinct     []string        // For list queries: the columns selected deduplicated
	groupBy      []string        // For list queries: the columns the rows are grouped by
	aggregates   []Aggregate     // For list queries: the aggregate functions, in chain order
	orderBy      []OrderColumn   // For list queries: sort columns, in order
	hasLimit     bool            // For list queries: whether LIMIT is set
	limit        int             // For list queries: fixed limit, 0 means the caller sets it
	hasOffset    bool            // For list queries: whether OFFSET is set
	upsert       bool            // For create queries: whether ON CONFLICT is set
	upsertFields []string        // For create queries: the conflict target, empty means the primary key
	upsertIgnore bool            // For create queries: keep the existing row instead of updating it
	name         string          // Custom query name
	contracts    []entlite.Layer
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

type createQuery struct {
	base Query
}

// marker method for sealed interface
func (createQuery) Query() {}

// Name overrides the auto-generated query/method name
func (q createQuery) Name(name string) CreateOperations {
	q.base.name = name
	return q
}

// Contracts limits the query to the given layers, sqlc or proto.
func (q createQuery) Contracts(contracts ...entlite.Layer) CreateOperations {
	q.base.contracts = contracts
	return q
}

// Upsert updates the existing row when the given fields collide
func (q createQuery) Upsert(fields ...string) CreateOperations {
	q.base.upsert = true
	q.base.upsertFields = fields
	return q
}

// Ignore keeps the existing row instead of updating it
func (q createQuery) Ignore() CreateOperations {
	q.base.upsertIgnore = true
	return q
}

type listAllQuery struct {
	base Query
}

// marker method for sealed interface
func (listAllQuery) Query() {}

// Name overrides the auto-generated query/method name
func (q listAllQuery) Name(name string) ListAllOperations {
	q.base.name = name
	return q
}

// Contracts limits the query to the given layers, sqlc or proto.
func (q listAllQuery) Contracts(contracts ...entlite.Layer) ListAllOperations {
	q.base.contracts = contracts
	return q
}

// Count adds a COUNT operation to the ListAll query
func (q listAllQuery) Count() ListAllOperations {
	q.base.count = true
	return q
}

// Distinct selects the deduplicated values of the given columns of the ListAll query
func (q listAllQuery) Distinct(fields ...string) ListAllOperations {
	q.base.distinct = fields
	return q
}

// GroupBy groups the rows of the ListAll query by the given columns
func (q listAllQuery) GroupBy(fields ...string) ListAllOperations {
	q.base.groupBy = fields
	return q
}

// Sum adds up the given column of the ListAll query
func (q listAllQuery) Sum(field string) ListAllOperations {
	q.base.addAggregate(FuncSum, field)
	return q
}

// Avg averages the given column of the ListAll query
func (q listAllQuery) Avg(field string) ListAllOperations {
	q.base.addAggregate(FuncAvg, field)
	return q
}

// Min takes the smallest value of the given column of the ListAll query
func (q listAllQuery) Min(field string) ListAllOperations {
	q.base.addAggregate(FuncMin, field)
	return q
}

// Max takes the largest value of the given column of the ListAll query
func (q listAllQuery) Max(field string) ListAllOperations {
	q.base.addAggregate(FuncMax, field)
	return q
}

// Asc appends a sort column to the ListAll query, ascending
func (q listAllQuery) Asc(field string) ListAllOperations {
	q.base.addOrder(field, false)
	return q
}

// Desc appends a sort column to the ListAll query, descending
func (q listAllQuery) Desc(field string) ListAllOperations {
	q.base.addOrder(field, true)
	return q
}

// Limit sets how many rows the ListAll query returns
func (q listAllQuery) Limit(rows ...int) ListAllOperations {
	q.base.setLimit(rows)
	return q
}

// Offset skips rows of the ListAll query, the caller gives the count
func (q listAllQuery) Offset() ListAllOperations {
	q.base.hasOffset = true
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

// Distinct selects the deduplicated values of the given columns of the ListBy query
func (q listByQuery) Distinct(fields ...string) ListByOperations {
	q.base.distinct = fields
	return q
}

// GroupBy groups the rows of the ListBy query by the given columns
func (q listByQuery) GroupBy(fields ...string) ListByOperations {
	q.base.groupBy = fields
	return q
}

// Sum adds up the given column of the ListBy query
func (q listByQuery) Sum(field string) ListByOperations {
	q.base.addAggregate(FuncSum, field)
	return q
}

// Avg averages the given column of the ListBy query
func (q listByQuery) Avg(field string) ListByOperations {
	q.base.addAggregate(FuncAvg, field)
	return q
}

// Min takes the smallest value of the given column of the ListBy query
func (q listByQuery) Min(field string) ListByOperations {
	q.base.addAggregate(FuncMin, field)
	return q
}

// Max takes the largest value of the given column of the ListBy query
func (q listByQuery) Max(field string) ListByOperations {
	q.base.addAggregate(FuncMax, field)
	return q
}

// Asc appends a sort column to the ListBy query, ascending
func (q listByQuery) Asc(field string) ListByOperations {
	q.base.addOrder(field, false)
	return q
}

// Desc appends a sort column to the ListBy query, descending
func (q listByQuery) Desc(field string) ListByOperations {
	q.base.addOrder(field, true)
	return q
}

// Limit sets how many rows the ListBy query returns
func (q listByQuery) Limit(rows ...int) ListByOperations {
	q.base.setLimit(rows)
	return q
}

// Offset skips rows of the ListBy query, the caller gives the count
func (q listByQuery) Offset() ListByOperations {
	q.base.hasOffset = true
	return q
}

// addAggregate appends one aggregate, the chain order is the column order
func (q *Query) addAggregate(fn Func, field string) {
	q.aggregates = append(q.aggregates, Aggregate{fn: fn, field: field})
}

// addOrder appends one sort column, the chain order is the sort order
func (q *Query) addOrder(field string, desc bool) {
	q.orderBy = append(q.orderBy, OrderColumn{name: field, desc: desc})
}

// setLimit marks the limit, a given value keeps it out of the request
func (q *Query) setLimit(rows []int) {
	q.hasLimit = true
	if len(rows) > 0 {
		q.limit = rows[0]
	}
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
func Create() CreateOperations {
	return createQuery{base: Query{typeName: TypeCreate}}
}

// CreateBulk inserts many records in one call.
func CreateBulk() CreateOperations {
	return createQuery{base: Query{typeName: TypeCreateBulk}}
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
func ListAll() ListAllOperations {
	return listAllQuery{base: Query{typeName: TypeListAll}}
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

// HasCount reports if a list query also returns a count.
func (q Query) HasCount() bool {
	return q.count
}

// GetDistinct returns the deduplicated columns, or nil when the query returns rows.
func (q Query) GetDistinct() []string {
	return q.distinct
}

// GetGroupBy returns the grouped columns, or nil when the query returns rows.
func (q Query) GetGroupBy() []string {
	return q.groupBy
}

// GetAggregates returns the aggregates in chain order, or nil when there is none.
func (q Query) GetAggregates() []Aggregate {
	return q.aggregates
}

// GetOrderBy returns the sort columns in order, or nil when there is none.
func (q Query) GetOrderBy() []OrderColumn {
	return q.orderBy
}

// HasLimit reports if the query limits the returned rows.
func (q Query) HasLimit() bool {
	return q.hasLimit
}

// GetLimit returns the fixed limit, or 0 when the caller sets it.
func (q Query) GetLimit() int {
	return q.limit
}

// HasOffset reports if the query skips rows.
func (q Query) HasOffset() bool {
	return q.hasOffset
}

// HasUpsert reports if a create query updates the row it collides with.
func (q Query) HasUpsert() bool {
	return q.upsert
}

// GetUpsertFields returns the conflict target, or nil for the primary key.
func (q Query) GetUpsertFields() []string {
	return q.upsertFields
}

// HasUpsertIgnore reports if an upsert keeps the existing row instead of updating it.
func (q Query) HasUpsertIgnore() bool {
	return q.upsertIgnore
}

// GetName returns the custom query name, or "" when auto-generated.
func (q Query) GetName() string {
	return q.name
}

// GetContracts returns the layers the query belongs to.
func (q Query) GetContracts() []entlite.Layer {
	return q.contracts
}

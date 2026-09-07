package schema

import (
	"strings"
)

type Schema struct {
	Entities []Entity
}

func (e Entity) HasSQLC() bool {
	for _, c := range e.Contracts {
		if c.Type == ContractSQLC {
			return true
		}
	}
	return false
}

func (e Entity) HasPROTO() bool {
	for _, c := range e.Contracts {
		if c.Type == ContractPROTO {
			return true
		}
	}
	return false
}

func (e Entity) IsFieldVirtual(field Field) bool {
	return e.HasSQLC() && field.IsVirtual()
}

func (e Entity) CanFieldWrite(field Field) bool {
	if e.HasPROTO() {
		return field.CanApiWrite()
	}

	return field.CanDbWrite()
}

func (e Entity) CanFieldRead(field Field) bool {
	if e.HasPROTO() {
		return field.CanApiRead()
	}

	return field.CanDbRead()
}

func (e Entity) GetContract(contractType ContractType) (Contract, bool) {
	for _, c := range e.Contracts {
		if c.Type == contractType {
			return c, true
		}
	}
	return Contract{}, false
}

// SQLCQueries returns the queries generated for the sqlc contract
func (e Entity) SQLCQueries() []Query {
	return e.contractQueries(ContractSQLC)
}

// ProtoQueries returns the queries generated for the proto contract
func (e Entity) ProtoQueries() []Query {
	return e.contractQueries(ContractPROTO)
}

func (e Entity) contractQueries(contractType ContractType) []Query {
	contract, ok := e.GetContract(contractType)
	if !ok {
		return nil
	}

	var queries []Query
	for _, query := range e.Queries {
		if !query.HasContract(contractType) {
			continue
		}
		if contract.Access == AccessRead && query.IsWrite() {
			continue
		}
		if contract.Access == AccessWrite && !query.IsWrite() {
			continue
		}
		queries = append(queries, query)
	}

	return queries
}

func FilterSQLC(entities []Entity) []Entity {
	var filtered []Entity
	for _, entity := range entities {
		if entity.HasSQLC() {
			filtered = append(filtered, entity)
		}
	}

	return filtered
}

func FilterPROTO(entities []Entity) []Entity {
	var filtered []Entity
	for _, entity := range entities {
		if entity.HasPROTO() {
			filtered = append(filtered, entity)
		}
	}

	return filtered
}

type Entity struct {
	Name      string
	Comment   string
	Fields    []Field
	Contracts []Contract
	Queries   []Query
	Indexes   []Index
}

func (e Entity) GetIdField() Field {
	for _, field := range e.Fields {
		if field.IsID() {
			return field
		}
	}

	panic("No id field detected")
}

func (e Entity) HasIdField() bool {
	for _, field := range e.Fields {
		if field.IsID() {
			return true
		}
	}

	return false
}

func (e Entity) PrimaryIndex() (Index, bool) {
	for _, idx := range e.Indexes {
		if idx.Type == IndexPrimary {
			return idx, true
		}
	}

	return Index{}, false
}

func (e Entity) PrimaryKeyFields() []Field {
	idx, ok := e.PrimaryIndex()
	if !ok {
		if !e.HasIdField() {
			return nil
		}
		return []Field{e.GetIdField()}
	}

	fields := make([]Field, 0, len(idx.Columns))
	for _, column := range idx.Columns {
		if field, found := e.GetFieldByName(column.Name); found {
			fields = append(fields, field)
		}
	}

	return fields
}

// UpsertTarget returns the conflict target columns of an upsert query. A query that
// names no fields targets the primary key.
func (e Entity) UpsertTarget(query Query) []string {
	if len(query.UpsertFields) > 0 {
		return query.UpsertFields
	}

	keyFields := e.PrimaryKeyFields()
	columns := make([]string, 0, len(keyFields))
	for _, field := range keyFields {
		columns = append(columns, field.Name)
	}

	return columns
}

// HasUniqueConstraint reports if the columns are the primary key, a Unique() field or
// the columns of a Unique() index. Only such a set can be an upsert conflict target.
func (e Entity) HasUniqueConstraint(columns []string) bool {
	if len(columns) == 0 {
		return false
	}

	keyFields := e.PrimaryKeyFields()
	keyColumns := make([]string, 0, len(keyFields))
	for _, field := range keyFields {
		keyColumns = append(keyColumns, field.Name)
	}
	if sameColumns(columns, keyColumns) {
		return true
	}

	if len(columns) == 1 {
		if field, ok := e.GetFieldByName(columns[0]); ok && field.Unique {
			return true
		}
	}

	for _, idx := range e.Indexes {
		if idx.Unique && sameColumns(columns, idx.FieldNames()) {
			return true
		}
	}

	return false
}

func (e Entity) UpsertSetFields(target []string) []Field {
	var fields []Field

	for _, field := range e.Fields {
		if !field.CanDbWrite() {
			continue
		}
		if field.IsID() || e.IsPrimaryKeyField(field) || field.Immutable {
			continue
		}
		if containsColumn(target, field.Name) {
			continue
		}
		fields = append(fields, field)
	}

	return fields
}

func (e Entity) CollidableConstraints() [][]string {
	var constraints [][]string

	// a compound primary key is made of schema fields, so the insert sends it
	if idx, ok := e.PrimaryIndex(); ok {
		constraints = append(constraints, idx.FieldNames())
	}

	for _, field := range e.Fields {
		// the generated id field is marked unique, but an insert either leaves it to the
		// database or mints it, so it does not collide
		if field.IsID() || field.IsVirtual() || !field.Unique {
			continue
		}
		constraints = append(constraints, []string{field.Name})
	}

	for _, idx := range e.Indexes {
		if idx.Type == IndexRegular && idx.Unique {
			constraints = append(constraints, idx.FieldNames())
		}
	}

	return dropRedundantConstraints(constraints)
}

func (e Entity) OtherCollidableConstraints(target []string) [][]string {
	var others [][]string

	for _, columns := range e.CollidableConstraints() {
		if sameColumns(columns, target) {
			continue
		}
		others = append(others, columns)
	}

	return others
}

func dropRedundantConstraints(constraints [][]string) [][]string {
	var kept [][]string

	for i, columns := range constraints {
		redundant := false
		for j, other := range constraints {
			if i == j {
				continue
			}
			if sameColumns(columns, other) {
				// keep the first of the duplicates
				if j < i {
					redundant = true
					break
				}
				continue
			}
			if len(other) < len(columns) && containsColumns(columns, other) {
				redundant = true
				break
			}
		}
		if !redundant {
			kept = append(kept, columns)
		}
	}

	return kept
}

func containsColumns(set, subset []string) bool {
	for _, want := range subset {
		if !containsColumn(set, want) {
			return false
		}
	}

	return true
}

func containsColumn(set []string, column string) bool {
	for _, got := range set {
		if strings.EqualFold(got, column) {
			return true
		}
	}

	return false
}

func (e Entity) HasAutoGeneratedKey() bool {
	if _, ok := e.PrimaryIndex(); ok {
		return false
	}
	if !e.HasIdField() {
		return false
	}

	return e.GetIdField().DefaultFunc == nil
}

// sameColumns compares two column lists as sets, a conflict target may name the
// columns of a compound key in any order
func sameColumns(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for _, want := range a {
		found := false
		for _, got := range b {
			if strings.EqualFold(want, got) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

func (e Entity) IsPrimaryKeyField(field Field) bool {
	for _, keyField := range e.PrimaryKeyFields() {
		if strings.EqualFold(keyField.Name, field.Name) {
			return true
		}
	}

	return false
}

func (e Entity) GetFieldByName(name string) (Field, bool) {
	for _, field := range e.Fields {
		if strings.EqualFold(field.Name, name) {
			return field, true
		}
	}

	return Field{}, false
}

type Field struct {
	Name         string
	Type         FieldType
	Primary      bool
	Unique       bool
	DefaultValue any
	DefaultFunc  func() any
	ProtoField   int
	Comment      string
	Contracts    []Contract
	Immutable    bool
	Optional     bool
	Validate     func() any
}

func (f Field) IsID() bool {
	return strings.ToLower(f.Name) == "id"
}

// IsVirtual reports a field that lives only in proto and not in sqlc.
func (f Field) IsVirtual() bool {
	_, ok := f.GetContract(ContractSQLC)
	return !ok
}

// GetContract returns the field contract of the given type
func (f Field) GetContract(contractType ContractType) (Contract, bool) {
	for _, c := range f.Contracts {
		if c.Type == contractType {
			return c, true
		}
	}

	return Contract{}, false
}

func (f Field) CanDbRead() bool {
	return f.canAccess(ContractSQLC, AccessWrite)
}

func (f Field) CanDbWrite() bool {
	return f.canAccess(ContractSQLC, AccessRead)
}

func (f Field) CanApiRead() bool {
	return f.canAccess(ContractPROTO, AccessWrite)
}

func (f Field) CanApiWrite() bool {
	return f.canAccess(ContractPROTO, AccessRead)
}

func (f Field) canAccess(contractType ContractType, blockedBy Access) bool {
	contract, ok := f.GetContract(contractType)
	if !ok {
		return false
	}

	return contract.Access != blockedBy
}

type FieldType string

const (
	FieldTypeString FieldType = "string"
	FieldTypeInt    FieldType = "int32"
	FieldTypeInt64  FieldType = "int64"
	FieldTypeFloat  FieldType = "float64"
	FieldTypeBool   FieldType = "bool"
	FieldTypeTime   FieldType = "time"
	FieldTypeByte   FieldType = "[]byte"
	FieldTypeJSON   FieldType = "json"
)

type Contract struct {
	Type   ContractType
	Access Access
}

// Access narrows what a contract allows. Empty means both read and write.
type Access string

const (
	AccessFull  Access = ""
	AccessRead  Access = "read"
	AccessWrite Access = "write"
)

type ContractType string

const (
	ContractSQLC  ContractType = "sqlc"
	ContractPROTO ContractType = "proto"
)

// TotalSizeColumn is the column Count() adds to a list query, named after google AIP-158
const TotalSizeColumn = "total_size"

type Query struct {
	Type         QueryType
	Fields       []string
	Filters      []QueryFilter
	Count        bool // Count() asks for the number of matching rows
	OrderBy      []OrderColumn
	HasLimit     bool
	Limit        int // fixed row count, 0 means the caller sets it
	HasOffset    bool
	Name         string
	Comment      string
	Contracts    []Contract
	PrimaryKey   bool
	Upsert       bool     // Upsert() updates the row the insert collides with
	UpsertFields []string // the conflict target, empty means the primary key
	UpsertIgnore bool     // Ignore() keeps the existing row instead of updating it
}

type OrderColumn struct {
	Name string
	Desc bool // false = ASC (default), true = DESC
}

func (q Query) LimitFromRequest() bool {
	return q.HasLimit && q.Limit == 0
}

func (q Query) IsList() bool {
	return q.Type == QueryListAll || q.Type == QueryListBy
}

func (q Query) IsCreate() bool {
	return q.Type == QueryCreate || q.Type == QueryCreateBulk
}

// returns the entity's first query of the given type
func (e Entity) QueryByType(queryType QueryType) (Query, bool) {
	for _, query := range e.Queries {
		if query.Type == queryType {
			return query, true
		}
	}

	return Query{}, false
}

// returns the get query keyed by the primary key
func (e Entity) PrimaryKeyGetQuery() (Query, bool) {
	keyFields := e.PrimaryKeyFields()

	for _, query := range e.Queries {
		if query.Type != QueryGetBy || len(query.Fields) != len(keyFields) {
			continue
		}

		if sameFieldNames(query.Fields, keyFields) {
			return query, true
		}
	}

	return Query{}, false
}

func sameFieldNames(names []string, fields []Field) bool {
	for i := range names {
		if !strings.EqualFold(names[i], fields[i].Name) {
			return false
		}
	}

	return true
}

func (q Query) HasContract(contractType ContractType) bool {
	for _, c := range q.Contracts {
		if c.Type == contractType {
			return true
		}
	}

	return false
}

func (q Query) IsWrite() bool {
	switch q.Type {
	case QueryCreate, QueryCreateBulk, QueryUpdate, QueryDelete, QueryDeleteAll:
		return true
	}

	return false
}

type QueryFilter struct {
	Type     QueryFilterType
	Field    string
	Optional bool
}

type QueryFilterType string

const (
	QueryFilterRange  QueryFilterType = "range"
	QueryFilterSearch QueryFilterType = "search"
	QueryFilterEq     QueryFilterType = "eq"
)

type Index struct {
	Type    IndexType
	Columns []IndexColumn
	Unique  bool
	Name    string
}

func (i Index) FieldNames() []string {
	names := make([]string, len(i.Columns))
	for idx, c := range i.Columns {
		names[idx] = c.Name
	}
	return names
}

type IndexColumn struct {
	Name string
	Desc bool // false = ASC (default), true = DESC
}

type IndexType string

const (
	IndexPrimary IndexType = "primary"
	IndexRegular IndexType = "index"
)

type QueryType string

const (
	QueryCreate     QueryType = "create"
	QueryCreateBulk QueryType = "create_bulk"
	QueryUpdate     QueryType = "update"
	QueryDelete     QueryType = "delete"
	QueryDeleteAll  QueryType = "delete_all"
	QueryGetBy      QueryType = "get_by"
	QueryListBy     QueryType = "list_by"
	QueryListAll    QueryType = "list_all"
)

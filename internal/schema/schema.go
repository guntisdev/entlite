package schema

import (
	"strings"

	"github.com/guntisdev/entlite/pkg/entlite/permissions"
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

type Entity struct {
	Name      string
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

func (e Entity) GetFieldByName(name string) (Field, bool) {
	for _, field := range e.Fields {
		if field.Name == name {
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
	Permissions  permissions.Permission
	Immutable    bool
	Optional     bool
	Validate     func() any
}

func (f Field) IsID() bool {
	return strings.ToLower(f.Name) == "id"
}

// IsVirtual reports a field that live only in proto and not in sqlc
func (f Field) IsVirtual() bool {
	return f.Permissions&(permissions.DbRead|permissions.DbWrite) == 0
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
	Type ContractType
}

type ContractType string

const (
	ContractSQLC  ContractType = "sqlc"
	ContractPROTO ContractType = "proto"
)

type Query struct {
	Type    QueryType
	Fields  []string
	Filters []QueryFilter
	Count   bool
	OrderBy string
	Name    string // custom query name; empty means auto-generated
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

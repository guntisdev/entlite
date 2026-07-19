package schema

import (
	"strings"

	"github.com/guntisdev/entlite/pkg/entlite/permissions"
)

type Schema struct {
	Entities []Entity
}

func (e Entity) HasMessage() bool {
	for _, ann := range e.Annotations {
		if ann.Type == AnnotationMessage {
			return true
		}
	}
	return false
}

func (e Entity) HasService() bool {
	for _, ann := range e.Annotations {
		if ann.Type == AnnotationGRPC {
			return true
		}
	}
	return false
}

type Entity struct {
	Name        string
	Fields      []Field
	Annotations []Annotation
	Queries     []Query
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

type FieldType string

const (
	FieldTypeString FieldType = "string"
	FieldTypeInt    FieldType = "int32"
	FieldTypeInt64  FieldType = "int64"
	FieldTypeFloat  FieldType = "float64"
	FieldTypeBool   FieldType = "bool"
	FieldTypeTime   FieldType = "time"
	FieldTypeByte   FieldType = "[]byte"
)

type Annotation struct {
	Type AnnotationType
}

type AnnotationType string

const (
	AnnotationMessage AnnotationType = "message"
	AnnotationGRPC    AnnotationType = "grpc"
)

type Query struct {
	Type    QueryType
	Fields  []string
	Filters []QueryFilter
	Count   bool
	OrderBy string
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

type QueryType string

const (
	QueryCreate    QueryType = "create"
	QueryUpdate    QueryType = "update"
	QueryDelete    QueryType = "delete"
	QueryDeleteAll QueryType = "delete_all"
	QueryGetBy     QueryType = "get_by"
	QueryListBy    QueryType = "list_by"
	QueryListAll   QueryType = "list_all"
)

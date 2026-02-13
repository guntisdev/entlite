package schema

import "strings"

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
		if ann.Type == AnnotationService {
			return true
		}
	}
	return false
}

func (e Entity) GetMethods() []Method {
	for _, ann := range e.Annotations {
		if ann.Type == AnnotationService && len(ann.Methods) > 0 {
			return ann.Methods
		}
	}

	return nil
}

type Entity struct {
	Name        string
	Fields      []Field
	Annotations []Annotation
}

type Field struct {
	Name         string
	Type         FieldType
	Unique       bool
	DefaultValue any
	DefaultFunc  func() any
	ProtoField   int
	Comment      string
	Immutable    bool
	Optional     bool
}

func (f Field) IsID() bool {
	return strings.ToLower(f.Name) == "id"
}

type FieldType string

const (
	FieldTypeString FieldType = "string"
	FieldTypeInt32  FieldType = "int32"
	FieldTypeBool   FieldType = "bool"
	FieldTypeTime   FieldType = "time"
)

type Annotation struct {
	Type    AnnotationType
	Methods []Method
}

type AnnotationType string

const (
	AnnotationMessage AnnotationType = "message"
	AnnotationService AnnotationType = "service"
	AnnotationMethods AnnotationType = "methods"
)

type Method string

const (
	MethodCreate Method = "create"
	MethodGet    Method = "get"
	MethodUpdate Method = "update"
	MethodDelete Method = "delete"
	MethodList   Method = "list"
)

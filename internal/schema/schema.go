package schema

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
	ProtoField   *int
}

type FieldType string

const (
	FieldTypeString FieldType = "string"
	FieldTypeInt32  FieldType = "int32"
	FieldTypeBool   FieldType = "bool"
	// FieldTypeTime   FieldType = "time"
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

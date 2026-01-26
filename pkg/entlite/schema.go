package entlite

type Schema struct{}

type Annotation interface {
	annotation()
}

type Field interface {
	field()
}

//	type Entity interface {
//		EntityName() string
//		EntityFields() []Field
//		EntityAnnotations() []Annotation
//	}
type MessageAnnotation struct{}

func (MessageAnnotation) annotation() {}

func Message() Annotation {
	return MessageAnnotation{}
}

type ServiceAnnotation struct{}

func (ServiceAnnotation) annotation() {}

func Service(annotations ...Annotation) Annotation {
	return ServiceAnnotation{}
}

// CRUD method type
type Method int

const (
	MethodCreate Method = 1 << iota
	MethodGet
	MethodUpdate
	MethodDelete
	MethodList
)

type MethodsAnnotation struct {
	Methods []Method
}

func (MethodsAnnotation) annotation() {}

func Methods(...Method) Annotation {
	var methodList []Method

	return MethodsAnnotation{Methods: methodList}
}

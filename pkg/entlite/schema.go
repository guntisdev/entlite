package entlite

type Schema struct{}

type Annotation interface {
	annotation()
}

type Field interface {
	field()
}

type Entity interface {
	EntityName() string
	EntityFields() []Field
	EntityAnnotations() []Annotation
}

type MessageAnnotation struct{}

func (MessageAnnotation) annotation() {}

func Message() Annotation {
	return MessageAnnotation{}
}

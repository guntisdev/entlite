package entlite

type Schema struct{}

type Annotation interface {
	Annotation()
}

type Field interface {
	Field()
}

type MessageAnnotation struct{}

func (MessageAnnotation) Annotation() {}

func Message() Annotation {
	return MessageAnnotation{}
}

type ServiceAnnotation struct{}

func (ServiceAnnotation) Annotation() {}

func Service(annotations ...Annotation) Annotation {
	return ServiceAnnotation{}
}

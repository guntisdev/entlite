package entlite

type Schema struct{}

type Annotation interface {
	Annotation()
}

type Field interface {
	Field()
}

type Query interface {
	Query()
}

type MessageAnnotation struct{}

func (MessageAnnotation) Annotation() {}

func Message() Annotation {
	return MessageAnnotation{}
}

type GRPCAnnotation struct{}

func (GRPCAnnotation) Annotation() {}

func GRPC(annotations ...Annotation) Annotation {
	return GRPCAnnotation{}
}

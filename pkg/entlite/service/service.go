package service

import "github.com/guntisdev/entlite/pkg/entlite"

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

func (MethodsAnnotation) Annotation() {}

func Methods(...Method) entlite.Annotation {
	var methodList []Method

	return MethodsAnnotation{Methods: methodList}
}

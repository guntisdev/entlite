package entlite

// --------------------------------- string ---------------------------------
type StringFieldBuilder interface {
	Unique() StringFieldBuilder
	Default(string) StringFieldBuilder
	ProtoField(int) StringFieldBuilder

	// to be used in Field
	field()
}

type StringField struct {
	name       string
	unique     bool
	defaultVal *string
	protoField *int
}

// marker method for sealed interface
func (*StringField) field() {}

// constructor
func String(name string) StringFieldBuilder {
	return &StringField{name: name}
}

func (f *StringField) GetUnique() bool {
	return f.unique
}

func (f *StringField) GetDefault() *string {
	return f.defaultVal
}

func (f *StringField) GetProtoField() *int {
	return f.protoField
}

// setters with chaining logic. uses mutable struct
func (f *StringField) Unique() StringFieldBuilder {
	f.unique = true
	return f
}

func (f *StringField) Default(value string) StringFieldBuilder {
	f.defaultVal = &value
	return f
}

func (f *StringField) ProtoField(num int) StringFieldBuilder {
	f.protoField = &num
	return f
}

// --------------------------------- bool ---------------------------------
type BoolField struct {
	name string
}

func (BoolField) field() {}

func Bool(name string) BoolField {
	return BoolField{name: name}
}

// --------------------------------- bool ---------------------------------
type Int32Field struct {
	name string
}

func (Int32Field) field() {}

func Int32(name string) Int32Field {
	return Int32Field{name: name}
}

package entlite

// --------------------------------- string ---------------------------------
type StringField struct {
	name       string
	unique     bool
	defaultVal *string
	protoField *int
}

// marker method for sealed interface
func (*StringField) field() {}

// constructor
func String(name string) *StringField {
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
func (f *StringField) Unique() *StringField {
	f.unique = true
	return f
}

func (f *StringField) Default(value string) *StringField {
	f.defaultVal = &value
	return f
}

func (f *StringField) ProtoField(num int) *StringField {
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

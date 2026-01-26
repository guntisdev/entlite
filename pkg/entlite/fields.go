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
type BoolFieldBuilder interface {
	Default(bool) BoolFieldBuilder
	ProtoField(int) BoolFieldBuilder

	field()
}

type BoolField struct {
	name       string
	defaultVal *bool
	protoField *int
}

func (*BoolField) field() {}

func Bool(name string) BoolFieldBuilder {
	return &BoolField{name: name}
}

func (f *BoolField) GetDefault() *bool {
	return f.defaultVal
}

func (f *BoolField) GetProtoField() *int {
	return f.protoField
}

func (f *BoolField) Default(value bool) BoolFieldBuilder {
	f.defaultVal = &value
	return f
}

func (f *BoolField) ProtoField(num int) BoolFieldBuilder {
	f.protoField = &num
	return f
}

// --------------------------------- int32 ---------------------------------
type Int32FieldBuilder interface {
	Default(int32) Int32FieldBuilder
	ProtoField(int) Int32FieldBuilder

	field()
}

type Int32Field struct {
	name       string
	defaultVal *int32
	protoField *int
}

func (*Int32Field) field() {}

func Int32(name string) Int32FieldBuilder {
	return &Int32Field{name: name}
}

func (f *Int32Field) GetDefault() *int32 {
	return f.defaultVal
}

func (f *Int32Field) GetProtoField() *int {
	return f.protoField
}

func (f *Int32Field) Default(value int32) Int32FieldBuilder {
	f.defaultVal = &value
	return f
}

func (f *Int32Field) ProtoField(num int) Int32FieldBuilder {
	f.protoField = &num
	return f
}

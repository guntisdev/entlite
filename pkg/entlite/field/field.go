package field

import "time"

// --------------------------------- string ---------------------------------
type StringFieldBuilder interface {
	Unique() StringFieldBuilder
	Default(string) StringFieldBuilder
	DefaultFunc(func() string) StringFieldBuilder
	ProtoField(int) StringFieldBuilder
	Comment(string) StringFieldBuilder
	Immutable() StringFieldBuilder
	Optional() StringFieldBuilder
	Validate(func() bool) StringFieldBuilder
	// maybe ProtoExclude() - meant to not send out password?

	// to satisfy entlite.Field interface
	Field()
}

type StringField struct {
	name        string
	unique      bool
	defaultVal  *string
	defaultFunc func() string
	protoField  *int
	comment     *string
	immutable   bool
	optional    bool
	validate    func() bool
}

// marker method for sealed interface
func (*StringField) Field() {}

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

func (f *StringField) GetDefaultFunc() func() string {
	return f.defaultFunc
}

func (f *StringField) GetProtoField() *int {
	return f.protoField
}

func (f *StringField) GetComment() *string {
	return f.comment
}

func (f *StringField) GetImmutable() bool {
	return f.immutable
}

func (f *StringField) GetOptional() bool {
	return f.optional
}

func (f *StringField) GetValidate() func() bool {
	return f.validate
}

// setters with chaining logic. uses mutable struct
func (f *StringField) Unique() StringFieldBuilder {
	f.unique = true
	return f
}

func (f *StringField) Default(value string) StringFieldBuilder {
	f.defaultVal = &value
	f.defaultFunc = nil
	return f
}

func (f *StringField) DefaultFunc(fn func() string) StringFieldBuilder {
	f.defaultFunc = fn
	f.defaultVal = nil
	return f
}

func (f *StringField) ProtoField(num int) StringFieldBuilder {
	f.protoField = &num
	return f
}

func (f *StringField) Comment(text string) StringFieldBuilder {
	f.comment = &text
	return f
}

func (f *StringField) Immutable() StringFieldBuilder {
	f.immutable = true
	return f
}

func (f *StringField) Optional() StringFieldBuilder {
	f.optional = true
	return f
}

func (f *StringField) Validate(fn func() bool) StringFieldBuilder {
	f.validate = fn
	return f
}

// --------------------------------- bool ---------------------------------
type BoolFieldBuilder interface {
	Default(bool) BoolFieldBuilder
	ProtoField(int) BoolFieldBuilder
	Comment(string) BoolFieldBuilder
	Validate(func() bool) BoolFieldBuilder

	// to satisfy entlite.Field interface
	Field()
}

type BoolField struct {
	name       string
	defaultVal *bool
	protoField *int
	comment    *string
	validate   func() bool
}

func (*BoolField) Field() {}

func Bool(name string) BoolFieldBuilder {
	return &BoolField{name: name}
}

func (f *BoolField) GetDefault() *bool {
	return f.defaultVal
}

func (f *BoolField) GetProtoField() *int {
	return f.protoField
}

func (f *BoolField) GetComment() *string {
	return f.comment
}

func (f *BoolField) GetValidate() func() bool {
	return f.validate
}

func (f *BoolField) Default(value bool) BoolFieldBuilder {
	f.defaultVal = &value
	return f
}

func (f *BoolField) ProtoField(num int) BoolFieldBuilder {
	f.protoField = &num
	return f
}

func (f *BoolField) Comment(text string) BoolFieldBuilder {
	f.comment = &text
	return f
}

func (f *BoolField) Validate(fn func() bool) BoolFieldBuilder {
	f.validate = fn
	return f
}

// --------------------------------- int32 ---------------------------------
type Int32FieldBuilder interface {
	Default(int32) Int32FieldBuilder
	ProtoField(int) Int32FieldBuilder
	Comment(string) Int32FieldBuilder
	Optional() Int32FieldBuilder
	Validate(func() bool) Int32FieldBuilder

	// to satisfy entlite.Field interface
	Field()
}

type Int32Field struct {
	name       string
	defaultVal *int32
	protoField *int
	comment    *string
	optional   bool
	validate   func() bool
}

func (*Int32Field) Field() {}

func Int32(name string) Int32FieldBuilder {
	return &Int32Field{name: name}
}

func (f *Int32Field) GetDefault() *int32 {
	return f.defaultVal
}

func (f *Int32Field) GetProtoField() *int {
	return f.protoField
}

func (f *Int32Field) GetComment() *string {
	return f.comment
}

func (f *Int32Field) GetOptional() bool {
	return f.optional
}

func (f *Int32Field) GetValidate() func() bool {
	return f.validate
}

func (f *Int32Field) Default(value int32) Int32FieldBuilder {
	f.defaultVal = &value
	return f
}

func (f *Int32Field) ProtoField(num int) Int32FieldBuilder {
	f.protoField = &num
	return f
}

func (f *Int32Field) Comment(text string) Int32FieldBuilder {
	f.comment = &text
	return f
}

func (f *Int32Field) Optional() Int32FieldBuilder {
	f.optional = true
	return f
}

func (f *Int32Field) Validate(fn func() bool) Int32FieldBuilder {
	f.validate = fn
	return f
}

// --------------------------------- time ---------------------------------
type TimeFieldBuilder interface {
	Default(time.Time) TimeFieldBuilder
	DefaultFunc(func() time.Time) TimeFieldBuilder
	ProtoField(int) TimeFieldBuilder
	Comment(string) TimeFieldBuilder
	Immutable() TimeFieldBuilder
	Optional() TimeFieldBuilder
	Validate(func() bool) TimeFieldBuilder

	Field()
}

type TimeField struct {
	name        string
	defaultVal  *time.Time
	defaultFunc func() time.Time
	protoField  *int
	comment     *string
	immutable   bool
	optional    bool
	validate    func() bool
}

func (*TimeField) Field() {}

func Time(name string) TimeFieldBuilder {
	return &TimeField{name: name}
}

func (f *TimeField) GetDefault() *time.Time {
	return f.defaultVal
}

func (f *TimeField) GetDefaultFunc() func() time.Time {
	return f.defaultFunc
}

func (f *TimeField) GetProtoField() *int {
	return f.protoField
}

func (f *TimeField) GetComment() *string {
	return f.comment
}

func (f *TimeField) GetImmutable() bool {
	return f.immutable
}

func (f *TimeField) GetOptional() bool {
	return f.optional
}

func (f *TimeField) GetValidate() func() bool {
	return f.validate
}

func (f *TimeField) Default(value time.Time) TimeFieldBuilder {
	f.defaultFunc = nil
	f.defaultVal = &value
	return f
}

func (f *TimeField) DefaultFunc(fn func() time.Time) TimeFieldBuilder {
	f.defaultVal = nil
	f.defaultFunc = fn
	return f
}

func (f *TimeField) ProtoField(num int) TimeFieldBuilder {
	f.protoField = &num
	return f
}

func (f *TimeField) Comment(text string) TimeFieldBuilder {
	f.comment = &text
	return f
}

func (f *TimeField) Immutable() TimeFieldBuilder {
	f.immutable = true
	return f
}

func (f *TimeField) Optional() TimeFieldBuilder {
	f.optional = true
	return f
}

func (f *TimeField) Validate(fn func() bool) TimeFieldBuilder {
	f.validate = fn
	return f
}

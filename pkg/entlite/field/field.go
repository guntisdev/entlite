package field

import "time"

// TODO implement Byte field, and then all int/float types like uint8, int32 etc

// --------------------------------- string ---------------------------------
type StringFieldBuilder interface {
	Unique() StringFieldBuilder
	Default(string) StringFieldBuilder
	DefaultFunc(func() string) StringFieldBuilder
	ProtoField(int) StringFieldBuilder
	Comment(string) StringFieldBuilder
	Immutable() StringFieldBuilder
	Optional() StringFieldBuilder
	Validate(func(string) bool) StringFieldBuilder
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
	validate    func(string) bool
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

func (f *StringField) GetValidate() func(string) bool {
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

func (f *StringField) Validate(fn func(string) bool) StringFieldBuilder {
	f.validate = fn
	return f
}

// --------------------------------- bool ---------------------------------
type BoolFieldBuilder interface {
	Default(bool) BoolFieldBuilder
	ProtoField(int) BoolFieldBuilder
	Comment(string) BoolFieldBuilder
	Validate(func(bool) bool) BoolFieldBuilder

	// to satisfy entlite.Field interface
	Field()
}

type BoolField struct {
	name       string
	defaultVal *bool
	protoField *int
	comment    *string
	validate   func(bool) bool
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

func (f *BoolField) GetValidate() func(bool) bool {
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

func (f *BoolField) Validate(fn func(bool) bool) BoolFieldBuilder {
	f.validate = fn
	return f
}

// --------------------------------- int ---------------------------------
// int uses int64 as type under the hood - for sqlite and JS compatibility
type IntFieldBuilder interface {
	Default(int64) IntFieldBuilder
	ProtoField(int) IntFieldBuilder
	Comment(string) IntFieldBuilder
	Optional() IntFieldBuilder
	Validate(func(int64) bool) IntFieldBuilder

	// to satisfy entlite.Field interface
	Field()
}

type IntField struct {
	name       string
	defaultVal *int64
	protoField *int
	comment    *string
	optional   bool
	validate   func(int64) bool
}

func (*IntField) Field() {}

func Int(name string) IntFieldBuilder {
	return &IntField{name: name}
}

func (f *IntField) GetDefault() *int64 {
	return f.defaultVal
}

func (f *IntField) GetProtoField() *int {
	return f.protoField
}

func (f *IntField) GetComment() *string {
	return f.comment
}

func (f *IntField) GetOptional() bool {
	return f.optional
}

func (f *IntField) GetValidate() func(int64) bool {
	return f.validate
}

func (f *IntField) Default(value int64) IntFieldBuilder {
	f.defaultVal = &value
	return f
}

func (f *IntField) ProtoField(num int) IntFieldBuilder {
	f.protoField = &num
	return f
}

func (f *IntField) Comment(text string) IntFieldBuilder {
	f.comment = &text
	return f
}

func (f *IntField) Optional() IntFieldBuilder {
	f.optional = true
	return f
}

func (f *IntField) Validate(fn func(int64) bool) IntFieldBuilder {
	f.validate = fn
	return f
}

// --------------------------------- float ---------------------------------
type FloatFieldBuilder interface {
	Default(float64) FloatFieldBuilder
	ProtoField(int) FloatFieldBuilder
	Comment(string) FloatFieldBuilder
	Optional() FloatFieldBuilder
	Validate(func(float64) bool) FloatFieldBuilder

	Field()
}

type FloatField struct {
	name       string
	defaultVal *float64
	protoField *int
	comment    *string
	optional   bool
	validate   func(float64) bool
}

func (*FloatField) Field() {}

func Float(name string) FloatFieldBuilder {
	return &FloatField{name: name}
}

func (f *FloatField) GetDefault() *float64 {
	return f.defaultVal
}

func (f *FloatField) GetProtoField() *int {
	return f.protoField
}

func (f *FloatField) GetComment() *string {
	return f.comment
}

func (f *FloatField) GetOptional() bool {
	return f.optional
}

func (f *FloatField) GetValidate() func(float64) bool {
	return f.validate
}

func (f *FloatField) Default(value float64) FloatFieldBuilder {
	f.defaultVal = &value
	return f
}

func (f *FloatField) ProtoField(num int) FloatFieldBuilder {
	f.protoField = &num
	return f
}

func (f *FloatField) Comment(text string) FloatFieldBuilder {
	f.comment = &text
	return f
}

func (f *FloatField) Optional() FloatFieldBuilder {
	f.optional = true
	return f
}

func (f *FloatField) Validate(fn func(float64) bool) FloatFieldBuilder {
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
	Validate(func(time.Time) bool) TimeFieldBuilder

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
	validate    func(time.Time) bool
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

func (f *TimeField) GetValidate() func(time.Time) bool {
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

func (f *TimeField) Validate(fn func(time.Time) bool) TimeFieldBuilder {
	f.validate = fn
	return f
}

// --------------------------------- byte ---------------------------------
type ByteFieldBuilder interface {
	Unique() ByteFieldBuilder
	Optional() ByteFieldBuilder
	Immutable() ByteFieldBuilder
	ProtoField(int) ByteFieldBuilder
	Comment(string) ByteFieldBuilder
	DefaultFunc(func() []byte) ByteFieldBuilder
	Validate(func([]byte) bool) ByteFieldBuilder

	Field()
}

type ByteField struct {
	name        string
	unique      bool
	optional    bool
	immutable   bool
	protoField  *int
	comment     *string
	defaultFunc func() []byte
	validate    func([]byte) bool
}

func (*ByteField) Field() {}

func Byte(name string) ByteFieldBuilder {
	return &ByteField{name: name}
}

func (f *ByteField) GetUnique() bool {
	return f.unique
}

func (f *ByteField) GetOptional() bool {
	return f.optional
}

func (f *ByteField) GetImmutable() bool {
	return f.immutable
}

func (f *ByteField) GetProtoField() *int {
	return f.protoField
}

func (f *ByteField) GetComment() *string {
	return f.comment
}

func (f *ByteField) GetDefaultFunc() func() []byte {
	return f.defaultFunc
}

func (f *ByteField) GetValidate() func([]byte) bool {
	return f.validate
}

func (f *ByteField) Unique() ByteFieldBuilder {
	f.unique = true
	return f
}

func (f *ByteField) Optional() ByteFieldBuilder {
	f.optional = true
	return f
}

func (f *ByteField) Immutable() ByteFieldBuilder {
	f.immutable = true
	return f
}

func (f *ByteField) ProtoField(num int) ByteFieldBuilder {
	f.protoField = &num
	return f
}

func (f *ByteField) Comment(text string) ByteFieldBuilder {
	f.comment = &text
	return f
}

func (f *ByteField) DefaultFunc(fn func() []byte) ByteFieldBuilder {
	f.defaultFunc = fn
	return f
}

func (f *ByteField) Validate(fn func([]byte) bool) ByteFieldBuilder {
	f.validate = fn
	return f
}

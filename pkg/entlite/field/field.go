package field

import (
	"time"

	"github.com/guntisdev/entlite/pkg/entlite"
)

// --------------------------------- string ---------------------------------
type StringFieldBuilder interface {
	Unique() StringFieldBuilder
	Default(string) StringFieldBuilder
	DefaultFunc(func() string) StringFieldBuilder
	ProtoField(int) StringFieldBuilder
	Contracts(contracts ...entlite.Contract) StringFieldBuilder
	Immutable() StringFieldBuilder
	Optional() StringFieldBuilder
	Validate(func(string) bool) StringFieldBuilder

	// to satisfy entlite.Field interface
	Field()
}

type StringField struct {
	name        string
	unique      bool
	defaultVal  *string
	defaultFunc func() string
	protoField  *int
	contracts   []entlite.Contract
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

func (f *StringField) GetContracts() []entlite.Contract {
	return f.contracts
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

func (f *StringField) Contracts(contracts ...entlite.Contract) StringFieldBuilder {
	f.contracts = contracts
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
	Contracts(contracts ...entlite.Contract) BoolFieldBuilder
	Validate(func(bool) bool) BoolFieldBuilder

	// to satisfy entlite.Field interface
	Field()
}

type BoolField struct {
	name       string
	defaultVal *bool
	protoField *int
	contracts  []entlite.Contract
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

func (f *BoolField) GetContracts() []entlite.Contract {
	return f.contracts
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

func (f *BoolField) Contracts(contracts ...entlite.Contract) BoolFieldBuilder {
	f.contracts = contracts
	return f
}

func (f *BoolField) Validate(fn func(bool) bool) BoolFieldBuilder {
	f.validate = fn
	return f
}

// --------------------------------- int ---------------------------------
// int is int32, to match a JS number
type IntFieldBuilder interface {
	Default(int32) IntFieldBuilder
	ProtoField(int) IntFieldBuilder
	Contracts(contracts ...entlite.Contract) IntFieldBuilder
	Optional() IntFieldBuilder
	Validate(func(int32) bool) IntFieldBuilder

	// to satisfy entlite.Field interface
	Field()
}

type IntField struct {
	name       string
	defaultVal *int32
	protoField *int
	contracts  []entlite.Contract
	optional   bool
	validate   func(int32) bool
}

func (*IntField) Field() {}

func Int(name string) IntFieldBuilder {
	return &IntField{name: name}
}

func (f *IntField) GetDefault() *int32 {
	return f.defaultVal
}

func (f *IntField) GetProtoField() *int {
	return f.protoField
}

func (f *IntField) GetContracts() []entlite.Contract {
	return f.contracts
}

func (f *IntField) GetOptional() bool {
	return f.optional
}

func (f *IntField) GetValidate() func(int32) bool {
	return f.validate
}

func (f *IntField) Default(value int32) IntFieldBuilder {
	f.defaultVal = &value
	return f
}

func (f *IntField) ProtoField(num int) IntFieldBuilder {
	f.protoField = &num
	return f
}

func (f *IntField) Contracts(contracts ...entlite.Contract) IntFieldBuilder {
	f.contracts = contracts
	return f
}

func (f *IntField) Optional() IntFieldBuilder {
	f.optional = true
	return f
}

func (f *IntField) Validate(fn func(int32) bool) IntFieldBuilder {
	f.validate = fn
	return f
}

// --------------------------------- int64 ---------------------------------
type Int64FieldBuilder interface {
	Default(int64) Int64FieldBuilder
	ProtoField(int) Int64FieldBuilder
	Contracts(contracts ...entlite.Contract) Int64FieldBuilder
	Optional() Int64FieldBuilder
	Validate(func(int64) bool) Int64FieldBuilder

	// to satisfy entlite.Field interface
	Field()
}

type Int64Field struct {
	name       string
	defaultVal *int64
	protoField *int
	contracts  []entlite.Contract
	optional   bool
	validate   func(int64) bool
}

func (*Int64Field) Field() {}

func Int64(name string) Int64FieldBuilder {
	return &Int64Field{name: name}
}

func (f *Int64Field) GetDefault() *int64 {
	return f.defaultVal
}

func (f *Int64Field) GetProtoField() *int {
	return f.protoField
}

func (f *Int64Field) GetContracts() []entlite.Contract {
	return f.contracts
}

func (f *Int64Field) GetOptional() bool {
	return f.optional
}

func (f *Int64Field) GetValidate() func(int64) bool {
	return f.validate
}

func (f *Int64Field) Default(value int64) Int64FieldBuilder {
	f.defaultVal = &value
	return f
}

func (f *Int64Field) ProtoField(num int) Int64FieldBuilder {
	f.protoField = &num
	return f
}

func (f *Int64Field) Contracts(contracts ...entlite.Contract) Int64FieldBuilder {
	f.contracts = contracts
	return f
}

func (f *Int64Field) Optional() Int64FieldBuilder {
	f.optional = true
	return f
}

func (f *Int64Field) Validate(fn func(int64) bool) Int64FieldBuilder {
	f.validate = fn
	return f
}

// --------------------------------- float ---------------------------------
type FloatFieldBuilder interface {
	Default(float64) FloatFieldBuilder
	ProtoField(int) FloatFieldBuilder
	Contracts(contracts ...entlite.Contract) FloatFieldBuilder
	Optional() FloatFieldBuilder
	Validate(func(float64) bool) FloatFieldBuilder

	Field()
}

type FloatField struct {
	name       string
	defaultVal *float64
	protoField *int
	contracts  []entlite.Contract
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

func (f *FloatField) GetContracts() []entlite.Contract {
	return f.contracts
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

func (f *FloatField) Contracts(contracts ...entlite.Contract) FloatFieldBuilder {
	f.contracts = contracts
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
	Contracts(contracts ...entlite.Contract) TimeFieldBuilder
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
	contracts   []entlite.Contract
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

func (f *TimeField) GetContracts() []entlite.Contract {
	return f.contracts
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

func (f *TimeField) Contracts(contracts ...entlite.Contract) TimeFieldBuilder {
	f.contracts = contracts
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
	Contracts(contracts ...entlite.Contract) ByteFieldBuilder
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
	contracts   []entlite.Contract
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

func (f *ByteField) GetContracts() []entlite.Contract {
	return f.contracts
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

func (f *ByteField) Contracts(contracts ...entlite.Contract) ByteFieldBuilder {
	f.contracts = contracts
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

// --------------------------------- json ---------------------------------
type JSONFieldBuilder interface {
	Optional() JSONFieldBuilder
	Immutable() JSONFieldBuilder
	ProtoField(int) JSONFieldBuilder
	Contracts(contracts ...entlite.Contract) JSONFieldBuilder
	// Default takes raw json text, e.g. `{}` or `{"theme":"dark"}`
	Default(string) JSONFieldBuilder
	DefaultFunc(func() string) JSONFieldBuilder
	Validate(func(string) bool) JSONFieldBuilder

	Field()
}

type JSONField struct {
	name        string
	optional    bool
	immutable   bool
	protoField  *int
	contracts   []entlite.Contract
	defaultVal  *string
	defaultFunc func() string
	validate    func(string) bool
}

func (*JSONField) Field() {}

func JSON(name string) JSONFieldBuilder {
	return &JSONField{name: name}
}

func (f *JSONField) GetOptional() bool {
	return f.optional
}

func (f *JSONField) GetImmutable() bool {
	return f.immutable
}

func (f *JSONField) GetProtoField() *int {
	return f.protoField
}

func (f *JSONField) GetContracts() []entlite.Contract {
	return f.contracts
}

func (f *JSONField) GetDefault() *string {
	return f.defaultVal
}

func (f *JSONField) GetDefaultFunc() func() string {
	return f.defaultFunc
}

func (f *JSONField) GetValidate() func(string) bool {
	return f.validate
}

func (f *JSONField) Optional() JSONFieldBuilder {
	f.optional = true
	return f
}

func (f *JSONField) Immutable() JSONFieldBuilder {
	f.immutable = true
	return f
}

func (f *JSONField) ProtoField(num int) JSONFieldBuilder {
	f.protoField = &num
	return f
}

func (f *JSONField) Contracts(contracts ...entlite.Contract) JSONFieldBuilder {
	f.contracts = contracts
	return f
}

func (f *JSONField) Default(value string) JSONFieldBuilder {
	f.defaultFunc = nil
	f.defaultVal = &value
	return f
}

func (f *JSONField) DefaultFunc(fn func() string) JSONFieldBuilder {
	f.defaultVal = nil
	f.defaultFunc = fn
	return f
}

func (f *JSONField) Validate(fn func(string) bool) JSONFieldBuilder {
	f.validate = fn
	return f
}

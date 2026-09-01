// Package field holds the field builders used to describe an entity schema.
package field

import (
	"time"

	"github.com/guntisdev/entlite/pkg/entlite"
)

// --------------------------------- string ---------------------------------

// StringFieldBuilder builds a string field.
type StringFieldBuilder interface {
	// Unique makes the value unique across all rows.
	Unique() StringFieldBuilder
	// Default sets a fixed default value, it clears DefaultFunc.
	Default(string) StringFieldBuilder
	// DefaultFunc sets a default value computed on insert, it clears Default.
	DefaultFunc(func() string) StringFieldBuilder
	// ProtoField pins the proto field number, so it stays stable.
	ProtoField(int) StringFieldBuilder
	// Contracts limits the field to the given layers, sqlc or proto.
	Contracts(contracts ...entlite.Contract) StringFieldBuilder
	// Immutable blocks changes after the row is created.
	Immutable() StringFieldBuilder
	// Optional allows the column to be NULL.
	Optional() StringFieldBuilder
	// Validate checks the value before it is written.
	Validate(func(string) bool) StringFieldBuilder

	// to satisfy entlite.Field interface
	Field()
}

// StringField holds the state of a string field.
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

// String creates a string field with the given column name.
func String(name string) StringFieldBuilder {
	return &StringField{name: name}
}

// GetUnique reports if the field is unique.
func (f *StringField) GetUnique() bool {
	return f.unique
}

// GetDefault returns the fixed default, or nil when there is none.
func (f *StringField) GetDefault() *string {
	return f.defaultVal
}

// GetDefaultFunc returns the default function, or nil when there is none.
func (f *StringField) GetDefaultFunc() func() string {
	return f.defaultFunc
}

// GetProtoField returns the pinned proto field number, or nil.
func (f *StringField) GetProtoField() *int {
	return f.protoField
}

// GetContracts returns the layers the field belongs to.
func (f *StringField) GetContracts() []entlite.Contract {
	return f.contracts
}

// GetImmutable reports if the field is immutable.
func (f *StringField) GetImmutable() bool {
	return f.immutable
}

// GetOptional reports if the field is optional.
func (f *StringField) GetOptional() bool {
	return f.optional
}

// GetValidate returns the validation function, or nil.
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

// BoolFieldBuilder builds a bool field.
type BoolFieldBuilder interface {
	// Default sets a fixed default value.
	Default(bool) BoolFieldBuilder
	// ProtoField pins the proto field number, so it stays stable.
	ProtoField(int) BoolFieldBuilder
	// Contracts limits the field to the given layers, sqlc or proto.
	Contracts(contracts ...entlite.Contract) BoolFieldBuilder
	// Validate checks the value before it is written.
	Validate(func(bool) bool) BoolFieldBuilder

	// to satisfy entlite.Field interface
	Field()
}

// BoolField holds the state of a bool field.
type BoolField struct {
	name       string
	defaultVal *bool
	protoField *int
	contracts  []entlite.Contract
	validate   func(bool) bool
}

// marker method for sealed interface
func (*BoolField) Field() {}

// Bool creates a bool field with the given column name.
func Bool(name string) BoolFieldBuilder {
	return &BoolField{name: name}
}

// GetDefault returns the fixed default, or nil when there is none.
func (f *BoolField) GetDefault() *bool {
	return f.defaultVal
}

// GetProtoField returns the pinned proto field number, or nil.
func (f *BoolField) GetProtoField() *int {
	return f.protoField
}

// GetContracts returns the layers the field belongs to.
func (f *BoolField) GetContracts() []entlite.Contract {
	return f.contracts
}

// GetValidate returns the validation function, or nil.
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

// IntFieldBuilder builds an int field. int is int32, to match a JS number.
type IntFieldBuilder interface {
	// Default sets a fixed default value.
	Default(int32) IntFieldBuilder
	// ProtoField pins the proto field number, so it stays stable.
	ProtoField(int) IntFieldBuilder
	// Contracts limits the field to the given layers, sqlc or proto.
	Contracts(contracts ...entlite.Contract) IntFieldBuilder
	// Optional allows the column to be NULL.
	Optional() IntFieldBuilder
	// Validate checks the value before it is written.
	Validate(func(int32) bool) IntFieldBuilder

	// to satisfy entlite.Field interface
	Field()
}

// IntField holds the state of an int field.
type IntField struct {
	name       string
	defaultVal *int32
	protoField *int
	contracts  []entlite.Contract
	optional   bool
	validate   func(int32) bool
}

// marker method for sealed interface
func (*IntField) Field() {}

// Int creates an int32 field with the given column name.
func Int(name string) IntFieldBuilder {
	return &IntField{name: name}
}

// GetDefault returns the fixed default, or nil when there is none.
func (f *IntField) GetDefault() *int32 {
	return f.defaultVal
}

// GetProtoField returns the pinned proto field number, or nil.
func (f *IntField) GetProtoField() *int {
	return f.protoField
}

// GetContracts returns the layers the field belongs to.
func (f *IntField) GetContracts() []entlite.Contract {
	return f.contracts
}

// GetOptional reports if the field is optional.
func (f *IntField) GetOptional() bool {
	return f.optional
}

// GetValidate returns the validation function, or nil.
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

// Int64FieldBuilder builds an int64 field.
type Int64FieldBuilder interface {
	// Default sets a fixed default value.
	Default(int64) Int64FieldBuilder
	// ProtoField pins the proto field number, so it stays stable.
	ProtoField(int) Int64FieldBuilder
	// Contracts limits the field to the given layers, sqlc or proto.
	Contracts(contracts ...entlite.Contract) Int64FieldBuilder
	// Optional allows the column to be NULL.
	Optional() Int64FieldBuilder
	// Validate checks the value before it is written.
	Validate(func(int64) bool) Int64FieldBuilder

	// to satisfy entlite.Field interface
	Field()
}

// Int64Field holds the state of an int64 field.
type Int64Field struct {
	name       string
	defaultVal *int64
	protoField *int
	contracts  []entlite.Contract
	optional   bool
	validate   func(int64) bool
}

// marker method for sealed interface
func (*Int64Field) Field() {}

// Int64 creates an int64 field with the given column name.
func Int64(name string) Int64FieldBuilder {
	return &Int64Field{name: name}
}

// GetDefault returns the fixed default, or nil when there is none.
func (f *Int64Field) GetDefault() *int64 {
	return f.defaultVal
}

// GetProtoField returns the pinned proto field number, or nil.
func (f *Int64Field) GetProtoField() *int {
	return f.protoField
}

// GetContracts returns the layers the field belongs to.
func (f *Int64Field) GetContracts() []entlite.Contract {
	return f.contracts
}

// GetOptional reports if the field is optional.
func (f *Int64Field) GetOptional() bool {
	return f.optional
}

// GetValidate returns the validation function, or nil.
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

// FloatFieldBuilder builds a float64 field.
type FloatFieldBuilder interface {
	// Default sets a fixed default value.
	Default(float64) FloatFieldBuilder
	// ProtoField pins the proto field number, so it stays stable.
	ProtoField(int) FloatFieldBuilder
	// Contracts limits the field to the given layers, sqlc or proto.
	Contracts(contracts ...entlite.Contract) FloatFieldBuilder
	// Optional allows the column to be NULL.
	Optional() FloatFieldBuilder
	// Validate checks the value before it is written.
	Validate(func(float64) bool) FloatFieldBuilder

	// to satisfy entlite.Field interface
	Field()
}

// FloatField holds the state of a float64 field.
type FloatField struct {
	name       string
	defaultVal *float64
	protoField *int
	contracts  []entlite.Contract
	optional   bool
	validate   func(float64) bool
}

// marker method for sealed interface
func (*FloatField) Field() {}

// Float creates a float64 field with the given column name.
func Float(name string) FloatFieldBuilder {
	return &FloatField{name: name}
}

// GetDefault returns the fixed default, or nil when there is none.
func (f *FloatField) GetDefault() *float64 {
	return f.defaultVal
}

// GetProtoField returns the pinned proto field number, or nil.
func (f *FloatField) GetProtoField() *int {
	return f.protoField
}

// GetContracts returns the layers the field belongs to.
func (f *FloatField) GetContracts() []entlite.Contract {
	return f.contracts
}

// GetOptional reports if the field is optional.
func (f *FloatField) GetOptional() bool {
	return f.optional
}

// GetValidate returns the validation function, or nil.
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

// TimeFieldBuilder builds a timestamp field.
type TimeFieldBuilder interface {
	// Default sets a fixed default value, it clears DefaultFunc.
	Default(time.Time) TimeFieldBuilder
	// DefaultFunc sets a default value computed on insert, it clears Default.
	DefaultFunc(func() time.Time) TimeFieldBuilder
	// ProtoField pins the proto field number, so it stays stable.
	ProtoField(int) TimeFieldBuilder
	// Contracts limits the field to the given layers, sqlc or proto.
	Contracts(contracts ...entlite.Contract) TimeFieldBuilder
	// Immutable blocks changes after the row is created.
	Immutable() TimeFieldBuilder
	// Optional allows the column to be NULL.
	Optional() TimeFieldBuilder
	// Validate checks the value before it is written.
	Validate(func(time.Time) bool) TimeFieldBuilder

	// to satisfy entlite.Field interface
	Field()
}

// TimeField holds the state of a time field.
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

// marker method for sealed interface
func (*TimeField) Field() {}

// Time creates a timestamp field with the given column name.
func Time(name string) TimeFieldBuilder {
	return &TimeField{name: name}
}

// GetDefault returns the fixed default, or nil when there is none.
func (f *TimeField) GetDefault() *time.Time {
	return f.defaultVal
}

// GetDefaultFunc returns the default function, or nil when there is none.
func (f *TimeField) GetDefaultFunc() func() time.Time {
	return f.defaultFunc
}

// GetProtoField returns the pinned proto field number, or nil.
func (f *TimeField) GetProtoField() *int {
	return f.protoField
}

// GetContracts returns the layers the field belongs to.
func (f *TimeField) GetContracts() []entlite.Contract {
	return f.contracts
}

// GetImmutable reports if the field is immutable.
func (f *TimeField) GetImmutable() bool {
	return f.immutable
}

// GetOptional reports if the field is optional.
func (f *TimeField) GetOptional() bool {
	return f.optional
}

// GetValidate returns the validation function, or nil.
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

// ByteFieldBuilder builds a raw bytes field.
type ByteFieldBuilder interface {
	// Unique makes the value unique across all rows.
	Unique() ByteFieldBuilder
	// Optional allows the column to be NULL.
	Optional() ByteFieldBuilder
	// Immutable blocks changes after the row is created.
	Immutable() ByteFieldBuilder
	// ProtoField pins the proto field number, so it stays stable.
	ProtoField(int) ByteFieldBuilder
	// Contracts limits the field to the given layers, sqlc or proto.
	Contracts(contracts ...entlite.Contract) ByteFieldBuilder
	// DefaultFunc sets a default value computed on insert.
	DefaultFunc(func() []byte) ByteFieldBuilder
	// Validate checks the value before it is written.
	Validate(func([]byte) bool) ByteFieldBuilder

	// to satisfy entlite.Field interface
	Field()
}

// ByteField holds the state of a bytes field.
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

// marker method for sealed interface
func (*ByteField) Field() {}

// Byte creates a bytes field with the given column name.
func Byte(name string) ByteFieldBuilder {
	return &ByteField{name: name}
}

// GetUnique reports if the field is unique.
func (f *ByteField) GetUnique() bool {
	return f.unique
}

// GetOptional reports if the field is optional.
func (f *ByteField) GetOptional() bool {
	return f.optional
}

// GetImmutable reports if the field is immutable.
func (f *ByteField) GetImmutable() bool {
	return f.immutable
}

// GetProtoField returns the pinned proto field number, or nil.
func (f *ByteField) GetProtoField() *int {
	return f.protoField
}

// GetContracts returns the layers the field belongs to.
func (f *ByteField) GetContracts() []entlite.Contract {
	return f.contracts
}

// GetDefaultFunc returns the default function, or nil when there is none.
func (f *ByteField) GetDefaultFunc() func() []byte {
	return f.defaultFunc
}

// GetValidate returns the validation function, or nil.
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

// JSONFieldBuilder builds a json field, stored as text.
type JSONFieldBuilder interface {
	// Optional allows the column to be NULL.
	Optional() JSONFieldBuilder
	// Immutable blocks changes after the row is created.
	Immutable() JSONFieldBuilder
	// ProtoField pins the proto field number, so it stays stable.
	ProtoField(int) JSONFieldBuilder
	// Contracts limits the field to the given layers, sqlc or proto.
	Contracts(contracts ...entlite.Contract) JSONFieldBuilder
	// Default takes raw json text, e.g. `{}` or `{"theme":"dark"}`
	Default(string) JSONFieldBuilder
	// DefaultFunc sets a default value computed on insert, it clears Default.
	DefaultFunc(func() string) JSONFieldBuilder
	// Validate checks the value before it is written.
	Validate(func(string) bool) JSONFieldBuilder

	// to satisfy entlite.Field interface
	Field()
}

// JSONField holds the state of a json field.
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

// marker method for sealed interface
func (*JSONField) Field() {}

// JSON creates a json field with the given column name.
func JSON(name string) JSONFieldBuilder {
	return &JSONField{name: name}
}

// GetOptional reports if the field is optional.
func (f *JSONField) GetOptional() bool {
	return f.optional
}

// GetImmutable reports if the field is immutable.
func (f *JSONField) GetImmutable() bool {
	return f.immutable
}

// GetProtoField returns the pinned proto field number, or nil.
func (f *JSONField) GetProtoField() *int {
	return f.protoField
}

// GetContracts returns the layers the field belongs to.
func (f *JSONField) GetContracts() []entlite.Contract {
	return f.contracts
}

// GetDefault returns the fixed default, or nil when there is none.
func (f *JSONField) GetDefault() *string {
	return f.defaultVal
}

// GetDefaultFunc returns the default function, or nil when there is none.
func (f *JSONField) GetDefaultFunc() func() string {
	return f.defaultFunc
}

// GetValidate returns the validation function, or nil.
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

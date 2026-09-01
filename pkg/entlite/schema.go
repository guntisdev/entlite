// Package entlite holds the core types of the schema dsl.
package entlite

// Schema is embedded in a user schema struct to mark it as an entity.
type Schema struct{}

// Contract is one output layer a field or query belongs to.
type Contract interface {
	Contract()
}

// Layer is a contract that can be narrowed to one direction.
type Layer interface {
	Contract
	// ReadOnly exposes the field or query only when reading.
	ReadOnly() Contract
	// WriteOnly exposes the field or query only when writing.
	WriteOnly() Contract
}

// Field is implemented by every field builder.
type Field interface {
	Field()
}

// Query is implemented by every query builder.
type Query interface {
	Query()
}

// Index is implemented by every index builder.
type Index interface {
	Index()
}

// SQLCContract is the sqlc layer, it drives the database code.
type SQLCContract struct {
	readOnly  bool
	writeOnly bool
}

// marker method for sealed interface
func (SQLCContract) Contract() {}

// ReadOnly exposes the field or query only when reading.
func (c SQLCContract) ReadOnly() Contract {
	c.readOnly = true
	return c
}

// WriteOnly exposes the field or query only when writing.
func (c SQLCContract) WriteOnly() Contract {
	c.writeOnly = true
	return c
}

// SQLC selects the sqlc layer.
func SQLC() Layer {
	return SQLCContract{}
}

// PROTOContract is the proto layer, it drives the api surface.
type PROTOContract struct {
	readOnly  bool
	writeOnly bool
}

// marker method for sealed interface
func (PROTOContract) Contract() {}

// ReadOnly exposes the field or query only when reading.
func (c PROTOContract) ReadOnly() Contract {
	c.readOnly = true
	return c
}

// WriteOnly exposes the field or query only when writing.
func (c PROTOContract) WriteOnly() Contract {
	c.writeOnly = true
	return c
}

// PROTO selects the proto layer.
func PROTO() Layer {
	return PROTOContract{}
}

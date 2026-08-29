package entlite

type Schema struct{}

type Contract interface {
	Contract()
}

type Layer interface {
	Contract
	ReadOnly() Contract
	WriteOnly() Contract
}

type Field interface {
	Field()
}

type Query interface {
	Query()
}

type Index interface {
	Index()
}

type SQLCContract struct {
	readOnly  bool
	writeOnly bool
}

func (SQLCContract) Contract() {}

func (c SQLCContract) ReadOnly() Contract {
	c.readOnly = true
	return c
}

func (c SQLCContract) WriteOnly() Contract {
	c.writeOnly = true
	return c
}

func SQLC() Layer {
	return SQLCContract{}
}

type PROTOContract struct {
	readOnly  bool
	writeOnly bool
}

func (PROTOContract) Contract() {}

func (c PROTOContract) ReadOnly() Contract {
	c.readOnly = true
	return c
}

func (c PROTOContract) WriteOnly() Contract {
	c.writeOnly = true
	return c
}

func PROTO() Layer {
	return PROTOContract{}
}

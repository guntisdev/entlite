package entlite

type Schema struct{}

type Contract interface {
	Contract()
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

type SQLCContract struct{}

func (SQLCContract) Contract() {}

func SQLC() Contract {
	return SQLCContract{}
}

type PROTOContract struct{}

func (PROTOContract) Contract() {}

func PROTO(contracts ...Contract) Contract {
	return PROTOContract{}
}

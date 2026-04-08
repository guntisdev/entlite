package query

type Type string

const (
	TypeGetBy  Type = "get_by"
	TypeListBy Type = "list_by"
)

type QueryBuilder interface {
	Query()
}

type Query struct {
	typeName Type
	field    string
}

// marker method for sealed interface
func (Query) Query() {}

func GetBy(field string) QueryBuilder {
	return Query{typeName: TypeGetBy, field: field}
}

func ListBy(field string) QueryBuilder {
	return Query{typeName: TypeListBy, field: field}
}

func (q Query) GetType() Type {
	return q.typeName
}

func (q Query) GetField() string {
	return q.field
}

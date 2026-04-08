package query

type Type string

const (
	TypeDefaultCRUD Type = "default_crud"
	TypeCreate      Type = "create"
	TypeGet         Type = "get"
	TypeUpdate      Type = "update"
	TypeDelete      Type = "delete"
	TypeList        Type = "list"
	TypeGetBy       Type = "get_by"
	TypeListBy      Type = "list_by"
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

func DefaultCRUD() QueryBuilder {
	return Query{typeName: TypeDefaultCRUD}
}

func Create() QueryBuilder {
	return Query{typeName: TypeCreate}
}

func Get() QueryBuilder {
	return Query{typeName: TypeGet}
}

func Update() QueryBuilder {
	return Query{typeName: TypeUpdate}
}

func Delete() QueryBuilder {
	return Query{typeName: TypeDelete}
}

func List() QueryBuilder {
	return Query{typeName: TypeList}
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

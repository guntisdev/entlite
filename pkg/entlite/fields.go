package entlite

// string
type StringField struct {
	name string
}

func (StringField) field() {}

func String(name string) StringField {
	return StringField{name: name}
}

// bool
type BoolField struct {
	name string
}

func (BoolField) field() {}

func Bool(name string) BoolField {
	return BoolField{name: name}
}

// int32
type Int32Field struct {
	name string
}

func (Int32Field) field() {}

func Int32(name string) Int32Field {
	return Int32Field{name: name}
}

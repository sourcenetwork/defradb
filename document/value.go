package document

// Value is an interface that points to a concrete Value implementation
// May collapse this down without an interface
type Value interface {
	Value() interface{}
	IsDocument() bool
}

type simpleValue struct {
	value interface{}
}

func newValue(val interface{}) simpleValue {
	return simpleValue{val}
}

// func (val simpleValue) Set(val interface{})

func (val simpleValue) Value() interface{} {
	return val.value
}

func (val simpleValue) IsDocument() bool {
	_, ok := val.value.(*Document)
	return ok
}

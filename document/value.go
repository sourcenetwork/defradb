package document

// Value is an interface that points to a concrete Value implementation
// May collapse this down without an interface
type Value interface {
	Value() interface{}
}

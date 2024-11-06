package connor

// FilterKey represents a type that may be used as a map key
// in a filter.
type FilterKey interface {
	// PropertyAndOperator returns the data and operator that should be used
	// to filter the value matching this key.
	//
	// If the key does not have an operator the given defaultOp will be returned.
	PropertyAndOperator(data any, defaultOp string) (any, string, error)
	// Equal returns true if other is equal, otherwise returns false.
	Equal(other FilterKey) bool
}

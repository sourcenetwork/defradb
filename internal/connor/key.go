package connor

// KeyResult represents the result of a filter key operation.
type KeyResult struct {
	// Data is the data that should be used to filter the value matching the key.
	Data any
	// MissProp is true if the key is missing a property, otherwise false.
	// It's relevant for object of dynamic type, like JSON.
	MissProp bool
	// Operator is the operator that should be used to filter the value matching the key.
	// If the key does not have an operator the given defaultOp will be returned.
	Operator string
	// Err is the error that occurred while filtering the value matching the key.
	Err error
}

// FilterKey represents a type that may be used as a map key
// in a filter.
type FilterKey interface {
	// PropertyAndOperator returns [KeyResult] that contains data and operator that should be
	// used to filter the value matching this key.
	PropertyAndOperator(data any, defaultOp string) KeyResult
	// Equal returns true if other is equal, otherwise returns false.
	Equal(other FilterKey) bool
}

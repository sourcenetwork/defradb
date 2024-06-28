package connor

// FilterKey represents a type that may be used as a map key
// in a filter.
type FilterKey interface {
	// GetProp returns the data that should be used with this key
	// from the given data.
	GetProp(data any) any
	// GetOperatorOrDefault returns either the operator that corresponds
	// to this key, or the given default.
	GetOperatorOrDefault(defaultOp string) string
	// Equal returns true if other is equal, otherwise returns false.
	Equal(other FilterKey) bool
}

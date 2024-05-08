package connor

// ne performs object inequality comparisons by inverting
// the result of the EqualOperator for non-error cases.
func ne(conditions, data any) (bool, error) {
	m, err := eq(conditions, data)

	if err != nil {
		return false, err
	}

	return !m, err
}

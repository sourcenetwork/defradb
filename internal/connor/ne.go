package connor

// ne performs object inequality comparisons by inverting
// the result of the EqualOperator for non-error cases.
func ne(conditions, data any, propExists bool) (bool, error) {
	// _ne operator should return false if the property does not exist.
	if !propExists {
		return false, nil
	}

	m, err := eq(conditions, data, propExists)

	if err != nil {
		return false, err
	}

	return !m, err
}

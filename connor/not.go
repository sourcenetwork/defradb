package connor

// not is an operator which performs object equality test
// and returns the inverse of the result.
func not(condition, data any) (bool, error) {
	m, err := eq(condition, data)
	if err != nil {
		return false, err
	}
	return !m, nil
}

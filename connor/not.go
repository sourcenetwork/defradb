package connor

// not is an operator which performs object equality test
// and returns the inverse of the result.
func not(condition, data any) (bool, error) {
	if m, err := eq(condition, data); err != nil {
		return false, err
	} else if m {
		return false, nil
	}

	return true, nil
}

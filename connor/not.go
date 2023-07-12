package connor

// like is an operator which performs string equality
// tests.
func not(condition, data any) (bool, error) {
	if m, err := eq(condition, data); err != nil {
		return false, err
	} else if m {
		return false, nil
	}

	return true, nil
}

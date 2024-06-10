package connor

// not is an operator which performs object equality test
// and returns the inverse of the result.
func not(condition, data any) (bool, error) {
	if m, ok := condition.(map[FilterKey]any); ok && len(m) == 0 {
		return false, NewErrEmptyObject()
	}
	m, err := eq(condition, data)
	if err != nil {
		return false, err
	}
	return !m, nil
}

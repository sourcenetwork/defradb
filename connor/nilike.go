package connor

// nilike performs case insensitive string inequality comparisons by inverting
// the result of the Like operator for non-error cases.
func nilike(conditions, data any) (bool, error) {
	m, err := ilike(conditions, data)
	if err != nil {
		return false, err
	}

	return !m, err
}

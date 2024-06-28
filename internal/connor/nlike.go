package connor

// nlike performs string inequality comparisons by inverting
// the result of the Like operator for non-error cases.
func nlike(conditions, data any) (bool, error) {
	m, err := like(conditions, data)

	if err != nil {
		return false, err
	}

	return !m, err
}

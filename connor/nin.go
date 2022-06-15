package connor

func init() {
	opMap["nin"] = nin
}

// nin performs set exclusion comparisons by inverting the results
// of the InOperator under non-error conditions.
func nin(conditions, data interface{}) (bool, error) {
	m, err := in(conditions, data)

	if err != nil {
		return false, err
	}

	return !m, err
}

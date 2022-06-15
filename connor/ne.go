package connor

func init() {
	opMap["ne"] = ne
}

// ne performs object inequality comparisons by inverting
// the result of the EqualOperator for non-error cases.
func ne(conditions, data interface{}) (bool, error) {
	m, err := matchWith("$eq", conditions, data)

	if err != nil {
		return false, err
	}

	return !m, err
}

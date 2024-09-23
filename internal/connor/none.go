package connor

// none is an operator which allows the evaluation of
// a number of conditions over a list of values
// matching if all of them do not match.
func none(condition, data any) (bool, error) {
	m, err := anyOp(condition, data)
	if err != nil {
		return false, err
	}
	return !m, nil
}

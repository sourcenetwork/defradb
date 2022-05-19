package connor

func init() {
	Register(&NotEqualOperator{})
}

// NotEqualOperator performs object inequality comparisons by inverting
// the result of the EqualOperator for non-error cases.
type NotEqualOperator struct {
}

func (o *NotEqualOperator) Name() string {
	return "ne"
}

func (o *NotEqualOperator) Evaluate(conditions, data interface{}) (bool, error) {
	m, err := MatchWith("$eq", conditions, data)

	if err != nil {
		return false, err
	}

	return !m, err
}

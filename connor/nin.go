package connor

func init() {
	Register(&NotInOperator{})
}

// NotInOperator performs set exclusion comparisons by inverting the results
// of the InOperator under non-error conditions.
type NotInOperator struct {
}

func (o *NotInOperator) Name() string {
	return "nin"
}

func (o *NotInOperator) Evaluate(conditions, data interface{}) (bool, error) {
	m, err := MatchWith("$in", conditions, data)

	if err != nil {
		return false, err
	}

	return !m, err
}

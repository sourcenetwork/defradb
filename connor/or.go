package connor

import "fmt"

func init() {
	Register(&OrOperator{})
}

// OrOperator is an operator which allows the evaluation of
// of a number of conditions, matching if any of them match.
type OrOperator struct {
}

func (o *OrOperator) Name() string {
	return "or"
}

func (o *OrOperator) Evaluate(condition, data interface{}) (bool, error) {
	switch cn := condition.(type) {
	case []interface{}:
		for _, c := range cn {
			if m, err := MatchWith("$eq", c, data); err != nil {
				return false, err
			} else if m {
				return true, nil
			}
		}

		return false, nil
	default:
		return false, fmt.Errorf("unknown or condition type '%#v'", cn)
	}
}

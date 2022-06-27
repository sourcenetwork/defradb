package connor

import "fmt"

func init() {
	Register(&AndOperator{})
}

// AndOperator is an operator which allows the evaluation of
// of a number of conditions, matching if all of them match.
type AndOperator struct {
}

func (o *AndOperator) Name() string {
	return "and"
}

func (o *AndOperator) Evaluate(condition, data interface{}) (bool, error) {
	switch cn := condition.(type) {
	case []interface{}:
		for _, c := range cn {
			if m, err := MatchWith("$eq", c, data); err != nil {
				return false, err
			} else if !m {
				return false, nil
			}
		}

		return true, nil
	default:
		return false, fmt.Errorf("unknown or condition type '%#v'", cn)
	}
}

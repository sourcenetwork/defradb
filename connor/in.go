package connor

import "fmt"

func init() {
	Register(&InOperator{})
}

// InOperator will determine whether a value exists within the
// condition's array of available values.
type InOperator struct {
}

func (o *InOperator) Name() string {
	return "in"
}

func (o *InOperator) Evaluate(conditions, data interface{}) (bool, error) {
	switch cn := conditions.(type) {
	case []interface{}:
		for _, ce := range cn {
			if m, err := MatchWith("$eq", ce, data); err != nil {
				return false, err
			} else if m {
				return true, nil
			}
		}

		return false, nil
	default:
		return false, fmt.Errorf("unknown value type")
	}
}

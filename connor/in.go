package connor

import "fmt"

// in will determine whether a value exists within the
// condition's array of available values.
func in(conditions, data interface{}) (bool, error) {
	switch cn := conditions.(type) {
	case []interface{}:
		for _, ce := range cn {
			if m, err := eq(ce, data); err != nil {
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

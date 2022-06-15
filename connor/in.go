package connor

import "fmt"

func init() {
	opMap["in"] = in
}

// in will determine whether a value exists within the
// condition's array of available values.
func in(conditions, data interface{}) (bool, error) {
	switch cn := conditions.(type) {
	case []interface{}:
		for _, ce := range cn {
			if m, err := matchWith("$eq", ce, data); err != nil {
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

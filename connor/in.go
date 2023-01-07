package connor

import "github.com/sourcenetwork/defradb/client"

// in will determine whether a value exists within the
// condition's array of available values.
func in(conditions, data any) (bool, error) {
	switch cn := conditions.(type) {
	case []any:
		for _, ce := range cn {
			if m, err := eq(ce, data); err != nil {
				return false, err
			} else if m {
				return true, nil
			}
		}

		return false, nil
	default:
		return false, client.NewErrUnhandledType("condition", cn)
	}
}

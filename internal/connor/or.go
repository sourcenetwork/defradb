package connor

import "github.com/sourcenetwork/defradb/client"

// or is an operator which allows the evaluation of
// of a number of conditions, matching if any of them match.
func or(condition, data any) (bool, error) {
	switch cn := condition.(type) {
	case []any:
		for _, c := range cn {
			if m, err := eq(c, data); err != nil {
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

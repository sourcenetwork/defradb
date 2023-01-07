package connor

import "github.com/sourcenetwork/defradb/client"

// and is an operator which allows the evaluation of
// of a number of conditions, matching if all of them match.
func and(condition, data any) (bool, error) {
	switch cn := condition.(type) {
	case []any:
		for _, c := range cn {
			if m, err := eq(c, data); err != nil {
				return false, err
			} else if !m {
				return false, nil
			}
		}

		return true, nil
	default:
		return false, client.NewErrUnhandledType("condition", cn)
	}
}

package connor

import (
	"github.com/sourcenetwork/defradb/client"

	"github.com/sourcenetwork/immutable"
)

// anyOp is an operator which allows the evaluation of
// a number of conditions over a list of values
// matching if any of them match.
func anyOp(condition, data any) (bool, error) {
	switch t := data.(type) {
	case []string:
		for _, c := range t {
			m, err := eq(condition, c)
			if err != nil {
				return false, err
			} else if m {
				return true, nil
			}
		}
		return false, nil

	case []immutable.Option[string]:
		for _, c := range t {
			m, err := eq(condition, c)
			if err != nil {
				return false, err
			} else if m {
				return true, nil
			}
		}
		return false, nil

	case []int64:
		for _, c := range t {
			m, err := eq(condition, c)
			if err != nil {
				return false, err
			} else if m {
				return true, nil
			}
		}
		return false, nil

	case []immutable.Option[int64]:
		for _, c := range t {
			m, err := eq(condition, c)
			if err != nil {
				return false, err
			} else if m {
				return true, nil
			}
		}
		return false, nil

	case []bool:
		for _, c := range t {
			m, err := eq(condition, c)
			if err != nil {
				return false, err
			} else if m {
				return true, nil
			}
		}
		return false, nil

	case []immutable.Option[bool]:
		for _, c := range t {
			m, err := eq(condition, c)
			if err != nil {
				return false, err
			} else if m {
				return true, nil
			}
		}
		return false, nil

	case []float64:
		for _, c := range t {
			m, err := eq(condition, c)
			if err != nil {
				return false, err
			} else if m {
				return true, nil
			}
		}
		return false, nil

	case []immutable.Option[float64]:
		for _, c := range t {
			m, err := eq(condition, c)
			if err != nil {
				return false, err
			} else if m {
				return true, nil
			}
		}
		return false, nil

	default:
		return false, client.NewErrUnhandledType("data", data)
	}
}

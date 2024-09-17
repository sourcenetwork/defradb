package connor

import (
	"github.com/sourcenetwork/defradb/client"

	"github.com/sourcenetwork/immutable"
)

// all is an operator which allows the evaluation of
// a number of conditions over a list of values
// matching if all of them match.
func all(condition, data any) (bool, error) {
	switch t := data.(type) {
	case []string:
		for _, c := range t {
			m, err := eq(condition, c)
			if err != nil {
				return false, err
			} else if !m {
				return false, nil
			}
		}
		return true, nil

	case []immutable.Option[string]:
		for _, c := range t {
			m, err := eq(condition, c)
			if err != nil {
				return false, err
			} else if !m {
				return false, nil
			}
		}
		return true, nil

	case []int64:
		for _, c := range t {
			m, err := eq(condition, c)
			if err != nil {
				return false, err
			} else if !m {
				return false, nil
			}
		}
		return true, nil

	case []immutable.Option[int64]:
		for _, c := range t {
			m, err := eq(condition, c)
			if err != nil {
				return false, err
			} else if !m {
				return false, nil
			}
		}
		return true, nil

	case []bool:
		for _, c := range t {
			m, err := eq(condition, c)
			if err != nil {
				return false, err
			} else if !m {
				return false, nil
			}
		}
		return true, nil

	case []immutable.Option[bool]:
		for _, c := range t {
			m, err := eq(condition, c)
			if err != nil {
				return false, err
			} else if !m {
				return false, nil
			}
		}
		return true, nil

	case []float64:
		for _, c := range t {
			m, err := eq(condition, c)
			if err != nil {
				return false, err
			} else if !m {
				return false, nil
			}
		}
		return true, nil

	case []immutable.Option[float64]:
		for _, c := range t {
			m, err := eq(condition, c)
			if err != nil {
				return false, err
			} else if !m {
				return false, nil
			}
		}
		return true, nil

	default:
		return false, client.NewErrUnhandledType("data", data)
	}
}

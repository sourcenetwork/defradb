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
		return allSlice(condition, t)

	case []immutable.Option[string]:
		return allSlice(condition, t)

	case []int64:
		return allSlice(condition, t)

	case []immutable.Option[int64]:
		return allSlice(condition, t)

	case []bool:
		return allSlice(condition, t)

	case []immutable.Option[bool]:
		return allSlice(condition, t)

	case []float64:
		return allSlice(condition, t)

	case []immutable.Option[float64]:
		return allSlice(condition, t)

	default:
		return false, client.NewErrUnhandledType("data", data)
	}
}

func allSlice[T any](condition any, data []T) (bool, error) {
	for _, c := range data {
		m, err := eq(condition, c)
		if err != nil {
			return false, err
		} else if !m {
			return false, nil
		}
	}
	return true, nil
}

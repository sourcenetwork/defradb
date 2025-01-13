package connor

import "github.com/sourcenetwork/immutable"

// none is an operator which allows the evaluation of
// a number of conditions over a list of values
// matching if all of them do not match.
func none(condition, data any) (bool, error) {
	switch t := data.(type) {
	case []any:
		return noneSlice(condition, t)

	case []string:
		return noneSlice(condition, t)

	case []immutable.Option[string]:
		return noneSlice(condition, t)

	case []int64:
		return noneSlice(condition, t)

	case []immutable.Option[int64]:
		return noneSlice(condition, t)

	case []bool:
		return noneSlice(condition, t)

	case []immutable.Option[bool]:
		return noneSlice(condition, t)

	case []float64:
		return noneSlice(condition, t)

	case []immutable.Option[float64]:
		return noneSlice(condition, t)

	default:
		return false, nil
	}
}

func noneSlice[T any](condition any, data []T) (bool, error) {
	for _, c := range data {
		m, err := eq(condition, c)
		if err != nil {
			return false, err
		} else if m {
			return false, nil
		}
	}
	return true, nil
}

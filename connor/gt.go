package connor

import (
	"fmt"

	"github.com/sourcenetwork/defradb/connor/numbers"
)

// gt does value comparisons to determine whether one
// value is strictly larger than another.
func gt(condition, data any) (bool, error) {
	if condition == nil {
		return data != nil, nil
	}

	switch cn := numbers.TryUpcast(condition).(type) {
	case float64:
		switch dn := numbers.TryUpcast(data).(type) {
		case float64:
			return dn > cn, nil
		case int64:
			return float64(dn) > cn, nil
		}

		return false, nil
	case int64:
		switch dn := numbers.TryUpcast(data).(type) {
		case float64:
			return dn > float64(cn), nil
		case int64:
			return dn > cn, nil
		}

		return false, nil
	default:
		return false, fmt.Errorf("unknown comparison type '%#v'", condition)
	}
}

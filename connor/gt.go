package connor

import (
	"fmt"
	"time"

	"github.com/sourcenetwork/defradb/connor/numbers"
	"github.com/sourcenetwork/defradb/errors"
)

// gt does value comparisons to determine whether one
// value is strictly larger than another.
func gt(condition, data any) (bool, error) {
	if condition == nil {
		return data != nil, nil
	}

	switch c := condition.(type) {
	case time.Time:
		switch d := data.(type) {
		case time.Time:
			return d.After(c), nil
		case string:
			dt, err := time.Parse(time.RFC3339, d)
			if err != nil {
				return false, err
			}
			return dt.After(c), nil
		default:
			return false, errors.New(fmt.Sprintf("3unknown comparison type '%#v'", condition))
		}
	default:
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
			return false, errors.New(fmt.Sprintf("4unknown comparison type '%#v'", condition))
		}
	}
}

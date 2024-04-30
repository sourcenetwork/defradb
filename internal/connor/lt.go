package connor

import (
	"time"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/connor/numbers"
)

// lt does value comparisons to determine whether one
// value is strictly less than another.
func lt(condition, data any) (bool, error) {
	if condition == nil {
		return false, nil
	}

	switch c := condition.(type) {
	case time.Time:
		switch d := data.(type) {
		case time.Time:
			return d.Before(c), nil
		case string:
			dt, err := time.Parse(time.RFC3339, d)
			if err != nil {
				return false, err
			}
			return dt.Before(c), nil
		default:
			return false, client.NewErrUnhandledType("data", d)
		}
	default:
		switch cn := numbers.TryUpcast(condition).(type) {
		case float64:
			switch dn := numbers.TryUpcast(data).(type) {
			case float64:
				return dn < cn, nil
			case int64:
				return float64(dn) < cn, nil
			}

			return false, nil
		case int64:
			switch dn := numbers.TryUpcast(data).(type) {
			case float64:
				return dn < float64(cn), nil
			case int64:
				return dn < cn, nil
			}

			return false, nil
		default:
			return false, client.NewErrUnhandledType("condition", cn)
		}
	}
}

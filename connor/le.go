package connor

import (
	"fmt"
	"time"

	"github.com/sourcenetwork/defradb/connor/numbers"
)

func init() {
	Register(&LessEqualOperator{})
}

// LessEqualOperator does value comparisons to determine whether one
// value is strictly less than another.
type LessEqualOperator struct {
}

func (o *LessEqualOperator) Name() string {
	return "le"
}

func (o *LessEqualOperator) Evaluate(condition, data interface{}) (bool, error) {
	switch cn := numbers.TryUpcast(condition).(type) {
	case float64:
		switch dn := numbers.TryUpcast(data).(type) {
		case float64:
			return dn <= cn, nil
		case int64:
			return float64(dn) <= cn, nil
		}

		return false, nil
	case int64:
		switch dn := numbers.TryUpcast(data).(type) {
		case float64:
			return dn <= float64(cn), nil
		case int64:
			return dn <= cn, nil
		}

		return false, nil
	case time.Time:
		switch dn := data.(type) {
		case time.Time:
			return !dn.After(cn), nil
		}
		return false, nil
	default:
		return false, fmt.Errorf("unknown comparison type '%#v'", condition)
	}
}

package connor

import (
	"reflect"
	"time"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/internal/connor/numbers"
	ctime "github.com/sourcenetwork/defradb/internal/connor/time"
	"github.com/sourcenetwork/defradb/internal/core"
)

// eq is an operator which performs object equality
// tests.
func eq(condition, data any) (bool, error) {
	switch arr := data.(type) {
	case []core.Doc:
		for _, item := range arr {
			m, err := eq(condition, item)
			if err != nil {
				return false, err
			}

			if m {
				return true, nil
			}
		}
		return false, nil

	case immutable.Option[bool]:
		if !arr.HasValue() {
			return condition == nil, nil
		}
		data = arr.Value()

	case immutable.Option[int64]:
		if !arr.HasValue() {
			return condition == nil, nil
		}
		data = arr.Value()

	case immutable.Option[float64]:
		if !arr.HasValue() {
			return condition == nil, nil
		}
		data = arr.Value()

	case immutable.Option[string]:
		if !arr.HasValue() {
			return condition == nil, nil
		}
		data = arr.Value()
	}

	switch cn := condition.(type) {
	case string:
		if d, ok := data.(string); ok {
			return d == cn, nil
		}
		return false, nil
	case int64:
		return numbers.Equal(cn, data), nil
	case int32:
		return numbers.Equal(cn, data), nil
	case float64:
		return numbers.Equal(cn, data), nil
	case map[FilterKey]any:
		m := true
		for prop, cond := range cn {
			var err error
			m, err = matchWith(prop.GetOperatorOrDefault("_eq"), cond, prop.GetProp(data))
			if err != nil {
				return false, err
			}

			if !m {
				// No need to evaluate after we fail
				break
			}
		}

		return m, nil
	case time.Time:
		return ctime.Equal(cn, data), nil
	default:
		return reflect.DeepEqual(condition, data), nil
	}
}

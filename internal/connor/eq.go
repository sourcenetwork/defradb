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
		return anySlice(condition, arr)

	case immutable.Option[bool]:
		data = immutableValueOrNil(arr)

	case immutable.Option[int64]:
		data = immutableValueOrNil(arr)

	case immutable.Option[float64]:
		data = immutableValueOrNil(arr)

	case immutable.Option[string]:
		data = immutableValueOrNil(arr)
	}

	switch cn := condition.(type) {
	case map[FilterKey]any:
		for prop, cond := range cn {
			m, err := matchWith(prop.GetOperatorOrDefault("_eq"), cond, prop.GetProp(data))
			if err != nil {
				return false, err
			} else if !m {
				return false, nil
			}
		}
		return true, nil

	case map[string]any:
		d, ok := data.(map[string]any)
		if !ok {
			return false, nil
		}
		for k, v := range d {
			m, err := eq(cn[k], v)
			if err != nil {
				return false, err
			} else if !m {
				return false, nil
			}
		}
		return true, nil

	case string:
		if d, ok := data.(string); ok {
			return d == cn, nil
		}
		return false, nil

	case uint64:
		return numbers.Equal(cn, data), nil

	case int64:
		return numbers.Equal(cn, data), nil

	case uint32:
		return numbers.Equal(cn, data), nil

	case int32:
		return numbers.Equal(cn, data), nil

	case float64:
		return numbers.Equal(cn, data), nil

	case time.Time:
		return ctime.Equal(cn, data), nil

	default:
		return reflect.DeepEqual(condition, data), nil
	}
}

func immutableValueOrNil[T any](data immutable.Option[T]) any {
	if data.HasValue() {
		return data.Value()
	}
	return nil
}

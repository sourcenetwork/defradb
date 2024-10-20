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
			m, err := matchWith(prop.GetOperatorOrDefault(EqualOp), cond, prop.GetProp(data))
			if err != nil {
				return false, err
			} else if !m {
				return false, nil
			}
		}
		return true, nil

	case map[string]any:
		return objectsEqual(cn, data)

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

	case time.Time:
		return ctime.Equal(cn, data), nil

	default:
		return reflect.DeepEqual(condition, data), nil
	}
}

// objectsEqual returns true if the given condition and data
// contain equal key value pairs.
func objectsEqual(condition map[string]any, data any) (bool, error) {
	if data == nil {
		return condition == nil, nil
	}
	d, ok := data.(map[string]any)
	if !ok || len(d) != len(condition) {
		return false, nil
	}
	for k, v := range d {
		m, err := eq(condition[k], v)
		if err != nil {
			return false, err
		} else if !m {
			return false, nil
		}
	}
	return true, nil
}

func immutableValueOrNil[T any](data immutable.Option[T]) any {
	if data.HasValue() {
		return data.Value()
	}
	return nil
}

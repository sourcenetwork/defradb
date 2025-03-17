package connor

import (
	"reflect"
	"time"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/internal/connor/numbers"
	ctime "github.com/sourcenetwork/defradb/internal/connor/time"
	"github.com/sourcenetwork/defradb/internal/core"
)

// eq is an operator which performs object equality tests.
// It also takes a propExists boolean to indicate if the property exists in the data.
// It's needed because the behavior of the operators can change if the property doesn't exist.
// For example, _ne operator should return true if the property doesn't exist.
// This can also be used in the future if we introduce operators line _has.
func eq(condition, data any, propExists bool) (bool, error) {
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
			res, err := prop.PropertyAndOperator(data, EqualOp)
			if err != nil {
				return false, err
			}
			// If the property doesn't exist, we should pass it forward to nested operators.
			m, err := matchWith(res.Operator, cond, res.Data, !res.MissProp && propExists)
			if err != nil {
				return false, err
			}
			if !m {
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
		m, err := eq(condition[k], v, true)
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

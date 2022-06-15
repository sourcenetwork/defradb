package connor

import (
	"reflect"

	"github.com/sourcenetwork/defradb/connor/numbers"
	"github.com/sourcenetwork/defradb/core"
)

func init() {
	Register(&EqualOperator{})
}

// EqualOperator is an operator which performs object equality
// tests.
type EqualOperator struct {
}

func (o *EqualOperator) Name() string {
	return "eq"
}

func (o *EqualOperator) Evaluate(condition, data interface{}) (bool, error) {
	switch arr := data.(type) {
	case []core.Doc:
		for _, item := range arr {
			m, err := matchWith("$eq", condition, item)
			if err != nil {
				return false, err
			}

			if m {
				return true, nil
			}
		}
		return false, nil
	}

	switch cn := condition.(type) {
	case string:
		if d, ok := data.(string); ok {
			return d == cn, nil
		}
		return false, nil
	case int64:
		return numbers.Equal(cn, data), nil
	case float64:
		return numbers.Equal(cn, data), nil
	case map[FilterKey]interface{}:
		m := true
		for prop, cond := range cn {
			if !m {
				// No need to evaluate after we fail
				continue
			}

			mm, err := matchWith(prop.GetOperatorOrDefault("$eq"), cond, prop.GetProp(data))
			if err != nil {
				return false, err
			}

			m = m && mm
		}

		return m, nil
	default:
		return reflect.DeepEqual(condition, data), nil
	}
}

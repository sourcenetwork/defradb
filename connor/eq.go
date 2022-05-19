package connor

import (
	"reflect"
	"strings"

	"github.com/sourcenetwork/defradb/connor/fields"
	"github.com/sourcenetwork/defradb/connor/numbers"
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
	case []interface{}:
		for _, item := range arr {
			m, err := MatchWith("$eq", condition, item)
			if err != nil {
				return false, err
			}

			if m {
				return true, nil
			}
		}
	case []map[string]interface{}:
		for _, item := range arr {
			m, err := MatchWith("$eq", condition, item)
			if err != nil {
				return false, err
			}

			if m {
				return true, nil
			}
		}
	}

	switch cn := condition.(type) {
	case string:
		if d, ok := data.(string); ok {
			return d == cn, nil
		}
		return false, nil
	case int8:
		return numbers.Equal(cn, data), nil
	case int16:
		return numbers.Equal(cn, data), nil
	case int32:
		return numbers.Equal(cn, data), nil
	case int64:
		return numbers.Equal(cn, data), nil
	case float32:
		return numbers.Equal(cn, data), nil
	case float64:
		return numbers.Equal(cn, data), nil
	case map[string]interface{}:
		m := true
		for prop, cond := range cn {
			if !m {
				// No need to evaluate after we fail
				continue
			}

			if strings.HasPrefix(prop, "$") {
				mm, err := MatchWith(prop, cond, data)
				if err != nil {
					return false, err
				}

				m = m && mm
			} else if d, ok := data.(map[string]interface{}); ok {
				mm, err := MatchWith("$eq", cond, fields.TryGet(d, prop))
				if err != nil {
					return false, err
				}

				m = m && mm
			} else {
				return reflect.DeepEqual(condition, data), nil
			}
		}

		return m, nil
	default:
		return reflect.DeepEqual(condition, data), nil
	}
}

package connor

import (
	"reflect"
	"strings"

	"github.com/sourcenetwork/immutable"
)

// like is an operator which performs sting equality
// tests.
func like(condition, data any) (bool, error) {
	switch arr := data.(type) {
	case immutable.Option[string]:
		if !arr.HasValue() {
			return condition == nil, nil
		}
		data = arr.Value()
	}

	switch cn := condition.(type) {
	case string:
		if d, ok := data.(string); ok {
			hasPrefix := false
			hasSuffix := false
			if len(cn) >= 2 {
				if cn[0] == '%' {
					hasPrefix = true
					cn = strings.TrimPrefix(cn, "%")
				}
				if cn[len(cn)-1] == '%' {
					hasSuffix = true
					cn = strings.TrimSuffix(cn, "%")
				}
			}

			switch {
			case hasPrefix && hasSuffix:
				return strings.Contains(d, cn), nil
			case hasPrefix:
				// if the condition has a prefix string, this means that we are matching
				// the condition has being a suffix to the data.
				return strings.HasSuffix(d, cn), nil
			case hasSuffix:
				// if the condition has a suffic string, this means that we are matching
				// the condition has being a prefix to the data.
				return strings.HasPrefix(d, cn), nil
			default:
				return cn == d, nil
			}
		}
		return false, nil
	default:
		return reflect.DeepEqual(condition, data), nil
	}
}

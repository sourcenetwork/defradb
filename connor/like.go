package connor

import (
	"strings"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
)

// like is an operator which performs string equality
// tests.
func like(condition, data any) (bool, error) {
	switch d := data.(type) {
	case immutable.Option[string]:
		if !d.HasValue() {
			return condition == nil, nil
		}
		data = d.Value()
	}

	switch cn := condition.(type) {
	case string:
		if d, ok := data.(string); ok {
			hasPrefix := false
			hasSuffix := false
			startAndEnd := []string{}

			if len(cn) >= 2 {
				if cn[0] == '%' {
					hasPrefix = true
					cn = strings.TrimPrefix(cn, "%")
				}
				if cn[len(cn)-1] == '%' {
					hasSuffix = true
					cn = strings.TrimSuffix(cn, "%")
				}
				if !hasPrefix && !hasSuffix {
					startAndEnd = strings.Split(cn, "%")
				}
			}

			switch {
			case hasPrefix && hasSuffix:
				return strings.Contains(d, cn), nil

			case hasPrefix:
				// if the condition has a prefix string `%`, this means that we are matching
				// the condition as being a suffix to the data.
				return strings.HasSuffix(d, cn), nil

			case hasSuffix:
				// if the condition has a suffix string `%`, this means that we are matching
				// the condition as being a prefix to the data.
				return strings.HasPrefix(d, cn), nil

			case len(startAndEnd) == 2:
				return strings.HasPrefix(d, startAndEnd[0]) && strings.HasSuffix(d, startAndEnd[1]), nil

			default:
				return cn == d, nil
			}
		}
		return false, nil
	default:
		return false, client.NewErrUnhandledType("condition", cn)
	}
}

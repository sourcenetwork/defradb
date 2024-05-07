package connor

import (
	"strings"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
)

// ilike is an operator which performs case insensitive string equality tests.
func ilike(condition, data any) (bool, error) {
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
			return like(strings.ToLower(cn), strings.ToLower(d))
		}
		return false, nil
	default:
		return false, client.NewErrUnhandledType("condition", cn)
	}
}

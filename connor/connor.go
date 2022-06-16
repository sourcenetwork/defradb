package connor

import (
	"fmt"

	"github.com/sourcenetwork/defradb/core"
)

// Match is the default method used in Connor to match some data to a
// set of conditions.
func Match(conditions map[FilterKey]interface{}, data core.Doc) (bool, error) {
	return eq(conditions, data)
}

// matchWith can be used to specify the exact operator to use when performing
// a match operation. This is primarily used when building custom operators or
// if you wish to override the behavior of another operator.
func matchWith(op string, conditions, data interface{}) (bool, error) {
	if op == "" {
		return false, fmt.Errorf("operator cannot be empty")
	}

	switch op[1:] {
	case "and":
		return and(conditions, data)
	case "eq":
		return eq(conditions, data)
	case "ge":
		return ge(conditions, data)
	case "gt":
		return gt(conditions, data)
	case "in":
		return in(conditions, data)
	case "le":
		return le(conditions, data)
	case "lt":
		return lt(conditions, data)
	case "ne":
		return ne(conditions, data)
	case "nin":
		return nin(conditions, data)
	case "or":
		return or(conditions, data)
	default:
		return false, fmt.Errorf("unknown operator '%s'", op[1:])
	}
}

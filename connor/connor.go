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

	operator, ok := opMap[op[1:]]
	if !ok {
		return false, fmt.Errorf("unknown operator '%s'", op[1:])
	}

	return operator(conditions, data)
}

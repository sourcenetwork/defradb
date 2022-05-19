package connor

import (
	"fmt"
	"strings"
)

func init() {
	Register(&ContainsOperator{})
}

// The ContainsOperator determines whether a string contains
// the provided substring.
type ContainsOperator struct{}

func (o *ContainsOperator) Name() string {
	return "contains"
}

func (o *ContainsOperator) Evaluate(conditions, data interface{}) (bool, error) {
	if c, ok := conditions.(string); ok {
		if d, ok := data.(string); ok {
			return strings.Contains(d, c), nil
		} else if data == nil {
			return false, nil
		}
	}

	return false, fmt.Errorf("contains operator only works with strings")
}

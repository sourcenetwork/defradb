package connor

import (
	"fmt"
	"strings"
)

func init() {
	opMap["contains"] = contains
}

// contains determines whether a string contains
// the provided substring.
func contains(conditions, data interface{}) (bool, error) {
	if c, ok := conditions.(string); ok {
		if d, ok := data.(string); ok {
			return strings.Contains(d, c), nil
		} else if data == nil {
			return false, nil
		}
	}

	return false, fmt.Errorf("contains operator only works with strings")
}

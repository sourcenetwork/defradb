package fields

import (
	"strconv"
	"strings"
)

// Get will fetch the specified field from data recursively, if possible..
func Get(data map[string]interface{}, field string) (interface{}, bool) {
	fps := strings.Split(field, ".")
	d := interface{}(data)
	for _, fp := range fps {
		switch td := d.(type) {
		case map[string]interface{}:
			f, ok := td[fp]
			if !ok {
				return nil, false
			}

			d = f
		case []interface{}:
			fpi, err := strconv.Atoi(fp)
			if err != nil || fpi >= len(td) || fpi < 0 {
				return nil, false
			}
			d = td[fpi]
		default:
			return nil, false
		}
	}

	return d, true
}

// TryGet will attempt to get a field and return nil if it could not
// be found. If you need to differentiate between a field which could
// not be found or one which was equal to nil, use .Get() instead.
func TryGet(data map[string]interface{}, field string) interface{} {
	value, _ := Get(data, field)
	return value
}

package time

import "time"

func Equal(condition, data any) bool {
	switch c := condition.(type) {
	case time.Time:
		switch d := data.(type) {
		case time.Time:
			return d.Equal(c)
		case string:
			dt, err := time.Parse(time.RFC3339, d)
			if err != nil {
				return false
			}
			return dt.Equal(c)
		default:
			return false
		}
	default:
		return false
	}
}

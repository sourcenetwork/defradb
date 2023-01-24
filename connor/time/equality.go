package time

import "time"

func Equal(condition, data any) bool {
	switch c := condition.(type) {
	case time.Time:
		switch d := data.(type) {
		case time.Time:
			return d.Equal(c)
		case string:
			// todo: Not sure if we should be
			// parsing incoming data here, or
			// if the DB should handle this.
			// (Note: This isnt the user provided
			// condition on a request, but the data
			// stored in DB for a document
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

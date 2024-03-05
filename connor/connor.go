/*
Package connor provides a domain-specific language to express conditions against data.

It is derived from https://github.com/SierraSoftworks/connor.
*/
package connor

// Match is the default method used in Connor to match some data to a
// set of conditions.
func Match(conditions map[FilterKey]any, data any) (bool, error) {
	return eq(conditions, data)
}

// matchWith can be used to specify the exact operator to use when performing
// a match operation. This is primarily used when building custom operators or
// if you wish to override the behavior of another operator.
func matchWith(op string, conditions, data any) (bool, error) {
	switch op {
	case "_and":
		return and(conditions, data)
	case "_eq":
		return eq(conditions, data)
	case "_ge":
		return ge(conditions, data)
	case "_gt":
		return gt(conditions, data)
	case "_in":
		return in(conditions, data)
	case "_le":
		return le(conditions, data)
	case "_lt":
		return lt(conditions, data)
	case "_ne":
		return ne(conditions, data)
	case "_nin":
		return nin(conditions, data)
	case "_or":
		return or(conditions, data)
	case "_like":
		return like(conditions, data)
	case "_nlike":
		return nlike(conditions, data)
	case "_ilike":
		return ilike(conditions, data)
	case "_nilike":
		return nilike(conditions, data)
	case "_not":
		return not(conditions, data)
	default:
		return false, NewErrUnknownOperator(op)
	}
}

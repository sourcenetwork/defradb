// Package connor provides a domain-specific language to express conditions against data.
//
// It is derived from https://github.com/SierraSoftworks/connor.
//
// Note to developers:
// Never assume that the data given to an operator is of a certain type.
// Operators should work with any data type so that untyped data
// such as JSON can be filtered as expected.
package connor

const (
	AndOp = "_and"
	OrOp  = "_or"
	NotOp = "_not"

	AnyOp  = "_any"
	AllOp  = "_all"
	NoneOp = "_none"

	EqualOp                  = "_eq"
	GreaterOrEqualOp         = "_ge"
	GreaterOp                = "_gt"
	InOp                     = "_in"
	LesserOrEqualOp          = "_le"
	LesserOp                 = "_lt"
	NotEqualOp               = "_ne"
	NotInOp                  = "_nin"
	LikeOp                   = "_like"
	NotLikeOp                = "_nlike"
	CaseInsensitiveLikeOp    = "_ilike"
	CaseInsensitiveNotLikeOp = "_nilike"
)

// IsOpSimple returns true if the given operator is simple (not compound).
//
// This is useful for checking if a filter operator requires further expansion.
func IsOpSimple(op string) bool {
	switch op {
	case EqualOp, GreaterOrEqualOp, GreaterOp, InOp,
		LesserOrEqualOp, LesserOp, NotEqualOp, NotInOp,
		LikeOp, NotLikeOp, CaseInsensitiveLikeOp, CaseInsensitiveNotLikeOp:
		return true
	default:
		return false
	}
}

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
	case AndOp:
		return and(conditions, data)
	case AnyOp:
		return anyOp(conditions, data)
	case AllOp:
		return all(conditions, data)
	case EqualOp:
		return eq(conditions, data)
	case GreaterOrEqualOp:
		return ge(conditions, data)
	case GreaterOp:
		return gt(conditions, data)
	case InOp:
		return in(conditions, data)
	case LesserOrEqualOp:
		return le(conditions, data)
	case LesserOp:
		return lt(conditions, data)
	case NotEqualOp:
		return ne(conditions, data)
	case NotInOp:
		return nin(conditions, data)
	case OrOp:
		return or(conditions, data)
	case LikeOp:
		return like(conditions, data)
	case NotLikeOp:
		return nlike(conditions, data)
	case CaseInsensitiveLikeOp:
		return ilike(conditions, data)
	case CaseInsensitiveNotLikeOp:
		return nilike(conditions, data)
	case NoneOp:
		return none(conditions, data)
	case NotOp:
		return not(conditions, data)
	default:
		return false, NewErrUnknownOperator(op)
	}
}

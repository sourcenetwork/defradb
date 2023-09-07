package filter

import (
	"github.com/sourcenetwork/defradb/connor"
	"github.com/sourcenetwork/defradb/planner/mapper"
)

// IsComplex returns true if the provided filter is complex.
// A filter is considered complex if it contains a relation
// object withing an _or operator not necessarily being
// its direct  child.
func IsComplex(filter *mapper.Filter) bool {
	if filter == nil {
		return false
	}
	return isComplex(filter.Conditions, false)
}

func isComplex(conditions any, isInsideOr bool) bool {
	switch typedCond := conditions.(type) {
	case map[connor.FilterKey]any:
		for k, v := range typedCond {
			if op, ok := k.(*mapper.Operator); ok && op.Operation == "_or" && len(v.([]any)) > 1 {
				if isComplex(v, true) {
					return true
				}
				continue
			}
			if _, isProp := k.(*mapper.PropertyIndex); isProp && isInsideOr {
				objMap := v.(map[connor.FilterKey]any)
				for objK := range objMap {
					if _, isRelation := objK.(*mapper.PropertyIndex); isRelation {
						return true
					}
				}
			}
			if isComplex(v, isInsideOr) {
				return true
			}
		}
	case []any:
		for _, v := range typedCond {
			if isComplex(v, isInsideOr) {
				return true
			}
		}
	default:
		return false
	}
	return false
}

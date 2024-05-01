// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.
package filter

import (
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/connor"
	"github.com/sourcenetwork/defradb/planner/mapper"
)

// IsComplex returns true if the provided filter is complex.
// A filter is considered complex if it contains a relation
// object withing an _or, _and or _not operator not necessarily being
// its direct  child.
func IsComplex(filter *mapper.Filter) bool {
	if filter == nil {
		return false
	}
	return isComplex(filter.Conditions, false)
}

func isComplex(conditions any, seekRelation bool) bool {
	switch typedCond := conditions.(type) {
	case map[connor.FilterKey]any:
		for k, v := range typedCond {
			if op, ok := k.(*mapper.Operator); ok {
				switch op.Operation {
				case request.FilterOpOr, request.FilterOpAnd, request.FilterOpNot:
					if v, ok := v.([]any); ok && len(v) == 0 {
						continue
					}
					if isComplex(v, true) {
						return true
					}
					continue
				}
			}
			if _, isProp := k.(*mapper.PropertyIndex); isProp && seekRelation {
				objMap := v.(map[connor.FilterKey]any)
				for objK := range objMap {
					if _, isRelation := objK.(*mapper.PropertyIndex); isRelation {
						return true
					}
				}
			}
			if isComplex(v, seekRelation) {
				return true
			}
		}
	case []any:
		for _, v := range typedCond {
			if isComplex(v, seekRelation) {
				return true
			}
		}
	default:
		return false
	}
	return false
}

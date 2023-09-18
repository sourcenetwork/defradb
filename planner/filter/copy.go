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
	"github.com/sourcenetwork/defradb/connor"
)

// Copy performs a deep copy of the provided filter.
func Copy(filter map[connor.FilterKey]any) map[connor.FilterKey]any {
	return copyFilterConditions(filter).(map[connor.FilterKey]any)
}

func copyFilterConditions(conditions any) any {
	switch typedCond := conditions.(type) {
	case map[connor.FilterKey]any:
		result := make(map[connor.FilterKey]any)
		for key, clause := range typedCond {
			result[key] = copyFilterConditions(clause)
		}
		return result
	case []any:
		resultArr := make([]any, len(typedCond))
		for i, elementClause := range typedCond {
			resultArr[i] = copyFilterConditions(elementClause)
		}
		return resultArr
	default:
		return conditions
	}
}

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
	"github.com/sourcenetwork/defradb/planner/mapper"
)

// copyField copies the given field from the provided filter.
// The result filter preserves the structure of the original filter.
func copyField(filter *mapper.Filter, field mapper.Field) *mapper.Filter {
	if filter == nil {
		return nil
	}
	conditionKey := &mapper.PropertyIndex{
		Index: field.Index,
	}

	resultFilter := &mapper.Filter{}
	conditionMap := traverseFilterByProperty(conditionKey, filter.Conditions, false)
	if len(conditionMap) > 0 {
		resultFilter.Conditions = conditionMap
		return resultFilter
	}
	return nil
}

func traverseFilterByProperty(
	key *mapper.PropertyIndex,
	conditions map[connor.FilterKey]any,
	shouldDelete bool,
) map[connor.FilterKey]any {
	result := conditions
	if !shouldDelete {
		result = make(map[connor.FilterKey]any)
	}
	for targetKey, clause := range conditions {
		if targetKey.Equal(key) {
			if shouldDelete {
				delete(result, targetKey)
			} else {
				result[key] = clause
			}
		} else if opKey, isOpKey := targetKey.(*mapper.Operator); isOpKey {
			clauseArr, isArr := clause.([]any)
			if isArr {
				resultArr := make([]any, 0)
				for _, elementClause := range clauseArr {
					elementMap, ok := elementClause.(map[connor.FilterKey]any)
					if !ok {
						continue
					}
					compoundCond := traverseFilterByProperty(key, elementMap, shouldDelete)
					if len(compoundCond) > 0 {
						resultArr = append(resultArr, compoundCond)
					}
				}
				if len(resultArr) > 0 {
					result[opKey] = resultArr
				}
			}
		}
	}
	return result
}

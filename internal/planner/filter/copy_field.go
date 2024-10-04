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
	"github.com/sourcenetwork/defradb/internal/connor"
	"github.com/sourcenetwork/defradb/internal/planner/mapper"
)

// CopyField copies the given field from the provided filter.
// Multiple fields can be passed to copy related objects with a certain field.
// In this case every subsequent field is a sub field of the previous one. Eg. bool.author.name
// The result filter preserves the structure of the original filter.
func CopyField(filter *mapper.Filter, fields ...mapper.Field) *mapper.Filter {
	if filter == nil || len(fields) == 0 {
		return nil
	}
	var conditionKeys []*mapper.PropertyIndex
	for _, field := range fields {
		conditionKeys = append(conditionKeys, &mapper.PropertyIndex{
			Index: field.Index,
		})
	}

	conditionMap := traverseFilterByProperty(conditionKeys, filter.Conditions, false)
	if len(conditionMap) > 0 {
		return &mapper.Filter{Conditions: conditionMap}
	}
	return nil
}

func traverseFilterByProperty(
	keys []*mapper.PropertyIndex,
	conditions map[connor.FilterKey]any,
	shouldDelete bool,
) map[connor.FilterKey]any {
	result := conditions
	if !shouldDelete {
		result = make(map[connor.FilterKey]any)
	}
	for targetKey, clause := range conditions {
		if targetKey.Equal(keys[0]) {
			if len(keys) > 1 {
				related := traverseFilterByProperty(keys[1:], clause.(map[connor.FilterKey]any), shouldDelete)
				if shouldDelete && len(related) == 0 {
					delete(result, targetKey)
				} else if len(related) > 0 && !shouldDelete {
					result[keys[0]] = clause
				}
			} else {
				if shouldDelete {
					delete(result, targetKey)
				} else {
					result[keys[0]] = clause
				}
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
					compoundCond := traverseFilterByProperty(keys, elementMap, shouldDelete)
					if len(compoundCond) > 0 {
						resultArr = append(resultArr, compoundCond)
					}
				}
				if len(resultArr) > 0 {
					result[opKey] = resultArr
				} else if shouldDelete {
					delete(result, opKey)
				}
			}
		}
	}
	return result
}

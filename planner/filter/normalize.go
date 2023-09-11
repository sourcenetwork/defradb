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

// normalize normalizes the provided filter conditions.
// The following cases are subject of normalization:
// - _and or _or with one element is removed flattened
// - double _not is removed
// - any number of consecutive _ands with any number of elements is flattened
// As the result object is a map with unique keys (a.k.a. properties),
// while performing flattening of compound operators if the same property
// is present in the result map, both conditions will be moved into an _and
func normalize(conditions map[connor.FilterKey]any) map[connor.FilterKey]any {
	return normalizeConditions(conditions, false).(map[connor.FilterKey]any)
}

func conditionsArrToMap(conditions []any) map[connor.FilterKey]any {
	result := make(map[connor.FilterKey]any)
	for _, clause := range conditions {
		if clauseMap, ok := clause.(map[connor.FilterKey]any); ok {
			for k, v := range clauseMap {
				result[k] = v
			}
		}
	}
	return result
}

func addNormalizedCondition(key connor.FilterKey, val any, m map[connor.FilterKey]any) {
	if _, isProp := key.(*mapper.PropertyIndex); isProp {
		var andOp *mapper.Operator
		var andContent []any
		for existingKey := range m {
			if op, isOp := existingKey.(*mapper.Operator); isOp && op.Operation == request.FilterOpAnd {
				andOp = op
				andContent = m[existingKey].([]any)
				break
			}
		}
		for existingKey := range m {
			if existingKey.Equal(key) {
				existingVal := m[existingKey]
				delete(m, existingKey)
				if andOp == nil {
					andOp = &mapper.Operator{Operation: request.FilterOpAnd}
				}
				m[andOp] = append(
					andContent,
					map[connor.FilterKey]any{existingKey: existingVal},
					map[connor.FilterKey]any{key: val},
				)
				return
			}
		}
		for _, andElement := range andContent {
			elementMap := andElement.(map[connor.FilterKey]any)
			for andElementKey := range elementMap {
				if andElementKey.Equal(key) {
					m[andOp] = append(andContent, map[connor.FilterKey]any{key: val})
					return
				}
			}
		}
	}
	m[key] = val
}

func normalizeConditions(conditions any, skipRoot bool) any {
	result := make(map[connor.FilterKey]any)
	switch typedConditions := conditions.(type) {
	case map[connor.FilterKey]any:
		for rootKey, rootVal := range typedConditions {
			rootOpKey, isRootOp := rootKey.(*mapper.Operator)
			if isRootOp {
				if rootOpKey.Operation == request.FilterOpAnd || rootOpKey.Operation == request.FilterOpOr {
					rootValArr := rootVal.([]any)
					if len(rootValArr) == 1 || rootOpKey.Operation == request.FilterOpAnd && !skipRoot {
						flat := normalizeConditions(conditionsArrToMap(rootValArr), false)
						flatMap := flat.(map[connor.FilterKey]any)
						for k, v := range flatMap {
							addNormalizedCondition(k, v, result)
						}
					} else {
						resultArr := []any{}
						for i := range rootValArr {
							norm := normalizeConditions(rootValArr[i], !skipRoot)
							normMap, ok := norm.(map[connor.FilterKey]any)
							if ok {
								for k, v := range normMap {
									resultArr = append(resultArr, map[connor.FilterKey]any{k: v})
								}
							} else {
								resultArr = append(resultArr, norm)
							}
						}
						addNormalizedCondition(rootKey, resultArr, result)
					}
				} else if rootOpKey.Operation == request.FilterOpNot {
					notMap := rootVal.(map[connor.FilterKey]any)
					if len(notMap) == 1 {
						var k connor.FilterKey
						for k = range notMap {
							break
						}
						norm := normalizeConditions(notMap, true).(map[connor.FilterKey]any)
						delete(notMap, k)
						var v any
						for k, v = range norm {
							break
						}
						if opKey, ok := k.(*mapper.Operator); ok && opKey.Operation == request.FilterOpNot {
							notNotMap := normalizeConditions(v, false).(map[connor.FilterKey]any)
							for notNotKey, notNotVal := range notNotMap {
								addNormalizedCondition(notNotKey, notNotVal, result)
							}
						} else {
							notMap[k] = v
							addNormalizedCondition(rootOpKey, notMap, result)
						}
					} else {
						addNormalizedCondition(rootKey, rootVal, result)
					}
				} else {
					addNormalizedCondition(rootKey, rootVal, result)
				}
			} else {
				addNormalizedCondition(rootKey, normalizeConditions(rootVal, false), result)
			}
		}
		return result
	case []any:
		return conditionsArrToMap(typedConditions)
	default:
		return conditions
	}
}

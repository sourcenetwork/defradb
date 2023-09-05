package filter

import (
	"github.com/sourcenetwork/defradb/connor"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/planner/mapper"
)

// split the provided filter into 2 filters based on field.
// It can be used for extracting a supType
// Eg. (filter: {age: 10, name: "bob", author: {birthday: "June 26, 1990", ...}, ...})
//
// In this case the root filter is the conditions that apply to the main type
// ie: {age: 10, name: "bob", ...}.
//
// And the subType filter is the conditions that apply to the queried sub type
// ie: {birthday: "June 26, 1990", ...}.
func SplitFilterByField(filter *mapper.Filter, field mapper.Field) (*mapper.Filter, *mapper.Filter) {
	if filter == nil {
		return nil, nil
	}
	conditionKey := &mapper.PropertyIndex{
		Index: field.Index,
	}

	keyFound, sub := removeConditionIndex(conditionKey, filter.Conditions)
	if !keyFound {
		return filter, nil
	}

	// create new splitup filter
	// our schema ensures that if sub exists, its of type map[string]any
	splitF := &mapper.Filter{
		Conditions:         map[connor.FilterKey]any{conditionKey: sub},
		ExternalConditions: map[string]any{field.Name: filter.ExternalConditions[field.Name]},
	}

	// check if we have any remaining filters
	if len(filter.Conditions) == 0 {
		return nil, splitF
	}
	delete(filter.ExternalConditions, field.Name)
	return filter, splitF
}

func traverseFilterByType(
	key *mapper.PropertyIndex,
	conditions map[connor.FilterKey]any,
) map[connor.FilterKey]any {
	result := make(map[connor.FilterKey]any)
	for targetKey, clause := range conditions {
		if indexKey, isIndexKey := targetKey.(*mapper.PropertyIndex); isIndexKey {
			if key.Index == indexKey.Index {
				result[key] = clause
			}
		} else if opKey, isOpKey := targetKey.(*mapper.Operator); isOpKey {
			clauseArr, isArr := clause.([]any)
			if isArr {
				for _, elementClause := range clauseArr {
					elementMap, ok := elementClause.(map[connor.FilterKey]any)
					if !ok {
						continue
					}
					compoundCond := traverseFilterByType(key, elementMap)
					if len(compoundCond) > 0 {
						resultElement, hasKey := result[opKey]
						if hasKey {
							result[opKey] = append(resultElement.([]any), compoundCond)
						} else {
							result[opKey] = []any{compoundCond}
						}
					}
				}
			}
		}
	}
	return result
}

func CopyFilterTreeNodesForField(filter *mapper.Filter, field mapper.Field) *mapper.Filter {
	if filter == nil {
		return nil
	}
	conditionKey := &mapper.PropertyIndex{
		Index: field.Index,
	}

	resultFilter := &mapper.Filter{}
	conditionMap := traverseFilterByType(conditionKey, filter.Conditions)
	if len(conditionMap) > 0 {
		resultFilter.Conditions = conditionMap
		return resultFilter
	}
	return nil
}

func IsFilterComplex(filter *mapper.Filter) bool {
	if filter == nil {
		return false
	}
	for op, _ := range filter.ExternalConditions {
		if op == "_or" {
			return true
		}
	}
	return false
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
			if op, isOp := existingKey.(*mapper.Operator); isOp && op.Operation == "_and" {
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
					andOp = &mapper.Operator{Operation: "_and"}
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
				if rootOpKey.Operation == "_and" || rootOpKey.Operation == "_or" {
					rootValArr := rootVal.([]any)
					if len(rootValArr) == 1 || rootOpKey.Operation == "_and" && !skipRoot {
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
				} else if rootOpKey.Operation == "_not" {
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
						if opKey, ok := k.(*mapper.Operator); ok && opKey.Operation == "_not" {
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

func NormalizeConditions(conditions map[connor.FilterKey]any) map[connor.FilterKey]any {
	return normalizeConditions(conditions, false).(map[connor.FilterKey]any)
}

func MergeFilterConditions(dest map[connor.FilterKey]any, src map[connor.FilterKey]any) map[connor.FilterKey]any {
	if dest == nil {
		dest = make(map[connor.FilterKey]any)
	}

	result := map[connor.FilterKey]any{
		&mapper.Operator{Operation: "_and"}: []any{
			dest, src,
		},
	}

	return NormalizeConditions(result)
}

func removeConditionIndex(
	key *mapper.PropertyIndex,
	filterConditions map[connor.FilterKey]any,
) (bool, any) {
	for targetKey, clause := range filterConditions {
		if indexKey, isIndexKey := targetKey.(*mapper.PropertyIndex); isIndexKey {
			if key.Index == indexKey.Index {
				delete(filterConditions, targetKey)
				return true, clause
			}
		}
	}
	return false, nil
}

func FilterConditionsToExternal(conditions map[connor.FilterKey]any, mapping *core.DocumentMapping) map[string]any {
	externalConditions := make(map[string]any)

	for key, clause := range conditions {
		var sourceKey string
		var propIndex int
		switch typedKey := key.(type) {
		case *mapper.Operator:
			sourceKey = typedKey.Operation
		case *mapper.PropertyIndex:
			for fieldName, indices := range mapping.IndexesByName {
				for _, index := range indices {
					if index == typedKey.Index {
						sourceKey = fieldName
						propIndex = index
						break
					}
				}
				if sourceKey != "" {
					break
				}
			}
		default:
			continue
		}

		switch typedClause := clause.(type) {
		case []any:
			externalClauses := []any{}
			for _, innerClause := range typedClause {
				extMap, isFilterMap := innerClause.(map[connor.FilterKey]any)
				if !isFilterMap {
					continue
				}
				externalClauses = append(externalClauses, FilterConditionsToExternal(extMap, mapping))
			}
			externalConditions[sourceKey] = externalClauses
		case map[connor.FilterKey]any:
			m := mapping
			if propIndex < len(mapping.ChildMappings) && mapping.ChildMappings[propIndex] != nil {
				m = mapping.ChildMappings[propIndex]
			}
			innerExternalClause := FilterConditionsToExternal(typedClause, m)
			externalConditions[sourceKey] = innerExternalClause
		default:
			externalConditions[sourceKey] = typedClause
		}
	}

	return externalConditions
}

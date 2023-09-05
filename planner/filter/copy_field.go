package filter

import (
	"github.com/sourcenetwork/defradb/connor"
	"github.com/sourcenetwork/defradb/planner/mapper"
)

func traverseFilterByType(
	key *mapper.PropertyIndex,
	conditions map[connor.FilterKey]any,
) map[connor.FilterKey]any {
	result := make(map[connor.FilterKey]any)
	for targetKey, clause := range conditions {
		if targetKey.Equal(key) {
			result[key] = clause
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

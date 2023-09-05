package filter

import (
	"github.com/sourcenetwork/defradb/connor"
	"github.com/sourcenetwork/defradb/planner/mapper"
)

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

func CopyFilterTreeNodesForField(filter *mapper.Filter, field mapper.Field) *mapper.Filter {
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

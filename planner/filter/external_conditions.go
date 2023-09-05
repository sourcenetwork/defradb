package filter

import (
	"github.com/sourcenetwork/defradb/connor"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/planner/mapper"
)

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

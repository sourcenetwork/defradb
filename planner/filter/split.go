package filter

import (
	"github.com/sourcenetwork/defradb/connor"
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
func SplitByField(filter *mapper.Filter, field mapper.Field) (*mapper.Filter, *mapper.Filter) {
	if filter == nil {
		return nil, nil
	}

	splitF := CopyField(filter, field)
	RemoveField(filter, field)

	return filter, splitF
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

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

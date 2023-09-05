package filter

import (
	"github.com/sourcenetwork/defradb/planner/mapper"
)

func RemoveFieldFromFilter(filter *mapper.Filter, field mapper.Field) {
	if filter == nil {
		return
	}
	conditionKey := &mapper.PropertyIndex{
		Index: field.Index,
	}

	traverseFilterByProperty(conditionKey, filter.Conditions, true)
}

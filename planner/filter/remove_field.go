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
	"github.com/sourcenetwork/defradb/planner/mapper"
)

// RemoveField removes the given field from the provided filter.
func RemoveField(filter *mapper.Filter, field mapper.Field) {
	if filter == nil {
		return
	}
	conditionKey := &mapper.PropertyIndex{
		Index: field.Index,
	}

	traverseFilterByProperty(conditionKey, filter.Conditions, true)
}

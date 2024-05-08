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
	"github.com/sourcenetwork/defradb/internal/planner/mapper"
)

// RemoveField removes the given field from the provided filter.
// Multiple fields can be passed to remove related objects with a certain field.
func RemoveField(filter *mapper.Filter, fields ...mapper.Field) {
	if filter == nil || len(fields) == 0 {
		return
	}
	var conditionKeys []*mapper.PropertyIndex
	for _, field := range fields {
		conditionKeys = append(conditionKeys, &mapper.PropertyIndex{
			Index: field.Index,
		})
	}

	traverseFilterByProperty(conditionKeys, filter.Conditions, true)
}

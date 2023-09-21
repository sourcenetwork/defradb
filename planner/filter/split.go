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

// SplitByField splits the provided filter into 2 filters based on field.
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

	if len(filter.Conditions) == 0 {
		filter = nil
	}

	return filter, splitF
}

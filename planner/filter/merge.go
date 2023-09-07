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
	"github.com/sourcenetwork/defradb/connor"
	"github.com/sourcenetwork/defradb/planner/mapper"
)

// Merge merges two filters into one.
// It basically applies _and to both filters and normalizes them.
func Merge(dest map[connor.FilterKey]any, src map[connor.FilterKey]any) map[connor.FilterKey]any {
	if dest == nil {
		dest = make(map[connor.FilterKey]any)
	}

	result := map[connor.FilterKey]any{
		&mapper.Operator{Operation: andID}: []any{
			dest, src,
		},
	}

	return Normalize(result)
}

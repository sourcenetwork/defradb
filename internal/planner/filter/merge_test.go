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
	"testing"

	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/internal/connor"
	"github.com/sourcenetwork/defradb/internal/planner/mapper"
)

func TestMergeFilterConditions(t *testing.T) {
	tests := []struct {
		name     string
		left     map[string]any
		right    map[string]any
		expected map[string]any
	}{
		{
			name: "basic merge",
			left: map[string]any{
				"name":     m("_eq", "John"),
				"verified": m("_eq", true),
			},
			right: map[string]any{
				"age": m("_gt", 55),
			},
			expected: map[string]any{
				"name":     m("_eq", "John"),
				"verified": m("_eq", true),
				"age":      m("_gt", 55),
			},
		},
		{
			name: "basic _and merge",
			left: m("_and", []any{
				m("name", m("_eq", "John")),
			}),
			right: m("_and", []any{
				m("age", m("_gt", 55)),
			}),
			expected: map[string]any{
				"name": m("_eq", "John"),
				"age":  m("_gt", 55),
			},
		},
	}

	mapping := getDocMapping()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			leftFilter := mapper.ToFilter(request.Filter{Conditions: tt.left}, mapping)
			rightFilter := mapper.ToFilter(request.Filter{Conditions: tt.right}, mapping)
			actualFilter := Merge(leftFilter.Conditions, rightFilter.Conditions)
			expectedFilter := mapper.ToFilter(request.Filter{Conditions: tt.expected}, mapping)
			AssertEqualFilterMap(t, expectedFilter.Conditions, actualFilter)
		})
	}
}

func TestMergeNullFilter(t *testing.T) {
	f := map[connor.FilterKey]any{
		&mapper.PropertyIndex{Index: 0}: "value1",
	}
	AssertEqualFilterMap(t, f, Merge(f, nil))
	AssertEqualFilterMap(t, f, Merge(nil, f))
}

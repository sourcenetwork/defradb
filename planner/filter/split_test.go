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
	"github.com/sourcenetwork/defradb/planner/mapper"
	"github.com/stretchr/testify/assert"
)

func TestSplitFilter(t *testing.T) {
	tests := []struct {
		name            string
		inputField      mapper.Field
		inputFilter     map[string]any
		expectedFilter1 map[string]any
		expectedFilter2 map[string]any
	}{
		{
			name: "flat structure",
			inputFilter: map[string]any{
				"name": m("_eq", "John"),
				"age":  m("_gt", 55),
			},
			inputField:      mapper.Field{Index: 1}, // age
			expectedFilter1: m("name", m("_eq", "John")),
			expectedFilter2: m("age", m("_gt", 55)),
		},
	}

	mapping := getDocMapping()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			inputFilter := mapper.ToFilter(request.Filter{Conditions: test.inputFilter}, mapping)
			actualFilter1, actualFilter2 := SplitByField(inputFilter, test.inputField)
			expectedFilter1 := mapper.ToFilter(request.Filter{Conditions: test.expectedFilter1}, mapping)
			expectedFilter2 := mapper.ToFilter(request.Filter{Conditions: test.expectedFilter2}, mapping)
			AssertEqualFilterMap(t, expectedFilter1.Conditions, actualFilter1.Conditions)
			AssertEqualFilterMap(t, expectedFilter2.Conditions, actualFilter2.Conditions)
		})
	}
}

func TestSplitNullFilter(t *testing.T) {
	actualFilter1, actualFilter2 := SplitByField(nil, mapper.Field{Index: 1})
	assert.Nil(t, actualFilter1)
	assert.Nil(t, actualFilter2)
}

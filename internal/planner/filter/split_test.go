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
	"github.com/sourcenetwork/defradb/internal/planner/mapper"

	"github.com/stretchr/testify/assert"
)

func TestSplitFilter(t *testing.T) {
	tests := []struct {
		name            string
		inputFields     []mapper.Field
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
			inputFields:     []mapper.Field{{Index: authorAgeInd}},
			expectedFilter1: m("name", m("_eq", "John")),
			expectedFilter2: m("age", m("_gt", 55)),
		},
		{
			name: "the only field",
			inputFilter: map[string]any{
				"age": m("_gt", 55),
			},
			inputFields:     []mapper.Field{{Index: authorAgeInd}},
			expectedFilter1: nil,
			expectedFilter2: m("age", m("_gt", 55)),
		},
		{
			name: "no field to delete",
			inputFilter: map[string]any{
				"name": m("_eq", "John"),
			},
			inputFields:     []mapper.Field{{Index: authorAgeInd}},
			expectedFilter1: m("name", m("_eq", "John")),
			expectedFilter2: nil,
		},
		{
			name: "split by 2 fields",
			inputFilter: map[string]any{
				"name":      m("_eq", "John"),
				"age":       m("_gt", 55),
				"published": m("_eq", true),
				"verified":  m("_eq", false),
			},
			inputFields:     []mapper.Field{{Index: authorNameInd}, {Index: authorAgeInd}, {Index: authorVerifiedInd}},
			expectedFilter1: m("published", m("_eq", true)),
			expectedFilter2: map[string]any{
				"name":     m("_eq", "John"),
				"age":      m("_gt", 55),
				"verified": m("_eq", false),
			},
		},
		{
			name: "split by fields that are not present",
			inputFilter: map[string]any{
				"name":     m("_eq", "John"),
				"age":      m("_gt", 55),
				"verified": m("_eq", false),
			},
			inputFields: []mapper.Field{
				{Index: authorNameInd},
				{Index: 100},
				{Index: authorAgeInd},
				{Index: 200},
			},
			expectedFilter1: m("verified", m("_eq", false)),
			expectedFilter2: map[string]any{
				"name": m("_eq", "John"),
				"age":  m("_gt", 55),
			},
		},
		{
			name: "filter with two []any slices",
			inputFilter: map[string]any{
				"age":  m("_in", []any{10, 20, 30}),
				"name": m("_in", []any{"John", "Bob"}),
			},
			inputFields: []mapper.Field{
				{Index: authorNameInd},
				{Index: authorAgeInd},
			},
			expectedFilter1: nil,
			expectedFilter2: map[string]any{
				"age":  m("_in", []any{10, 20, 30}),
				"name": m("_in", []any{"John", "Bob"}),
			},
		},
	}

	mapping := getDocMapping()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			inputFilter := mapper.ToFilter(request.Filter{Conditions: test.inputFilter}, mapping)
			actualFilter1, actualFilter2 := SplitByFields(inputFilter, test.inputFields...)
			expectedFilter1 := mapper.ToFilter(request.Filter{Conditions: test.expectedFilter1}, mapping)
			expectedFilter2 := mapper.ToFilter(request.Filter{Conditions: test.expectedFilter2}, mapping)
			if expectedFilter1 != nil || actualFilter1 != nil {
				AssertEqualFilterMap(t, expectedFilter1.Conditions, actualFilter1.Conditions)
			}
			if expectedFilter2 != nil || actualFilter2 != nil {
				AssertEqualFilterMap(t, expectedFilter2.Conditions, actualFilter2.Conditions)
			}
		})
	}
}

func TestSplitFilter_WithNoFields_ReturnsInputFilter(t *testing.T) {
	mapping := getDocMapping()
	inputFilterConditions := map[string]any{
		"name": m("_eq", "John"),
		"age":  m("_gt", 55),
	}
	inputFilter := mapper.ToFilter(request.Filter{Conditions: inputFilterConditions}, mapping)
	actualFilter1, actualFilter2 := SplitByFields(inputFilter)
	AssertEqualFilterMap(t, inputFilter.Conditions, actualFilter1.Conditions)
	assert.Nil(t, actualFilter2)
}

func TestSplitNullFilter(t *testing.T) {
	actualFilter1, actualFilter2 := SplitByFields(nil, mapper.Field{Index: authorAgeInd})
	assert.Nil(t, actualFilter1)
	assert.Nil(t, actualFilter2)
}

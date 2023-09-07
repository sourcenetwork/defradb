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
	"reflect"
	"testing"

	"github.com/sourcenetwork/defradb/connor"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/planner/mapper"
)

func TestFilterConditionsToExternal(t *testing.T) {
	tests := []struct {
		name            string
		inputConditions map[connor.FilterKey]any
		inputMapping    *core.DocumentMapping
		expectedOutput  map[string]any
	}{
		{
			name: "Test single condition",
			inputConditions: map[connor.FilterKey]any{
				&mapper.PropertyIndex{Index: 0}: "value1",
			},
			inputMapping: &core.DocumentMapping{
				IndexesByName: map[string][]int{"field1": {0}},
			},
			expectedOutput: map[string]any{"field1": "value1"},
		},
		{
			name: "Test multiple conditions",
			inputConditions: map[connor.FilterKey]any{
				&mapper.PropertyIndex{Index: 0}: "value1",
				&mapper.PropertyIndex{Index: 1}: "value2",
			},
			inputMapping: &core.DocumentMapping{
				IndexesByName: map[string][]int{"field1": {0}, "field2": {1}},
			},
			expectedOutput: map[string]any{"field1": "value1", "field2": "value2"},
		},
		{
			name: "Test operator",
			inputConditions: map[connor.FilterKey]any{
				&mapper.Operator{Operation: "_eq"}: "value",
			},
			inputMapping:   &core.DocumentMapping{},
			expectedOutput: map[string]any{"_eq": "value"},
		},
		{
			name: "Test complex condition",
			inputConditions: map[connor.FilterKey]any{
				&mapper.Operator{Operation: "_or"}: []any{
					map[connor.FilterKey]any{
						&mapper.PropertyIndex{Index: 0}: map[connor.FilterKey]any{
							&mapper.Operator{Operation: "_eq"}: "Some name",
						},
					},
					map[connor.FilterKey]any{
						&mapper.Operator{Operation: "_and"}: []any{
							map[connor.FilterKey]any{
								&mapper.PropertyIndex{Index: 1}: map[connor.FilterKey]any{
									&mapper.Operator{Operation: "_gt"}: 64,
								},
							},
							map[connor.FilterKey]any{
								&mapper.PropertyIndex{Index: 2}: map[connor.FilterKey]any{
									&mapper.PropertyIndex{Index: 1}: map[connor.FilterKey]any{
										&mapper.Operator{Operation: "_gt"}: 4.8,
									},
								},
							},
						},
					},
					map[connor.FilterKey]any{
						&mapper.Operator{Operation: "_and"}: []any{
							map[connor.FilterKey]any{
								&mapper.PropertyIndex{Index: 1}: map[connor.FilterKey]any{
									&mapper.Operator{Operation: "_lt"}: 64,
								},
							},
							map[connor.FilterKey]any{
								&mapper.PropertyIndex{Index: 2}: map[connor.FilterKey]any{
									&mapper.PropertyIndex{Index: 1}: map[connor.FilterKey]any{
										&mapper.Operator{Operation: "_lt"}: 4.8,
									},
								},
							},
						},
					},
				},
			},
			inputMapping: &core.DocumentMapping{
				IndexesByName: map[string][]int{"name": {0}, "age": {1}, "published": {2}, "rating": {3}},
				ChildMappings: []*core.DocumentMapping{nil, nil, {
					IndexesByName: map[string][]int{"rating": {1}},
				}},
			},
			expectedOutput: map[string]any{
				"_or": []any{
					map[string]any{"name": map[string]any{"_eq": "Some name"}},

					map[string]any{"_and": []any{
						map[string]any{"age": map[string]any{"_gt": 64}},
						map[string]any{"published": map[string]any{"rating": map[string]any{"_gt": 4.8}}},
					}},

					map[string]any{"_and": []any{
						map[string]any{"age": map[string]any{"_lt": 64}},
						map[string]any{"published": map[string]any{"rating": map[string]any{"_lt": 4.8}}},
					}},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output := BuildExternalConditions(test.inputConditions, test.inputMapping)
			if !reflect.DeepEqual(output, test.expectedOutput) {
				t.Errorf("Expected %v, but got %v", test.expectedOutput, output)
			}
		})
	}
}

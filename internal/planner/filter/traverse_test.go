// Copyright 2025 Democratized Data Foundation
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

	"github.com/stretchr/testify/assert"

	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/internal/connor"
	"github.com/sourcenetwork/defradb/internal/planner/mapper"
)

func TestTraverseFieldOperators(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		expected [][2]string
	}{
		{
			name:     "simple field",
			input:    map[string]any{"name": map[string]any{"_eq": "John"}},
			expected: [][2]string{{"name", "_eq"}},
		},
		{
			name: "multiple fields",
			input: map[string]any{
				"name": map[string]any{"_eq": "John"},
				"age":  map[string]any{"_gt": 25},
			},
			expected: [][2]string{{"name", "_eq"}, {"age", "_gt"}},
		},
		{
			name: "field of a related collection reports no operator",
			input: map[string]any{
				"author": map[string]any{
					"books": map[string]any{"title": map[string]any{"_eq": "Sample"}},
				},
			},
			expected: [][2]string{{"author", OpRelatedFilter}},
		},
		{
			name: "with _or operator",
			input: map[string]any{
				request.FilterOpOr: []any{
					map[string]any{"name": map[string]any{"_eq": "John"}},
					map[string]any{"age": map[string]any{"_gt": 30}},
				},
			},
			expected: [][2]string{{"name", "_eq"}, {"age", "_gt"}},
		},
		{
			name: "with _and operator",
			input: map[string]any{
				request.FilterOpAnd: []any{
					map[string]any{"name": map[string]any{"_like": "%Jo%"}},
				},
			},
			expected: [][2]string{{"name", "_like"}},
		},
		{
			name: "with alias operator",
			input: map[string]any{
				request.AliasFieldName: map[string]any{
					"age": map[string]any{"_eq": 30},
				},
			},
			expected: [][2]string{{"age", "_eq"}},
		},
		{
			name: "conditions under _not are skipped",
			input: map[string]any{
				request.FilterOpNot: map[string]any{
					"name": map[string]any{"_eq": "John"},
				},
				"age": map[string]any{"_gt": 30},
			},
			expected: [][2]string{{"age", "_gt"}},
		},
		{
			name:     "_docID is a field, not an operator",
			input:    map[string]any{request.DocIDFieldName: map[string]any{"_eq": "bae-1"}},
			expected: [][2]string{{request.DocIDFieldName, "_eq"}},
		},
		{
			name:     "operator with nil value",
			input:    map[string]any{request.FilterOpOr: nil},
			expected: nil,
		},
		{
			name:     "operator with value of invalid type",
			input:    map[string]any{request.FilterOpAnd: 1},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var actual [][2]string
			TraverseFieldOperators("", tt.input, func(field, op string) {
				actual = append(actual, [2]string{field, op})
			})

			assert.ElementsMatch(t, tt.expected, actual)
		})
	}
}

func TestTraverseProperties(t *testing.T) {
	tests := []struct {
		name           string
		input          map[connor.FilterKey]any
		expectedProps  []int
		expectedValues map[int]any
	}{
		{
			name: "simple property",
			input: map[connor.FilterKey]any{
				&mapper.PropertyIndex{Index: 1}: map[connor.FilterKey]any{
					&mapper.Operator{Operation: "_eq"}: "John",
				},
			},
			expectedProps:  []int{1},
			expectedValues: map[int]any{1: "John"},
		},
		{
			name: "multiple properties",
			input: map[connor.FilterKey]any{
				&mapper.PropertyIndex{Index: 1}: map[connor.FilterKey]any{
					&mapper.Operator{Operation: "_eq"}: "John",
				},
				&mapper.PropertyIndex{Index: 2}: map[connor.FilterKey]any{
					&mapper.Operator{Operation: "_gt"}: 25,
				},
			},
			expectedProps:  []int{1, 2},
			expectedValues: map[int]any{1: "John", 2: 25},
		},
		{
			name: "nested in operator",
			input: map[connor.FilterKey]any{
				&mapper.Operator{Operation: "_or"}: map[connor.FilterKey]any{
					&mapper.PropertyIndex{Index: 1}: map[connor.FilterKey]any{
						&mapper.Operator{Operation: "_eq"}: "John",
					},
					&mapper.PropertyIndex{Index: 2}: map[connor.FilterKey]any{
						&mapper.Operator{Operation: "_gt"}: 25,
					},
				},
			},
			expectedProps:  []int{1, 2},
			expectedValues: map[int]any{1: "John", 2: 25},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var actualProps []int
			actualValues := make(map[int]any)

			TraverseProperties(tt.input, func(prop *mapper.PropertyIndex, val map[connor.FilterKey]any) bool {
				actualProps = append(actualProps, prop.Index)
				// Extract the actual value from the operator map
				for _, v := range val {
					actualValues[prop.Index] = v
					break // We only expect one operator per property in our test cases
				}
				return true
			})

			assert.ElementsMatch(t, tt.expectedProps, actualProps)
			assert.Equal(t, tt.expectedValues, actualValues)
		})
	}
}

func TestTraverseProperties_EarlyExit(t *testing.T) {
	input := map[connor.FilterKey]any{
		&mapper.Operator{Operation: "_and"}: map[connor.FilterKey]any{
			&mapper.PropertyIndex{Index: 1}: map[connor.FilterKey]any{
				&mapper.Operator{Operation: "_eq"}: "John",
			},
			&mapper.PropertyIndex{Index: 2}: map[connor.FilterKey]any{
				&mapper.Operator{Operation: "_gt"}: 25,
			},
			&mapper.PropertyIndex{Index: 3}: map[connor.FilterKey]any{
				&mapper.Operator{Operation: "_eq"}: "New York",
			},
		},
	}

	var visitCount int
	TraverseProperties(input, func(prop *mapper.PropertyIndex, val map[connor.FilterKey]any) bool {
		visitCount++
		return visitCount < 2 // Stop after visiting 2 properties
	})

	assert.Equal(t, 2, visitCount, "should have stopped after visiting 2 properties")
}

func TestTraversePropertiesWithIgnoreNodes(t *testing.T) {
	tests := []struct {
		name       string
		conditions map[connor.FilterKey]any
		skipOps    []string
		expected   []int
	}{
		{
			name: "ignore _not operator",
			conditions: map[connor.FilterKey]any{
				&mapper.Operator{Operation: "_not"}: map[connor.FilterKey]any{
					&mapper.PropertyIndex{Index: 1}: map[connor.FilterKey]any{
						&mapper.Operator{Operation: "_eq"}: "John",
					},
				},
				&mapper.PropertyIndex{Index: 2}: map[connor.FilterKey]any{
					&mapper.Operator{Operation: "_gt"}: 18,
				},
			},
			skipOps:  []string{"_not"},
			expected: []int{2},
		},
		{
			name: "ignore multiple operators",
			conditions: map[connor.FilterKey]any{
				&mapper.Operator{Operation: "_not"}: map[connor.FilterKey]any{
					&mapper.PropertyIndex{Index: 1}: map[connor.FilterKey]any{
						&mapper.Operator{Operation: "_eq"}: "John",
					},
				},
				&mapper.Operator{Operation: "_and"}: map[connor.FilterKey]any{
					&mapper.PropertyIndex{Index: 2}: map[connor.FilterKey]any{
						&mapper.Operator{Operation: "_eq"}: true,
					},
				},
				&mapper.Operator{Operation: "_or"}: map[connor.FilterKey]any{
					&mapper.PropertyIndex{Index: 3}: map[connor.FilterKey]any{
						&mapper.Operator{Operation: "_gt"}: 18,
					},
				},
			},
			skipOps:  []string{"_not", "_or"},
			expected: []int{2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var visited []int
			TraverseProperties(tt.conditions, func(p *mapper.PropertyIndex, conditions map[connor.FilterKey]any) bool {
				visited = append(visited, p.Index)
				return true
			}, tt.skipOps...)

			assert.ElementsMatch(t, tt.expected, visited)
		})
	}
}

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

	"github.com/sourcenetwork/defradb/internal/connor"
	"github.com/sourcenetwork/defradb/internal/planner/mapper"
	"github.com/stretchr/testify/assert"
)

func TestTraverseFields(t *testing.T) {
	tests := []struct {
		name          string
		input         map[string]any
		expectedPaths [][]string
		expectedVals  []any
	}{
		{
			name: "simple field",
			input: map[string]any{
				"name": map[string]any{
					"_eq": "John",
				},
			},
			expectedPaths: [][]string{{"name"}},
			expectedVals:  []any{"John"},
		},
		{
			name: "multiple fields",
			input: map[string]any{
				"name": map[string]any{"_eq": "John"},
				"age":  map[string]any{"_gt": 25},
			},
			expectedPaths: [][]string{{"name"}, {"age"}},
			expectedVals:  []any{"John", 25},
		},
		{
			name: "nested fields",
			input: map[string]any{
				"author": map[string]any{
					"books": map[string]any{
						"title": map[string]any{
							"_eq": "Sample Book",
						},
					},
				},
			},
			expectedPaths: [][]string{{"author", "books", "title"}},
			expectedVals:  []any{"Sample Book"},
		},
		{
			name: "with array operator",
			input: map[string]any{
				"_or": []any{
					map[string]any{
						"name": map[string]any{"_eq": "John"},
					},
					map[string]any{
						"age": map[string]any{"_gt": 30},
					},
				},
			},
			expectedPaths: [][]string{{"name"}, {"age"}},
			expectedVals:  []any{"John", 30},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var actualPaths [][]string
			var actualVals []any

			TraverseFields(tt.input, func(path []string, val any) bool {
				pathCopy := make([]string, len(path))
				copy(pathCopy, path)
				actualPaths = append(actualPaths, pathCopy)
				actualVals = append(actualVals, val)
				return true // continue traversal
			})

			assert.ElementsMatch(t, tt.expectedPaths, actualPaths)
			assert.ElementsMatch(t, tt.expectedVals, actualVals)
		})
	}
}

func TestTraverseFieldsEarlyExit(t *testing.T) {
	tests := []struct {
		name          string
		input         map[string]any
		expectedCount int
		exitAfter     int
	}{
		{
			name: "exit in flat fields",
			input: map[string]any{
				"name": map[string]any{"_eq": "John"},
				"age":  map[string]any{"_gt": 25},
				"city": map[string]any{"_eq": "New York"},
			},
			expectedCount: 2,
			exitAfter:     2,
		},
		{
			name: "exit in nested fields",
			input: map[string]any{
				"author": map[string]any{
					"name": map[string]any{"_eq": "John"},
					"books": map[string]any{
						"title": map[string]any{"_eq": "Book 1"},
						"year":  map[string]any{"_gt": 2000},
					},
				},
			},
			expectedCount: 1,
			exitAfter:     1,
		},
		{
			name: "exit in array operator",
			input: map[string]any{
				"_or": []any{
					map[string]any{
						"name": map[string]any{"_eq": "John"},
					},
					map[string]any{
						"age": map[string]any{"_gt": 30},
					},
					map[string]any{
						"city": map[string]any{"_eq": "Paris"},
					},
				},
			},
			expectedCount: 2,
			exitAfter:     2,
		},
		{
			name: "exit in mixed operators",
			input: map[string]any{
				"_and": []any{
					map[string]any{
						"name": map[string]any{"_eq": "John"},
					},
					map[string]any{
						"_or": []any{
							map[string]any{"age": map[string]any{"_gt": 30}},
							map[string]any{"city": map[string]any{"_eq": "Paris"}},
						},
					},
				},
			},
			expectedCount: 1,
			exitAfter:     1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var actualPaths [][]string
			var actualVals []any

			TraverseFields(tt.input, func(path []string, val any) bool {
				pathCopy := make([]string, len(path))
				copy(pathCopy, path)
				actualPaths = append(actualPaths, pathCopy)
				actualVals = append(actualVals, val)
				return len(actualPaths) < tt.exitAfter
			})

			assert.Equal(t, tt.expectedCount, len(actualPaths),
				"should have stopped after %d fields", tt.expectedCount)
			assert.Equal(t, tt.expectedCount, len(actualVals),
				"should have stopped after %d values", tt.expectedCount)
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

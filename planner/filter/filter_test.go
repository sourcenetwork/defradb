package filter

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/connor"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/planner/mapper"
)

func assertEqualFilterMap(expected, actual map[connor.FilterKey]any, prefix string) string {
	if len(expected) != len(actual) {
		return fmt.Sprintf("Mismatch at %s: Expected map length: %d, but got: %d", prefix, len(expected), len(actual))
	}

	findMatchingKey := func(key connor.FilterKey, m map[connor.FilterKey]any) connor.FilterKey {
		for k := range m {
			if k.Equal(key) {
				return k
			}
		}
		return nil
	}

	for expKey, expVal := range expected {
		actKey := findMatchingKey(expKey, actual)
		if actKey == nil {
			return fmt.Sprintf("Mismatch at %s: Expected key %v not found in actual map", prefix, expKey)
		}
		actVal := actual[actKey]

		newPrefix := fmt.Sprintf("%s.%v", prefix, expKey)
		switch expTypedVal := expVal.(type) {
		case map[connor.FilterKey]any:
			actTypedVal, ok := actVal.(map[connor.FilterKey]any)
			if !ok {
				return fmt.Sprintf("Mismatch at %s: Expected a nested map[FilterKey]any for key %v, but got: %v", prefix, expKey, actVal)
			}
			errMsg := assertEqualFilterMap(expTypedVal, actTypedVal, newPrefix)
			if errMsg != "" {
				return errMsg
			}
		case []any:
			actTypedVal, ok := actVal.([]any)
			if !ok {
				return fmt.Sprintf("Mismatch at %s: Expected a nested []any for key %v, but got: %v", newPrefix, expKey, actVal)
			}
			if len(expTypedVal) != len(actTypedVal) {
				return fmt.Sprintf("Mismatch at %s: Expected slice length: %d, but got: %d", newPrefix, len(expTypedVal), len(actTypedVal))
			}
			numElements := len(expTypedVal)
			for i := 0; i < numElements; i++ {
				for j := 0; j < numElements; j++ {
					errMsg := compareElements(expTypedVal[i], actTypedVal[j], expKey, newPrefix)
					if errMsg == "" {
						actTypedVal = append(actTypedVal[:j], actTypedVal[j+1:]...)
						break
					}
				}
				if len(actTypedVal) != numElements-i-1 {
					return fmt.Sprintf("Mismatch at %s: Expected element not found: %d", newPrefix, expTypedVal[i])
				}
			}
		default:
			if !reflect.DeepEqual(expVal, actVal) {
				return fmt.Sprintf("Mismatch at %s: Expected value %v for key %v, but got %v", prefix, expVal, expKey, actVal)
			}
		}
	}
	return ""
}

func compareElements(expected, actual any, key connor.FilterKey, prefix string) string {
	switch expElem := expected.(type) {
	case map[connor.FilterKey]any:
		actElem, ok := actual.(map[connor.FilterKey]any)
		if !ok {
			return fmt.Sprintf("Mismatch at %s: Expected a nested map[FilterKey]any for key %v, but got: %v", prefix, key, actual)
		}
		return assertEqualFilterMap(expElem, actElem, prefix)
	default:
		if !reflect.DeepEqual(expElem, actual) {
			return fmt.Sprintf("Mismatch at %s: Expected value %v for key %v, but got %v", prefix, expElem, key, actual)
		}
	}
	return ""
}

func AssertEqualFilterMap(t *testing.T, expected, actual map[connor.FilterKey]any) {
	errMsg := assertEqualFilterMap(expected, actual, "root")
	if errMsg != "" {
		t.Fatal(errMsg)
	}
}

func AssertEqualFilter(t *testing.T, expected, actual *mapper.Filter) {
	if expected == nil && actual == nil {
		return
	}

	if expected == nil || actual == nil {
		t.Fatalf("Expected %v, but got %v", expected, actual)
		return
	}

	AssertEqualFilterMap(t, expected.Conditions, actual.Conditions)

	if !reflect.DeepEqual(expected.ExternalConditions, actual.ExternalConditions) {
		t.Errorf("Expected external conditions \n\t%v\n, but got \n\t%v",
			expected.ExternalConditions, actual.ExternalConditions)
	}
}

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
			output := FilterConditionsToExternal(test.inputConditions, test.inputMapping)
			if !reflect.DeepEqual(output, test.expectedOutput) {
				t.Errorf("Expected %v, but got %v", test.expectedOutput, output)
			}
		})
	}
}

func TestCopyFilterTreeNodesForField(t *testing.T) {
	tests := []struct {
		name        string
		inputFilter *mapper.Filter
		inputField  mapper.Field
		expected    *mapper.Filter
	}{
		{
			name: "Basic Test",
			inputFilter: &mapper.Filter{
				Conditions: map[connor.FilterKey]any{
					&mapper.PropertyIndex{Index: 1}: map[connor.FilterKey]any{
						&mapper.Operator{Operation: "_eq"}: "value1",
					},
					&mapper.PropertyIndex{Index: 2}: map[connor.FilterKey]any{
						&mapper.Operator{Operation: "_eq"}: "value2",
					},
				},
			},
			inputField: mapper.Field{Index: 1},
			expected: &mapper.Filter{
				Conditions: map[connor.FilterKey]any{
					&mapper.PropertyIndex{Index: 1}: map[connor.FilterKey]any{
						&mapper.Operator{Operation: "_eq"}: "value1",
					},
				},
			},
		},
		{
			name: "With Operator",
			inputFilter: &mapper.Filter{
				Conditions: map[connor.FilterKey]any{
					&mapper.Operator{Operation: "_and"}: []any{
						map[connor.FilterKey]any{
							&mapper.PropertyIndex{Index: 1}: map[connor.FilterKey]any{
								&mapper.Operator{Operation: "_eq"}: "value1",
							},
							&mapper.PropertyIndex{Index: 2}: map[connor.FilterKey]any{
								&mapper.Operator{Operation: "_eq"}: "value2",
							},
						},
					},
				},
			},
			inputField: mapper.Field{Index: 1},
			expected: &mapper.Filter{
				Conditions: map[connor.FilterKey]any{
					&mapper.Operator{Operation: "_and"}: []any{
						map[connor.FilterKey]any{
							&mapper.PropertyIndex{Index: 1}: map[connor.FilterKey]any{
								&mapper.Operator{Operation: "_eq"}: "value1",
							},
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output := CopyFilterTreeNodesForField(test.inputFilter, test.inputField)
			AssertEqualFilter(t, test.expected, output)
		})
	}
}

func TestNormalizeConditions(t *testing.T) {
	m := func(op string, val any) map[string]any {
		return map[string]any{op: val}
	}
	r := func(op string, vals ...any) map[string]any {
		return m(op, vals)
	}

	tests := []struct {
		name     string
		input    map[string]any
		expected map[string]any
	}{
		{
			name: "don't normalize already normalized conditions",
			input: map[string]any{
				"name":     m("_eq", "John"),
				"verified": m("_eq", true),
			},
			expected: map[string]any{
				"name":     m("_eq", "John"),
				"verified": m("_eq", true),
			},
		},
		{
			name: "flatten single _and condition",
			input: r("_and",
				m("name", m("_eq", "John")),
				m("verified", m("_eq", true)),
			),
			expected: map[string]any{
				"name":     m("_eq", "John"),
				"verified": m("_eq", true),
			},
		},
		{
			name: "don't touch single _or condition",
			input: r("_or",
				m("name", m("_eq", "John")),
				m("verified", m("_eq", true)),
			),
			expected: r("_or",
				m("name", m("_eq", "John")),
				m("verified", m("_eq", true)),
			),
		},
		{
			name: "flatten _and with single condition",
			input: map[string]any{
				"_and": []any{
					m("name", m("_eq", "John")),
				},
				"verified": m("_eq", true),
			},
			expected: map[string]any{
				"name":     m("_eq", "John"),
				"verified": m("_eq", true),
			},
		},
		{
			name: "flatten _or with single condition",
			input: map[string]any{
				"_or": []any{
					m("name", m("_eq", "John")),
				},
				"verified": m("_eq", true),
			},
			expected: map[string]any{
				"name":     m("_eq", "John"),
				"verified": m("_eq", true),
			},
		},
		{
			name: "flatten long _and/_or chain",
			input: r("_or", r("_and", r("_or", r("_or", r("_and", r("_and", r("_and",
				m("name", m("_eq", "John")),
				m("verified", m("_eq", true)),
			))))))),
			expected: map[string]any{
				"name":     m("_eq", "John"),
				"verified": m("_eq", true),
			},
		},
		{
			name: "normalize sibling _and with few conditions",
			input: map[string]any{
				"_and": []any{
					r("_and",
						m("age", m("_gt", 30)),
						m("published", m("rating", m("_lt", 4.8))),
					),
					r("_and", m("verified", m("_eq", true))),
				},
				"name": m("_eq", "John"),
			},
			expected: map[string]any{
				"name":      m("_eq", "John"),
				"published": m("rating", m("_lt", 4.8)),
				"age":       m("_gt", 30),
				"verified":  m("_eq", true),
			},
		},
		{
			name:     "don't touch single _not",
			input:    m("_not", m("name", m("_eq", "John"))),
			expected: m("_not", m("name", m("_eq", "John"))),
		},
		{
			name:     "remove double _not",
			input:    m("_not", m("_not", m("name", m("_eq", "John")))),
			expected: m("name", m("_eq", "John")),
		},
		{
			name: "remove double _not (sibling)",
			input: map[string]any{
				"_not": m("_not", m("name", m("_eq", "John"))),
				"age":  m("_eq", 65),
			},
			expected: map[string]any{
				"name": m("_eq", "John"),
				"age":  m("_eq", 65),
			},
		},
		{
			name: "don't touch double _not if first has few elements",
			input: m("_not", map[string]any{
				"_not":     m("name", m("_eq", "John")),
				"verified": m("_eq", true),
			}),
			expected: m("_not", map[string]any{
				"_not":     m("name", m("_eq", "John")),
				"verified": m("_eq", true),
			}),
		},
		{
			name:     "normalize long _not chain",
			input:    m("_not", m("_not", m("_not", m("_not", m("_not", m("name", m("_eq", "John"))))))),
			expected: m("_not", m("name", m("_eq", "John"))),
		},
		{
			name: "normalize _not content",
			input: m("_not", r("_and",
				m("name", m("_eq", "John")),
				r("_and",
					m("age", m("_eq", 30)),
					m("verified", m("_eq", true)),
				),
			)),
			expected: m("_not", r("_and",
				m("name", m("_eq", "John")),
				m("age", m("_eq", 30)),
				m("verified", m("_eq", true)),
			)),
		},
		{
			name:     "normalize long _not,_and,_or chain",
			input:    m("_not", r("_and", m("_not", r("_or", m("_not", m("name", m("_eq", "John"))))))),
			expected: m("_not", m("name", m("_eq", "John"))),
		},
		{
			name: "normalize nested arr elements",
			input: r("_and",
				r("_and", r("_and", m("name", m("_eq", "John")))),
				r("_and", m("verified", m("_eq", true))),
				r("_and", r("_and",
					r("_and", m("age", m("_lt", 55))),
					m("published", m("rating", m("_gt", 4.4))),
				)),
			),
			expected: map[string]any{
				"name":      m("_eq", "John"),
				"verified":  m("_eq", true),
				"age":       m("_lt", 55),
				"published": m("rating", m("_gt", 4.4)),
			},
		},
		{
			name: "do not flatten _and, child of _or",
			input: r("_or",
				r("_and",
					m("name", m("_eq", "John")),
					m("verified", m("_eq", true)),
				),
				r("_and",
					m("name", m("_eq", "Islam")),
					m("verified", m("_eq", false)),
				),
			),
			expected: r("_or",
				r("_and",
					m("name", m("_eq", "John")),
					m("verified", m("_eq", true)),
				),
				r("_and",
					m("name", m("_eq", "Islam")),
					m("verified", m("_eq", false)),
				),
			),
		},
		{
			name: "flatten _and, grand children of _or",
			input: r("_or",
				r("_and",
					r("_and",
						m("name", m("_eq", "Islam")),
						m("age", m("_eq", "30")),
					),
					m("verified", m("_eq", false)),
				),
				r("_and",
					m("name", m("_eq", "John")),
					m("verified", m("_eq", true)),
				),
			),
			expected: r("_or",
				r("_and",
					m("name", m("_eq", "Islam")),
					m("age", m("_eq", "30")),
					m("verified", m("_eq", false)),
				),
				r("_and",
					m("name", m("_eq", "John")),
					m("verified", m("_eq", true)),
				),
			),
		},
		{
			name: "squash same keys into _and",
			input: map[string]any{
				"_and": []any{
					r("_and",
						m("age", m("_gt", 30)),
						m("published", m("rating", m("_lt", 4.8))),
					),
					r("_and", m("age", m("_lt", 55))),
					m("age", m("_ne", 33)),
				},
				"name": m("_eq", "John"),
			},
			expected: map[string]any{
				"name":      m("_eq", "John"),
				"published": m("rating", m("_lt", 4.8)),
				"_and": []any{
					m("age", m("_gt", 30)),
					m("age", m("_lt", 55)),
					m("age", m("_ne", 33)),
				},
			},
		},
		{
			name: "squash same keys into _and (with more matching keys)",
			input: map[string]any{
				"_and": []any{
					m("published", m("rating", m("_lt", 4.8))),
					r("_and", m("name", m("_ne", "Islam"))),
					r("_and",
						m("age", m("_gt", 30)),
						m("published", m("genre", m("_eq", "Thriller"))),
						m("verified", m("_eq", true)),
					),
					r("_and",
						m("age", m("_lt", 55)),
						m("published", m("rating", m("_gt", 4.4)))),
				},
				"name": m("_eq", "John"),
			},
			expected: map[string]any{
				"_and": []any{
					m("name", m("_eq", "John")),
					m("name", m("_ne", "Islam")),
					m("published", m("rating", m("_gt", 4.4))),
					m("published", m("rating", m("_lt", 4.8))),
					m("published", m("genre", m("_eq", "Thriller"))),
					m("age", m("_gt", 30)),
					m("age", m("_lt", 55)),
				},
				"verified": m("_eq", true),
			},
		},
	}

	mapping := &core.DocumentMapping{
		IndexesByName: map[string][]int{"name": {0}, "age": {1}, "published": {2}, "verified": {3}},
		ChildMappings: []*core.DocumentMapping{nil, nil, {
			IndexesByName: map[string][]int{"rating": {11}, "genre": {12}},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputFilter := mapper.ToFilter(request.Filter{Conditions: tt.input}, mapping)
			actualFilter := NormalizeConditions(inputFilter.Conditions)
			expectedFilter := mapper.ToFilter(request.Filter{Conditions: tt.expected}, mapping)
			AssertEqualFilterMap(t, expectedFilter.Conditions, actualFilter)
		})
	}
}

func TestMergeFilterConditions(t *testing.T) {
	m := func(op string, val any) map[string]any {
		return map[string]any{op: val}
	}

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

	mapping := &core.DocumentMapping{
		IndexesByName: map[string][]int{"name": {0}, "age": {1}, "published": {2}, "verified": {3}},
		ChildMappings: []*core.DocumentMapping{nil, nil, {
			IndexesByName: map[string][]int{"rating": {1}},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			leftFilter := mapper.ToFilter(request.Filter{Conditions: tt.left}, mapping)
			rightFilter := mapper.ToFilter(request.Filter{Conditions: tt.right}, mapping)
			actualFilter := MergeFilterConditions(leftFilter.Conditions, rightFilter.Conditions)
			expectedFilter := mapper.ToFilter(request.Filter{Conditions: tt.expected}, mapping)
			AssertEqualFilterMap(t, expectedFilter.Conditions, actualFilter)
		})
	}
}

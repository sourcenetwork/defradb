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
)

func TestNormalizeConditions(t *testing.T) {
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
		// {
		// 	name: "flatten _and, grand children of _or",
		// 	input: r("_or",
		// 		r("_and",
		// 			r("_and",
		// 				m("name", m("_eq", "Islam")),
		// 				m("age", m("_eq", "30")),
		// 			),
		// 			m("verified", m("_eq", false)),
		// 		),
		// 		r("_and",
		// 			m("name", m("_eq", "John")),
		// 			m("verified", m("_eq", true)),
		// 		),
		// 	),
		// 	expected: r("_or",
		// 		r("_and",
		// 			m("name", m("_eq", "Islam")),
		// 			m("age", m("_eq", "30")),
		// 			m("verified", m("_eq", false)),
		// 		),
		// 		r("_and",
		// 			m("name", m("_eq", "John")),
		// 			m("verified", m("_eq", true)),
		// 		),
		// 	),
		// },
		// {
		// 	name: "squash same keys into _and",
		// 	input: map[string]any{
		// 		"_and": []any{
		// 			r("_and",
		// 				m("age", m("_gt", 30)),
		// 				m("published", m("rating", m("_lt", 4.8))),
		// 			),
		// 			r("_and", m("age", m("_lt", 55))),
		// 			m("age", m("_ne", 33)),
		// 		},
		// 		"name": m("_eq", "John"),
		// 	},
		// 	expected: map[string]any{
		// 		"name":      m("_eq", "John"),
		// 		"published": m("rating", m("_lt", 4.8)),
		// 		"_and": []any{
		// 			m("age", m("_gt", 30)),
		// 			m("age", m("_lt", 55)),
		// 			m("age", m("_ne", 33)),
		// 		},
		// 	},
		// },
		// {
		// 	name: "squash same keys into _and (with more matching keys)",
		// 	input: map[string]any{
		// 		"_and": []any{
		// 			m("published", m("rating", m("_lt", 4.8))),
		// 			r("_and", m("name", m("_ne", "Islam"))),
		// 			r("_and",
		// 				m("age", m("_gt", 30)),
		// 				m("published", m("genre", m("_eq", "Thriller"))),
		// 				m("verified", m("_eq", true)),
		// 			),
		// 			r("_and",
		// 				m("age", m("_lt", 55)),
		// 				m("published", m("rating", m("_gt", 4.4)))),
		// 		},
		// 		"name": m("_eq", "John"),
		// 	},
		// 	expected: map[string]any{
		// 		"_and": []any{
		// 			m("name", m("_eq", "John")),
		// 			m("name", m("_ne", "Islam")),
		// 			m("published", m("rating", m("_gt", 4.4))),
		// 			m("published", m("rating", m("_lt", 4.8))),
		// 			m("published", m("genre", m("_eq", "Thriller"))),
		// 			m("age", m("_gt", 30)),
		// 			m("age", m("_lt", 55)),
		// 		},
		// 		"verified": m("_eq", true),
		// 	},
		// },
	}

	mapping := getDocMapping()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputFilter := mapper.ToFilter(request.Filter{Conditions: tt.input}, mapping)
			actualFilter := normalize(inputFilter.Conditions)
			expectedFilter := mapper.ToFilter(request.Filter{Conditions: tt.expected}, mapping)
			AssertEqualFilterMap(t, expectedFilter.Conditions, actualFilter)
		})
	}
}

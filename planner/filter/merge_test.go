package filter

import (
	"testing"

	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/planner/mapper"
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
			actualFilter := MergeFilterConditions(leftFilter.Conditions, rightFilter.Conditions)
			expectedFilter := mapper.ToFilter(request.Filter{Conditions: tt.expected}, mapping)
			AssertEqualFilterMap(t, expectedFilter.Conditions, actualFilter)
		})
	}
}

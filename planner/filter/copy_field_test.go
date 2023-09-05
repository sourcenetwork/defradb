package filter

import (
	"testing"

	"github.com/sourcenetwork/defradb/connor"
	"github.com/sourcenetwork/defradb/planner/mapper"
)

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

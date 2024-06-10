package connor

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type fakeFilterKey struct {
	operator string
	value    any
}

func (f fakeFilterKey) GetProp(data any) any {
	if f.value != nil {
		return f.value
	}
	return data
}

func (f fakeFilterKey) GetOperatorOrDefault(defaultOp string) string {
	if f.operator != "" {
		return f.operator
	}
	return defaultOp
}

func (f fakeFilterKey) Equal(other FilterKey) bool {
	if otherKey, isOk := other.(fakeFilterKey); isOk && f.operator == otherKey.operator && f.value == otherKey.value {
		return true
	}
	return false
}

var _ FilterKey = (*fakeFilterKey)(nil)

func TestMatchWithLogicalOperators(t *testing.T) {
	testCases := []struct {
		name       string
		conditions map[FilterKey]any
		data       any
		wantMatch  bool
		wantErr    bool
	}{
		{
			name: "_eq",
			conditions: map[FilterKey]any{
				&fakeFilterKey{"_eq", nil}: "hello",
			},
			data:      "hello",
			wantMatch: true,
		},
		{
			name: "_gt (true)",
			conditions: map[FilterKey]any{
				&fakeFilterKey{"_gt", nil}: 10,
			},
			data:      20,
			wantMatch: true,
		},
		{
			name: "_gt (false)",
			conditions: map[FilterKey]any{
				&fakeFilterKey{"_gt", nil}: 50,
			},
			data:      20,
			wantMatch: false,
		},
		{
			name: "_lt (true)",
			conditions: map[FilterKey]any{
				&fakeFilterKey{"_lt", nil}: 5,
			},
			data:      3,
			wantMatch: true,
		},
		{
			name: "_lt (false)",
			conditions: map[FilterKey]any{
				&fakeFilterKey{"_lt", nil}: 2,
			},
			data:      3,
			wantMatch: false,
		},
		{
			name: "_le (true)",
			conditions: map[FilterKey]any{
				&fakeFilterKey{"_le", nil}: 5,
			},
			data:      5,
			wantMatch: true,
		},
		{
			name: "_le (false)",
			conditions: map[FilterKey]any{
				&fakeFilterKey{"_le", nil}: 5,
			},
			data:      6,
			wantMatch: false,
		},
		{
			name: "_ge (true)",
			conditions: map[FilterKey]any{
				&fakeFilterKey{"_ge", nil}: 5,
			},
			data:      5,
			wantMatch: true,
		},
		{
			name: "_ge (false)",
			conditions: map[FilterKey]any{
				&fakeFilterKey{"_ge", nil}: 5,
			},
			data:      4,
			wantMatch: false,
		},
		{
			name: "Unsupported operator",
			conditions: map[FilterKey]any{
				&fakeFilterKey{"_xyz", nil}: "test",
			},
			data: map[string]any{
				"key": "test",
			},
			wantErr: true,
		},
		{
			name: "_in match",
			conditions: map[FilterKey]any{
				fakeFilterKey{"_in", "item1"}: []any{"item1", "item2", "item3"},
			},
			wantMatch: true,
		},
		{
			name: "_in no match",
			conditions: map[FilterKey]any{
				fakeFilterKey{"_in", "item4"}: []any{"item1", "item2", "item3"},
			},
			wantMatch: false,
		},
		{
			name: "_nin match",
			conditions: map[FilterKey]any{
				fakeFilterKey{"_nin", "item4"}: []any{"item1", "item2", "item3"},
			},
			wantMatch: true,
		},
		{
			name: "_nin no match",
			conditions: map[FilterKey]any{
				fakeFilterKey{"_nin", "item1"}: []any{"item1", "item2", "item3"},
			},
			wantMatch: false,
		},
		{
			name: "_and match",
			conditions: map[FilterKey]any{
				&fakeFilterKey{"_and", nil}: []any{
					map[FilterKey]any{&fakeFilterKey{"", "value1"}: "value1"},
					map[FilterKey]any{&fakeFilterKey{"", "value2"}: "value2"},
				},
			},
			wantMatch: true,
		},
		{
			name: "_and no match (first false)",
			conditions: map[FilterKey]any{
				&fakeFilterKey{"_and", nil}: []any{
					map[FilterKey]any{&fakeFilterKey{"", "value1"}: "wrong"},
					map[FilterKey]any{&fakeFilterKey{"", "value2"}: "value2"},
				},
			},
			wantMatch: false,
		},
		{
			name: "_and no match (second false)",
			conditions: map[FilterKey]any{
				&fakeFilterKey{"_and", nil}: []any{
					map[FilterKey]any{&fakeFilterKey{"", "value1"}: "value1"},
					map[FilterKey]any{&fakeFilterKey{"", "value2"}: "wrong"},
				},
			},
			wantMatch: false,
		},
		{
			name: "_and no match (both false)",
			conditions: map[FilterKey]any{
				&fakeFilterKey{"_and", nil}: []any{
					map[FilterKey]any{&fakeFilterKey{"", "value1"}: "wrong"},
					map[FilterKey]any{&fakeFilterKey{"", "value2"}: "wrong"},
				},
			},
			wantMatch: false,
		},
		{
			name: "_or match (both true)",
			conditions: map[FilterKey]any{
				&fakeFilterKey{"_or", nil}: []any{
					map[FilterKey]any{&fakeFilterKey{"", "value1"}: "value1"},
					map[FilterKey]any{&fakeFilterKey{"", "value2"}: "value2"},
				},
			},
			wantMatch: true,
		},
		{
			name: "_or match (first true)",
			conditions: map[FilterKey]any{
				&fakeFilterKey{"_or", nil}: []any{
					map[FilterKey]any{&fakeFilterKey{"", "value1"}: "value1"},
					map[FilterKey]any{&fakeFilterKey{"", "value2"}: "wrong"},
				},
			},
			wantMatch: true,
		},
		{
			name: "_or match (second true)",
			conditions: map[FilterKey]any{
				&fakeFilterKey{"_or", nil}: []any{
					map[FilterKey]any{&fakeFilterKey{"", "value1"}: "wrong"},
					map[FilterKey]any{&fakeFilterKey{"", "value2"}: "value2"},
				},
			},
			wantMatch: true,
		},
		{
			name: "_or no match",
			conditions: map[FilterKey]any{
				&fakeFilterKey{"_or", nil}: []any{
					map[FilterKey]any{&fakeFilterKey{"", "value1"}: "wrong"},
					map[FilterKey]any{&fakeFilterKey{"", "value2"}: "wrong"},
				},
			},
			wantMatch: false,
		},
		{
			name: "_like match",
			conditions: map[FilterKey]any{
				&fakeFilterKey{"_like", nil}: "value",
			},
			data:      "value",
			wantMatch: true,
		},
		{
			name: "_like no match",
			conditions: map[FilterKey]any{
				&fakeFilterKey{"_like", nil}: "value",
			},
			data:      "Value",
			wantMatch: false,
		},
		{
			name: "_nlike match",
			conditions: map[FilterKey]any{
				&fakeFilterKey{"_nlike", nil}: "value",
			},
			data:      "Value",
			wantMatch: true,
		},
		{
			name: "_nlike no match",
			conditions: map[FilterKey]any{
				&fakeFilterKey{"_nlike", nil}: "value",
			},
			data:      "value",
			wantMatch: false,
		},
		{
			name: "_ilike match",
			conditions: map[FilterKey]any{
				&fakeFilterKey{"_ilike", nil}: "Value",
			},
			data:      "VALUE",
			wantMatch: true,
		},
		{
			name: "_ilike no match",
			conditions: map[FilterKey]any{
				&fakeFilterKey{"_ilike", nil}: "Value",
			},
			data:      "wrong",
			wantMatch: false,
		},
		{
			name: "_not match",
			conditions: map[FilterKey]any{
				&fakeFilterKey{"_not", nil}: map[FilterKey]any{
					&fakeFilterKey{"_eq", nil}: "value",
				},
			},
			data:      "wrong",
			wantMatch: true,
		},
		{
			name: "_not no match",
			conditions: map[FilterKey]any{
				&fakeFilterKey{"_not", nil}: map[FilterKey]any{
					&fakeFilterKey{"_eq", nil}: "value",
				},
			},
			data:      "value",
			wantMatch: false,
		},
		{
			name: "nested properties",
			conditions: map[FilterKey]any{
				&fakeFilterKey{"", nil}: map[FilterKey]any{
					&fakeFilterKey{"", nil}: map[FilterKey]any{
						&fakeFilterKey{"_eq", nil}: "value",
					},
				}},
			data:      "value",
			wantMatch: true,
		},
		{
			name: "nested properties no match",
			conditions: map[FilterKey]any{
				&fakeFilterKey{"", nil}: map[FilterKey]any{
					&fakeFilterKey{"", nil}: map[FilterKey]any{
						&fakeFilterKey{"_eq", nil}: "value",
					},
				}},
			data:      "other",
			wantMatch: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			match, err := Match(tc.conditions, tc.data)
			if tc.wantErr {
				require.Error(t, err, "Test '%s' failed: Match returned an unexpected error: %v", tc.name, err)
				return
			}
			require.NoError(t, err, "Test '%s' failed: Match returned an unexpected error: %v", tc.name, err)

			if match != tc.wantMatch {
				t.Errorf("Test '%s' failed: Expected match result %v, got %v", tc.name, tc.wantMatch, match)
			}
		})
	}
}

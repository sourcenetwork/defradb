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

	"github.com/stretchr/testify/assert"

	"github.com/sourcenetwork/defradb/connor"
	"github.com/sourcenetwork/defradb/planner/mapper"
)

func TestCopyFilter(t *testing.T) {
	getFilter := func() map[connor.FilterKey]any {
		return map[connor.FilterKey]any{
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
		}
	}

	getFirstArrContent := func(f map[connor.FilterKey]any) []any {
		for _, val := range f {
			arr, isArr := val.([]any)
			if isArr {
				return arr
			}
		}
		return nil
	}

	hasGtOperator := func(f map[connor.FilterKey]any) bool {
		orContent := getFirstArrContent(f)
		for _, val := range orContent {
			andContent := getFirstArrContent(val.(map[connor.FilterKey]any))
			for _, andEl := range andContent {
				elMap := andEl.(map[connor.FilterKey]any)
				for _, v := range elMap {
					vMap := v.(map[connor.FilterKey]any)
					for k := range vMap {
						if op, ok := k.(*mapper.Operator); ok && op.Operation == "_gt" {
							return true
						}
					}
				}
			}
		}
		return false
	}

	tests := []struct {
		name string
		act  func(t *testing.T, original, copyFilter map[connor.FilterKey]any)
	}{
		{
			name: "add new value to top level",
			act: func(t *testing.T, original, copyFilter map[connor.FilterKey]any) {
				assert.Len(t, original, 1)

				original[&mapper.Operator{Operation: "_and"}] = []any{}
				assert.Len(t, original, 2)
				assert.Len(t, copyFilter, 1)
			},
		},
		{
			name: "change array value",
			act: func(t *testing.T, original, copyFilter map[connor.FilterKey]any) {
				orContent := getFirstArrContent(original)
				assert.True(t, hasGtOperator(original))

				copy(orContent[1:], orContent[2:])
				assert.False(t, hasGtOperator(original))
				assert.True(t, hasGtOperator(copyFilter))
			},
		},
		{
			name: "change nested map value",
			act: func(t *testing.T, original, copyFilter map[connor.FilterKey]any) {
				getFirstOrEl := func(f map[connor.FilterKey]any) map[connor.FilterKey]any {
					orContent := getFirstArrContent(f)
					return orContent[0].(map[connor.FilterKey]any)
				}
				elMap := getFirstOrEl(original)
				assert.Len(t, elMap, 1)

				elMap[&mapper.Operator{Operation: "_and"}] = []any{}

				assert.Len(t, getFirstOrEl(original), 2)
				assert.Len(t, getFirstOrEl(copyFilter), 1)
			},
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			f := getFilter()
			test.act(t, f, Copy(f))
		})
	}
}

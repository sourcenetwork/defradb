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
	"github.com/sourcenetwork/defradb/connor"
	"github.com/sourcenetwork/defradb/planner/mapper"

	"github.com/stretchr/testify/assert"
)

func TestUnwrapRelation(t *testing.T) {
	tests := []struct {
		name           string
		inputFilter    map[string]any
		expectedFilter map[string]any
	}{
		{
			name:           "simple",
			inputFilter:    m("published", m("rating", m("_gt", 4.0))),
			expectedFilter: m("rating", m("_gt", 4.0)),
		},
		{
			name: "no relation object",
			inputFilter: map[string]any{
				"name": m("_eq", "John"),
				"age":  m("_gt", 55),
			},
			expectedFilter: nil,
		},
		{
			name: "within _or and _and",
			inputFilter: r("_and",
				r("_or",
					r("_and",
						m("name", m("_eq", "John")),
						m("age", m("_gt", 30)),
						m("published", m("stores", m("address", m("_eq", "123 Main St")))),
						m("published", m("rating", m("_gt", 4.0))),
					),
				),
				r("_or",
					m("published", m("stores", m("address", m("_eq", "2 Ave")))),
				),
				m("published", m("genre", m("_eq", "Comedy"))),
			),
			expectedFilter: r("_and",
				r("_or",
					r("_and",
						m("stores", m("address", m("_eq", "123 Main St"))),
						m("rating", m("_gt", 4.0)),
					),
				),
				r("_or",
					m("stores", m("address", m("_eq", "2 Ave"))),
				),
				m("genre", m("_eq", "Comedy")),
			),
		},
	}

	mapping := getDocMapping()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			inputFilter := mapper.ToFilter(request.Filter{Conditions: test.inputFilter}, mapping)
			actualFilter := UnwrapRelation(inputFilter, mapper.Field{Index: authorPublishedInd})
			childMapping := mapping.ChildMappings[authorPublishedInd]
			expectedFilter := mapper.ToFilter(request.Filter{Conditions: test.expectedFilter}, childMapping)
			if expectedFilter == nil && actualFilter == nil {
				return
			}
			AssertEqualFilterMap(t, expectedFilter.Conditions, actualFilter.Conditions)
		})
	}
}

func TestUnwrapRelationOfNullFilter(t *testing.T) {
	actualFilter := CopyField(nil, mapper.Field{Index: 1})
	assert.Nil(t, actualFilter)
}

func TestUnwrapRelationWithNoFieldGiven(t *testing.T) {
	filter := mapper.NewFilter()
	filter.Conditions = map[connor.FilterKey]any{
		&mapper.PropertyIndex{Index: 0}: &mapper.Operator{Operation: "_eq"},
	}
	actualFilter := CopyField(filter)
	assert.Nil(t, actualFilter)
}

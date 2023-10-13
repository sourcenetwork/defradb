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

func TestCopyField(t *testing.T) {
	tests := []struct {
		name           string
		inputField     []mapper.Field
		inputFilter    map[string]any
		expectedFilter map[string]any
	}{
		{
			name: "flat structure",
			inputFilter: map[string]any{
				"name": m("_eq", "John"),
				"age":  m("_gt", 55),
			},
			inputField:     []mapper.Field{{Index: authorAgeInd}},
			expectedFilter: m("age", m("_gt", 55)),
		},
		{
			name: "within _and",
			inputFilter: r("_and",
				m("name", m("_eq", "John")),
				m("age", m("_gt", 55)),
			),
			inputField: []mapper.Field{{Index: authorAgeInd}},
			expectedFilter: r("_and",
				m("age", m("_gt", 55)),
			),
		},
		{
			name: "within _or and _and",
			inputFilter: r("_and",
				r("_or",
					r("_and",
						m("name", m("_eq", "John")),
						m("age", m("_gt", 30)),
					),
				),
				r("_or",
					m("name", m("_eq", "Islam")),
					m("age", m("_lt", 55)),
				),
			),
			inputField: []mapper.Field{{Index: authorAgeInd}},
			expectedFilter: r("_and",
				r("_or",
					r("_and",
						m("age", m("_gt", 30)),
					),
				),
				r("_or",
					m("age", m("_lt", 55)),
				),
			),
		},
		{
			name: "field of related object",
			inputFilter: r("_and",
				r("_or",
					r("_and",
						m("published", m("rating", m("_gt", 4.0))),
						m("age", m("_gt", 30)),
					),
				),
				m("published", m("genre", m("_eq", "Comedy"))),
				m("name", m("_eq", "John")),
			),
			inputField: []mapper.Field{{Index: authorPublishedInd}, {Index: bookRatingInd}},
			expectedFilter: r("_and",
				r("_or",
					r("_and",
						m("published", m("rating", m("_gt", 4.0))),
					),
				),
			),
		},
		{
			name: "field of related object (deeper)",
			inputFilter: r("_and",
				m("published", m("rating", m("_gt", 4.0))),
				m("age", m("_gt", 30)),
				m("published", m("stores", m("address", m("_eq", "123 Main St")))),
				m("published", m("genre", m("_eq", "Comedy"))),
				m("name", m("_eq", "John")),
			),
			inputField: []mapper.Field{{Index: authorPublishedInd}, {Index: bookStoresInd}, {Index: storeAddressInd}},
			expectedFilter: r("_and",
				m("published", m("stores", m("address", m("_eq", "123 Main St")))),
			),
		},
	}

	mapping := getDocMapping()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			inputFilter := mapper.ToFilter(request.Filter{Conditions: test.inputFilter}, mapping)
			actualFilter := CopyField(inputFilter, test.inputField...)
			expectedFilter := mapper.ToFilter(request.Filter{Conditions: test.expectedFilter}, mapping)
			AssertEqualFilterMap(t, expectedFilter.Conditions, actualFilter.Conditions)
		})
	}
}

func TestCopyFieldOfNullFilter(t *testing.T) {
	actualFilter := CopyField(nil, mapper.Field{Index: 1})
	assert.Nil(t, actualFilter)
}

func TestCopyFieldWithNoFieldGiven(t *testing.T) {
	filter := mapper.NewFilter()
	filter.Conditions = map[connor.FilterKey]any{
		&mapper.PropertyIndex{Index: 0}: &mapper.Operator{Operation: "_eq"},
	}
	actualFilter := CopyField(filter)
	assert.Nil(t, actualFilter)
}

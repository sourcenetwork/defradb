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
	"reflect"
	"testing"

	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/planner/mapper"

	"github.com/stretchr/testify/assert"
)

func TestExtractProperties(t *testing.T) {
	tests := []struct {
		name           string
		inputFilter    map[string]any
		expectedFilter map[int]Property
	}{
		{
			name: "no nesting",
			inputFilter: map[string]any{
				"name": m("_eq", "John"),
				"age":  m("_gt", 55),
			},
			expectedFilter: map[int]Property{
				authorNameInd: {Index: authorNameInd},
				authorAgeInd:  {Index: authorAgeInd},
			},
		},
		{
			name: "within _and, _or and _not",
			inputFilter: r("_or",
				m("name", m("_eq", "John")),
				r("_and",
					m("age", m("_gt", 55)),
					m("_not",
						r("_or",
							m("verified", m("_eq", true)),
						),
					),
				),
			),
			expectedFilter: map[int]Property{
				authorNameInd:     {Index: authorNameInd},
				authorAgeInd:      {Index: authorAgeInd},
				authorVerifiedInd: {Index: authorVerifiedInd},
			},
		},
		{
			name: "related field",
			inputFilter: r("_or",
				m("name", m("_eq", "John")),
				m("published", m("genre", m("_eq", "Comedy"))),
			),
			expectedFilter: map[int]Property{
				authorNameInd: {Index: authorNameInd},
				authorPublishedInd: {
					Index:  authorPublishedInd,
					Fields: map[int]Property{bookGenreInd: {Index: bookGenreInd}},
				},
			},
		},
		{
			name: "several related field with deeper nesting",
			inputFilter: r("_or",
				m("name", m("_eq", "John")),
				m("published", m("genre", m("_eq", "Comedy"))),
				m("published", m("rating", m("_gt", 55))),
				m("published", m("stores", m("name", m("_eq", "Amazon")))),
				m("published", m("stores", m("address", m("_gt", "5th Avenue")))),
			),
			expectedFilter: map[int]Property{
				authorNameInd: {Index: authorNameInd},
				authorPublishedInd: {
					Index: authorPublishedInd,
					Fields: map[int]Property{
						bookGenreInd:  {Index: bookGenreInd},
						bookRatingInd: {Index: bookRatingInd},
						bookStoresInd: {
							Index: bookStoresInd,
							Fields: map[int]Property{
								storeNameInd:    {Index: storeNameInd},
								storeAddressInd: {Index: storeAddressInd},
							},
						},
					},
				},
			},
		},
	}

	mapping := getDocMapping()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			inputFilter := mapper.ToFilter(request.Filter{Conditions: test.inputFilter}, mapping)
			actualFilter := ExtractProperties(inputFilter.Conditions)
			reflect.DeepEqual(test.expectedFilter, actualFilter)
			assert.Equal(t, test.expectedFilter, actualFilter)
		})
	}
}

func TestExtractPropertiesOfNullFilter(t *testing.T) {
	actualFilter := CopyField(nil, mapper.Field{Index: 1})
	assert.Nil(t, actualFilter)
}

// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package inline_array

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/db/tests"
)

func TestQueryInlineArrayWithBooleans(t *testing.T) {
	tests := []testUtils.QueryTestCase{
		{
			Description: "Simple inline array with no filter, nil boolean array",
			Query: `query {
						users {
							Name
							LikedIndexes
						}
					}`,
			Docs: map[int][]string{
				0: {
					(`{
					"Name": "John",
					"LikedIndexes": null
				}`),
				},
			},
			Results: []map[string]interface{}{
				{
					"Name":         "John",
					"LikedIndexes": nil,
				},
			},
		},
		{
			Description: "Simple inline array with no filter, empty boolean array",
			Query: `query {
						users {
							Name
							LikedIndexes
						}
					}`,
			Docs: map[int][]string{
				0: {
					(`{
					"Name": "John",
					"LikedIndexes": []
				}`),
				},
			},
			Results: []map[string]interface{}{
				{
					"Name":         "John",
					"LikedIndexes": []bool{},
				},
			},
		},
		{
			Description: "Simple inline array with no filter, booleans",
			Query: `query {
						users {
							Name
							LikedIndexes
						}
					}`,
			Docs: map[int][]string{
				0: {
					(`{
					"Name": "John",
					"LikedIndexes": [true, true, false, true]
				}`),
				},
			},
			Results: []map[string]interface{}{
				{
					"Name":         "John",
					"LikedIndexes": []bool{true, true, false, true},
				},
			},
		},
	}

	for _, test := range tests {
		executeTestCase(t, test)
	}
}

func TestQueryInlineArrayWithIntegers(t *testing.T) {
	tests := []testUtils.QueryTestCase{
		{
			Description: "Simple inline array with no filter, nil integer array",
			Query: `query {
						users {
							Name
							FavouriteIntegers
						}
					}`,
			Docs: map[int][]string{
				0: {
					(`{
					"Name": "John",
					"FavouriteIntegers": null
				}`),
				},
			},
			Results: []map[string]interface{}{
				{
					"Name":              "John",
					"FavouriteIntegers": nil,
				},
			},
		},
		{
			Description: "Simple inline array with no filter, empty integer array",
			Query: `query {
						users {
							Name
							FavouriteIntegers
						}
					}`,
			Docs: map[int][]string{
				0: {
					(`{
					"Name": "John",
					"FavouriteIntegers": []
				}`),
				},
			},
			Results: []map[string]interface{}{
				{
					"Name":              "John",
					"FavouriteIntegers": []int64{},
				},
			},
		},
		{
			Description: "Simple inline array with no filter, positive integers",
			Query: `query {
						users {
							Name
							FavouriteIntegers
						}
					}`,
			Docs: map[int][]string{
				0: {
					(`{
					"Name": "John",
					"FavouriteIntegers": [1, 2, 3, 5, 8]
				}`),
				},
			},
			Results: []map[string]interface{}{
				{
					"Name":              "John",
					"FavouriteIntegers": []int64{1, 2, 3, 5, 8},
				},
			},
		},
		{
			Description: "Simple inline array with no filter, negative integers",
			Query: `query {
						users {
							Name
							FavouriteIntegers
						}
					}`,
			Docs: map[int][]string{
				0: {
					(`{
					"Name": "Andy",
					"FavouriteIntegers": [-1, -2, -3, -5, -8]
				}`),
				},
			},
			Results: []map[string]interface{}{
				{
					"Name":              "Andy",
					"FavouriteIntegers": []int64{-1, -2, -3, -5, -8},
				},
			},
		},
		{
			Description: "Simple inline array with no filter, mixed integers",
			Query: `query {
						users {
							Name
							FavouriteIntegers
						}
					}`,
			Docs: map[int][]string{
				0: {
					(`{
					"Name": "Shahzad",
					"FavouriteIntegers": [-1, 2, -1, 1, 0]
				}`),
				},
			},
			Results: []map[string]interface{}{
				{
					"Name":              "Shahzad",
					"FavouriteIntegers": []int64{-1, 2, -1, 1, 0},
				},
			},
		},
	}

	for _, test := range tests {
		executeTestCase(t, test)
	}
}

func TestQueryInlineArrayWithFloats(t *testing.T) {
	tests := []testUtils.QueryTestCase{
		{
			Description: "Simple inline array with no filter, nil float array",
			Query: `query {
						users {
							Name
							FavouriteFloats
						}
					}`,
			Docs: map[int][]string{
				0: {
					(`{
					"Name": "John",
					"FavouriteFloats": null
				}`),
				},
			},
			Results: []map[string]interface{}{
				{
					"Name":            "John",
					"FavouriteFloats": nil,
				},
			},
		},
		{
			Description: "Simple inline array with no filter, empty float array",
			Query: `query {
						users {
							Name
							FavouriteFloats
						}
					}`,
			Docs: map[int][]string{
				0: {
					(`{
					"Name": "John",
					"FavouriteFloats": []
				}`),
				},
			},
			Results: []map[string]interface{}{
				{
					"Name":            "John",
					"FavouriteFloats": []float64{},
				},
			},
		},
		{
			Description: "Simple inline array with no filter, positive floats",
			Query: `query {
						users {
							Name
							FavouriteFloats
						}
					}`,
			Docs: map[int][]string{
				0: {
					(`{
					"Name": "John",
					"FavouriteFloats": [3.1425, 0.00000000001, 10]
				}`),
				},
			},
			Results: []map[string]interface{}{
				{
					"Name":            "John",
					"FavouriteFloats": []float64{3.1425, 0.00000000001, 10},
				},
			},
		},
	}

	for _, test := range tests {
		executeTestCase(t, test)
	}
}

func TestQueryInlineArrayWithStrings(t *testing.T) {
	tests := []testUtils.QueryTestCase{
		{
			Description: "Simple inline array with no filter, nil string array",
			Query: `query {
						users {
							Name
							PreferredStrings
						}
					}`,
			Docs: map[int][]string{
				0: {
					(`{
					"Name": "John",
					"PreferredStrings": null
				}`),
				},
			},
			Results: []map[string]interface{}{
				{
					"Name":             "John",
					"PreferredStrings": nil,
				},
			},
		},
		{
			Description: "Simple inline array with no filter, empty string array",
			Query: `query {
						users {
							Name
							PreferredStrings
						}
					}`,
			Docs: map[int][]string{
				0: {
					(`{
					"Name": "John",
					"PreferredStrings": []
				}`),
				},
			},
			Results: []map[string]interface{}{
				{
					"Name":             "John",
					"PreferredStrings": []string{},
				},
			},
		},
		{
			Description: "Simple inline array with no filter, strings",
			Query: `query {
						users {
							Name
							PreferredStrings
						}
					}`,
			Docs: map[int][]string{
				0: {
					(`{
					"Name": "John",
					"PreferredStrings": ["", "the previous", "the first", "empty string"]
				}`),
				},
			},
			Results: []map[string]interface{}{
				{
					"Name":             "John",
					"PreferredStrings": []string{"", "the previous", "the first", "empty string"},
				},
			},
		},
	}

	for _, test := range tests {
		executeTestCase(t, test)
	}
}

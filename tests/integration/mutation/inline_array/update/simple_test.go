// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package update

import (
	"testing"

	. "github.com/sourcenetwork/defradb/client"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	inlineArray "github.com/sourcenetwork/defradb/tests/integration/mutation/inline_array"
)

func TestMutationInlineArrayUpdateWithBooleans(t *testing.T) {
	tests := []testUtils.QueryTestCase{
		{
			Description: "Simple update mutation with boolean array, replace with nil",
			Query: `mutation {
						update_users(data: "{\"LikedIndexes\": null}") {
							Name
							LikedIndexes
						}
					}`,
			Docs: map[int][]string{
				0: {
					`{
						"Name": "John",
						"LikedIndexes": [true, true, false, true]
					}`,
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
			Description: "Simple update mutation with boolean array, replace with empty",
			Query: `mutation {
						update_users(data: "{\"LikedIndexes\": []}") {
							Name
							LikedIndexes
						}
					}`,
			Docs: map[int][]string{
				0: {
					`{
						"Name": "John",
						"LikedIndexes": [true, true, false, true]
					}`,
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
			Description: "Simple update mutation with boolean array, replace with same size",
			Query: `mutation {
						update_users(data: "{\"LikedIndexes\": [true, false, true, false]}") {
							Name
							LikedIndexes
						}
					}`,
			Docs: map[int][]string{
				0: {
					`{
						"Name": "John",
						"LikedIndexes": [true, true, false, true]
					}`,
				},
			},
			Results: []map[string]interface{}{
				{
					"Name":         "John",
					"LikedIndexes": []bool{true, false, true, false},
				},
			},
		},
		{
			Description: "Simple update mutation with boolean array, replace with smaller size",
			Query: `mutation {
						update_users(data: "{\"LikedIndexes\": [false, true]}") {
							Name
							LikedIndexes
						}
					}`,
			Docs: map[int][]string{
				0: {
					`{
						"Name": "John",
						"LikedIndexes": [true, true, false, true]
					}`,
				},
			},
			Results: []map[string]interface{}{
				{
					"Name":         "John",
					"LikedIndexes": []bool{false, true},
				},
			},
		},
		{
			Description: "Simple update mutation with boolean array, replace with larger size",
			Query: `mutation {
						update_users(data: "{\"LikedIndexes\": [true, false, true, false, true, true]}") {
							Name
							LikedIndexes
						}
					}`,
			Docs: map[int][]string{
				0: {
					`{
						"Name": "John",
						"LikedIndexes": [true, true, false, true]
					}`,
				},
			},
			Results: []map[string]interface{}{
				{
					"Name":         "John",
					"LikedIndexes": []bool{true, false, true, false, true, true},
				},
			},
		},
	}

	for _, test := range tests {
		inlineArray.ExecuteTestCase(t, test)
	}
}

func TestMutationInlineArrayWithNillableBooleans(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple inline array with no filter, booleans",
		Query: `mutation {
					update_users(data: "{\"IndexLikesDislikes\": [true, true, false, true, null]}") {
						Name
						IndexLikesDislikes
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"IndexLikesDislikes": [true, true, false, true]
				}`,
			},
		},
		Results: []map[string]interface{}{
			{
				"Name":               "John",
				"IndexLikesDislikes": []Option[bool]{Some(true), Some(true), Some(false), Some(true), None[bool]()},
			},
		},
	}

	inlineArray.ExecuteTestCase(t, test)
}

func TestMutationInlineArrayUpdateWithIntegers(t *testing.T) {
	tests := []testUtils.QueryTestCase{
		{
			Description: "Simple update mutation with integer array, replace with nil",
			Query: `mutation {
						update_users(data: "{\"FavouriteIntegers\": null}") {
							Name
							FavouriteIntegers
						}
					}`,
			Docs: map[int][]string{
				0: {
					`{
						"Name": "John",
						"FavouriteIntegers": [1, 2, 3, 5, 8]
					}`,
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
			Description: "Simple update mutation with integer array, replace with empty",
			Query: `mutation {
						update_users(data: "{\"FavouriteIntegers\": []}") {
							Name
							FavouriteIntegers
						}
					}`,
			Docs: map[int][]string{
				0: {
					`{
						"Name": "John",
						"FavouriteIntegers": [1, 2, 3, 5, 8]
					}`,
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
			Description: "Simple update mutation with integer array, replace with same size, positive values",
			Query: `mutation {
						update_users(data: "{\"FavouriteIntegers\": [8, 5, 3, 2, 1]}") {
							Name
							FavouriteIntegers
						}
					}`,
			Docs: map[int][]string{
				0: {
					`{
						"Name": "John",
						"FavouriteIntegers": [1, 2, 3, 5, 8]
					}`,
				},
			},
			Results: []map[string]interface{}{
				{
					"Name":              "John",
					"FavouriteIntegers": []int64{8, 5, 3, 2, 1},
				},
			},
		},
		{
			Description: "Simple update mutation with integer array, replace with same size, positive to mixed values",
			Query: `mutation {
						update_users(data: "{\"FavouriteIntegers\": [-1, 2, -3, 5, -8]}") {
							Name
							FavouriteIntegers
						}
					}`,
			Docs: map[int][]string{
				0: {
					`{
						"Name": "John",
						"FavouriteIntegers": [1, 2, 3, 5, 8]
					}`,
				},
			},
			Results: []map[string]interface{}{
				{
					"Name":              "John",
					"FavouriteIntegers": []int64{-1, 2, -3, 5, -8},
				},
			},
		},
		{
			Description: "Simple update mutation with integer array, replace with smaller size, positive values",
			Query: `mutation {
						update_users(data: "{\"FavouriteIntegers\": [1, 2, 3]}") {
							Name
							FavouriteIntegers
						}
					}`,
			Docs: map[int][]string{
				0: {
					`{
						"Name": "John",
						"FavouriteIntegers": [1, 2, 3, 5, 8]
					}`,
				},
			},
			Results: []map[string]interface{}{
				{
					"Name":              "John",
					"FavouriteIntegers": []int64{1, 2, 3},
				},
			},
		},
		{
			Description: "Simple update mutation with integer array, replace with larger size, positive values",
			Query: `mutation {
						update_users(data: "{\"FavouriteIntegers\": [1, 2, 3, 5, 8, 13, 21]}") {
							Name
							FavouriteIntegers
						}
					}`,
			Docs: map[int][]string{
				0: {
					`{
						"Name": "John",
						"FavouriteIntegers": [1, 2, 3, 5, 8]
					}`,
				},
			},
			Results: []map[string]interface{}{
				{
					"Name":              "John",
					"FavouriteIntegers": []int64{1, 2, 3, 5, 8, 13, 21},
				},
			},
		},
	}

	for _, test := range tests {
		inlineArray.ExecuteTestCase(t, test)
	}
}

func TestMutationInlineArrayWithNillableInts(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple inline array with no filter, nillable ints",
		Query: `mutation {
					update_users(data: "{\"TestScores\": [null, 2, 3, null, 8]}") {
						Name
						TestScores
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"TestScores": [1, null, 3]
				}`,
			},
		},
		Results: []map[string]interface{}{
			{
				"Name":       "John",
				"TestScores": []Option[int64]{None[int64](), Some[int64](2), Some[int64](3), None[int64](), Some[int64](8)},
			},
		},
	}

	inlineArray.ExecuteTestCase(t, test)
}

func TestMutationInlineArrayUpdateWithFloats(t *testing.T) {
	tests := []testUtils.QueryTestCase{
		{
			Description: "Simple update mutation with float array, replace with nil",
			Query: `mutation {
						update_users(data: "{\"FavouriteFloats\": null}") {
							Name
							FavouriteFloats
						}
					}`,
			Docs: map[int][]string{
				0: {
					`{
						"Name": "John",
						"FavouriteFloats": [3.1425, 0.00000000001, 10]
					}`,
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
			Description: "Simple update mutation with float array, replace with empty",
			Query: `mutation {
						update_users(data: "{\"FavouriteFloats\": []}") {
							Name
							FavouriteFloats
						}
					}`,
			Docs: map[int][]string{
				0: {
					`{
						"Name": "John",
						"FavouriteFloats": [3.1425, 0.00000000001, 10]
					}`,
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
			Description: "Simple update mutation with float array, replace with same size",
			Query: `mutation {
						update_users(data: "{\"FavouriteFloats\": [3.1425, -0.00000000001, 1000000]}") {
							Name
							FavouriteFloats
						}
					}`,
			Docs: map[int][]string{
				0: {
					`{
						"Name": "John",
						"FavouriteFloats": [3.1425, 0.00000000001, 10]
					}`,
				},
			},
			Results: []map[string]interface{}{
				{
					"Name":            "John",
					"FavouriteFloats": []float64{3.1425, -0.00000000001, 1000000},
				},
			},
		},
		{
			Description: "Simple update mutation with float array, replace with smaller size",
			Query: `mutation {
						update_users(data: "{\"FavouriteFloats\": [3.14]}") {
							Name
							FavouriteFloats
						}
					}`,
			Docs: map[int][]string{
				0: {
					`{
						"Name": "John",
						"FavouriteFloats": [3.1425, 0.00000000001, 10]
					}`,
				},
			},
			Results: []map[string]interface{}{
				{
					"Name":            "John",
					"FavouriteFloats": []float64{3.14},
				},
			},
		},
		{
			Description: "Simple update mutation with float array, replace with larger size",
			Query: `mutation {
						update_users(data: "{\"FavouriteFloats\": [3.1425, 0.00000000001, -10, 6.626070]}") {
							Name
							FavouriteFloats
						}
					}`,
			Docs: map[int][]string{
				0: {
					`{
						"Name": "John",
						"FavouriteFloats": [3.1425, 0.00000000001, 10]
					}`,
				},
			},
			Results: []map[string]interface{}{
				{
					"Name":            "John",
					"FavouriteFloats": []float64{3.1425, 0.00000000001, -10, 6.626070},
				},
			},
		},
	}

	for _, test := range tests {
		inlineArray.ExecuteTestCase(t, test)
	}
}

func TestMutationInlineArrayWithNillableFloats(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple inline array with no filter, nillable floats",
		Query: `mutation {
					update_users(data: "{\"PageRatings\": [3.1425, -0.00000000001, null, 10]}") {
						Name
						PageRatings
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"PageRatings": [3.1425, null, -0.00000000001, 10]
				}`,
			},
		},
		Results: []map[string]interface{}{
			{
				"Name":        "John",
				"PageRatings": []Option[float64]{Some(3.1425), Some(-0.00000000001), None[float64](), Some[float64](10)},
			},
		},
	}

	inlineArray.ExecuteTestCase(t, test)
}

func TestMutationInlineArrayUpdateWithStrings(t *testing.T) {
	tests := []testUtils.QueryTestCase{
		{
			Description: "Simple update mutation with string array, replace with nil",
			Query: `mutation {
						update_users(data: "{\"PreferredStrings\": null}") {
							Name
							PreferredStrings
						}
					}`,
			Docs: map[int][]string{
				0: {
					`{
						"Name": "John",
						"PreferredStrings": ["", "the previous", "the first", "empty string"]
					}`,
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
			Description: "Simple update mutation with string array, replace with empty",
			Query: `mutation {
						update_users(data: "{\"PreferredStrings\": []}") {
							Name
							PreferredStrings
						}
					}`,
			Docs: map[int][]string{
				0: {
					`{
						"Name": "John",
						"PreferredStrings": ["", "the previous", "the first", "empty string"]
					}`,
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
			Description: "Simple update mutation with string array, replace with same size",
			Query: `mutation {
						update_users(data: "{\"PreferredStrings\": [null, \"the previous\", \"the first\", \"null string\"]}") {
							Name
							PreferredStrings
						}
					}`,
			Docs: map[int][]string{
				0: {
					`{
						"Name": "John",
						"PreferredStrings": ["", "the previous", "the first", "empty string"]
					}`,
				},
			},
			Results: []map[string]interface{}{
				{
					"Name":             "John",
					"PreferredStrings": []string{"", "the previous", "the first", "null string"},
				},
			},
		},
		{
			Description: "Simple update mutation with string array, replace with smaller size",
			Query: `mutation {
						update_users(data: "{\"PreferredStrings\": [\"\", \"the first\"]}") {
							Name
							PreferredStrings
						}
					}`,
			Docs: map[int][]string{
				0: {
					`{
						"Name": "John",
						"PreferredStrings": ["", "the previous", "the first", "empty string"]
					}`,
				},
			},
			Results: []map[string]interface{}{
				{
					"Name":             "John",
					"PreferredStrings": []string{"", "the first"},
				},
			},
		},
		{
			Description: "Simple update mutation with string array, replace with larger size",
			Query: `mutation {
						update_users(data: "{\"PreferredStrings\": [\"\", \"the previous\", \"the first\", \"empty string\", \"blank string\", \"hitchi\"]}") {
							Name
							PreferredStrings
						}
					}`,
			Docs: map[int][]string{
				0: {
					`{
						"Name": "John",
						"PreferredStrings": ["", "the previous", "the first", "empty string"]
					}`,
				},
			},
			Results: []map[string]interface{}{
				{
					"Name": "John",
					"PreferredStrings": []string{
						"",
						"the previous",
						"the first",
						"empty string",
						"blank string",
						"hitchi",
					},
				},
			},
		},
	}

	for _, test := range tests {
		inlineArray.ExecuteTestCase(t, test)
	}
}

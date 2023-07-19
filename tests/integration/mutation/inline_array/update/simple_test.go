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

	"github.com/sourcenetwork/immutable"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	inlineArray "github.com/sourcenetwork/defradb/tests/integration/mutation/inline_array"
)

func TestMutationInlineArrayUpdateWithBooleans(t *testing.T) {
	tests := []testUtils.RequestTestCase{
		{
			Description: "Simple update mutation with boolean array, replace with nil",
			Request: `mutation {
						update_Users(data: "{\"likedIndexes\": null}") {
							name
							likedIndexes
						}
					}`,
			Docs: map[int][]string{
				0: {
					`{
						"name": "John",
						"likedIndexes": [true, true, false, true]
					}`,
				},
			},
			Results: []map[string]any{
				{
					"name":         "John",
					"likedIndexes": nil,
				},
			},
		},
		{
			Description: "Simple update mutation with boolean array, replace with empty",
			Request: `mutation {
						update_Users(data: "{\"likedIndexes\": []}") {
							name
							likedIndexes
						}
					}`,
			Docs: map[int][]string{
				0: {
					`{
						"name": "John",
						"likedIndexes": [true, true, false, true]
					}`,
				},
			},
			Results: []map[string]any{
				{
					"name":         "John",
					"likedIndexes": []bool{},
				},
			},
		},
		{
			Description: "Simple update mutation with boolean array, replace with same size",
			Request: `mutation {
						update_Users(data: "{\"likedIndexes\": [true, false, true, false]}") {
							name
							likedIndexes
						}
					}`,
			Docs: map[int][]string{
				0: {
					`{
						"name": "John",
						"likedIndexes": [true, true, false, true]
					}`,
				},
			},
			Results: []map[string]any{
				{
					"name":         "John",
					"likedIndexes": []bool{true, false, true, false},
				},
			},
		},
		{
			Description: "Simple update mutation with boolean array, replace with smaller size",
			Request: `mutation {
						update_Users(data: "{\"likedIndexes\": [false, true]}") {
							name
							likedIndexes
						}
					}`,
			Docs: map[int][]string{
				0: {
					`{
						"name": "John",
						"likedIndexes": [true, true, false, true]
					}`,
				},
			},
			Results: []map[string]any{
				{
					"name":         "John",
					"likedIndexes": []bool{false, true},
				},
			},
		},
		{
			Description: "Simple update mutation with boolean array, replace with larger size",
			Request: `mutation {
						update_Users(data: "{\"likedIndexes\": [true, false, true, false, true, true]}") {
							name
							likedIndexes
						}
					}`,
			Docs: map[int][]string{
				0: {
					`{
						"name": "John",
						"likedIndexes": [true, true, false, true]
					}`,
				},
			},
			Results: []map[string]any{
				{
					"name":         "John",
					"likedIndexes": []bool{true, false, true, false, true, true},
				},
			},
		},
	}

	for _, test := range tests {
		inlineArray.ExecuteTestCase(t, test)
	}
}

func TestMutationInlineArrayWithNillableBooleans(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple inline array with no filter, booleans",
		Request: `mutation {
					update_Users(data: "{\"indexLikesDislikes\": [true, true, false, true, null]}") {
						name
						indexLikesDislikes
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"name": "John",
					"indexLikesDislikes": [true, true, false, true]
				}`,
			},
		},
		Results: []map[string]any{
			{
				"name": "John",
				"indexLikesDislikes": []immutable.Option[bool]{
					immutable.Some(true),
					immutable.Some(true),
					immutable.Some(false),
					immutable.Some(true),
					immutable.None[bool](),
				},
			},
		},
	}

	inlineArray.ExecuteTestCase(t, test)
}

func TestMutationInlineArrayUpdateWithIntegers(t *testing.T) {
	tests := []testUtils.RequestTestCase{
		{
			Description: "Simple update mutation with integer array, replace with nil",
			Request: `mutation {
						update_Users(data: "{\"favouriteIntegers\": null}") {
							name
							favouriteIntegers
						}
					}`,
			Docs: map[int][]string{
				0: {
					`{
						"name": "John",
						"favouriteIntegers": [1, 2, 3, 5, 8]
					}`,
				},
			},
			Results: []map[string]any{
				{
					"name":              "John",
					"favouriteIntegers": nil,
				},
			},
		},
		{
			Description: "Simple update mutation with integer array, replace with empty",
			Request: `mutation {
						update_Users(data: "{\"favouriteIntegers\": []}") {
							name
							favouriteIntegers
						}
					}`,
			Docs: map[int][]string{
				0: {
					`{
						"name": "John",
						"favouriteIntegers": [1, 2, 3, 5, 8]
					}`,
				},
			},
			Results: []map[string]any{
				{
					"name":              "John",
					"favouriteIntegers": []int64{},
				},
			},
		},
		{
			Description: "Simple update mutation with integer array, replace with same size, positive values",
			Request: `mutation {
						update_Users(data: "{\"favouriteIntegers\": [8, 5, 3, 2, 1]}") {
							name
							favouriteIntegers
						}
					}`,
			Docs: map[int][]string{
				0: {
					`{
						"name": "John",
						"favouriteIntegers": [1, 2, 3, 5, 8]
					}`,
				},
			},
			Results: []map[string]any{
				{
					"name":              "John",
					"favouriteIntegers": []int64{8, 5, 3, 2, 1},
				},
			},
		},
		{
			Description: "Simple update mutation with integer array, replace with same size, positive to mixed values",
			Request: `mutation {
						update_Users(data: "{\"favouriteIntegers\": [-1, 2, -3, 5, -8]}") {
							name
							favouriteIntegers
						}
					}`,
			Docs: map[int][]string{
				0: {
					`{
						"name": "John",
						"favouriteIntegers": [1, 2, 3, 5, 8]
					}`,
				},
			},
			Results: []map[string]any{
				{
					"name":              "John",
					"favouriteIntegers": []int64{-1, 2, -3, 5, -8},
				},
			},
		},
		{
			Description: "Simple update mutation with integer array, replace with smaller size, positive values",
			Request: `mutation {
						update_Users(data: "{\"favouriteIntegers\": [1, 2, 3]}") {
							name
							favouriteIntegers
						}
					}`,
			Docs: map[int][]string{
				0: {
					`{
						"name": "John",
						"favouriteIntegers": [1, 2, 3, 5, 8]
					}`,
				},
			},
			Results: []map[string]any{
				{
					"name":              "John",
					"favouriteIntegers": []int64{1, 2, 3},
				},
			},
		},
		{
			Description: "Simple update mutation with integer array, replace with larger size, positive values",
			Request: `mutation {
						update_Users(data: "{\"favouriteIntegers\": [1, 2, 3, 5, 8, 13, 21]}") {
							name
							favouriteIntegers
						}
					}`,
			Docs: map[int][]string{
				0: {
					`{
						"name": "John",
						"favouriteIntegers": [1, 2, 3, 5, 8]
					}`,
				},
			},
			Results: []map[string]any{
				{
					"name":              "John",
					"favouriteIntegers": []int64{1, 2, 3, 5, 8, 13, 21},
				},
			},
		},
	}

	for _, test := range tests {
		inlineArray.ExecuteTestCase(t, test)
	}
}

func TestMutationInlineArrayWithNillableInts(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple inline array with no filter, nillable ints",
		Request: `mutation {
					update_Users(data: "{\"testScores\": [null, 2, 3, null, 8]}") {
						name
						testScores
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"name": "John",
					"testScores": [1, null, 3]
				}`,
			},
		},
		Results: []map[string]any{
			{
				"name": "John",
				"testScores": []immutable.Option[int64]{
					immutable.None[int64](),
					immutable.Some[int64](2),
					immutable.Some[int64](3),
					immutable.None[int64](),
					immutable.Some[int64](8),
				},
			},
		},
	}

	inlineArray.ExecuteTestCase(t, test)
}

func TestMutationInlineArrayUpdateWithFloats(t *testing.T) {
	tests := []testUtils.RequestTestCase{
		{
			Description: "Simple update mutation with float array, replace with nil",
			Request: `mutation {
						update_Users(data: "{\"favouriteFloats\": null}") {
							name
							favouriteFloats
						}
					}`,
			Docs: map[int][]string{
				0: {
					`{
						"name": "John",
						"favouriteFloats": [3.1425, 0.00000000001, 10]
					}`,
				},
			},
			Results: []map[string]any{
				{
					"name":            "John",
					"favouriteFloats": nil,
				},
			},
		},
		{
			Description: "Simple update mutation with float array, replace with empty",
			Request: `mutation {
						update_Users(data: "{\"favouriteFloats\": []}") {
							name
							favouriteFloats
						}
					}`,
			Docs: map[int][]string{
				0: {
					`{
						"name": "John",
						"favouriteFloats": [3.1425, 0.00000000001, 10]
					}`,
				},
			},
			Results: []map[string]any{
				{
					"name":            "John",
					"favouriteFloats": []float64{},
				},
			},
		},
		{
			Description: "Simple update mutation with float array, replace with same size",
			Request: `mutation {
						update_Users(data: "{\"favouriteFloats\": [3.1425, -0.00000000001, 1000000]}") {
							name
							favouriteFloats
						}
					}`,
			Docs: map[int][]string{
				0: {
					`{
						"name": "John",
						"favouriteFloats": [3.1425, 0.00000000001, 10]
					}`,
				},
			},
			Results: []map[string]any{
				{
					"name":            "John",
					"favouriteFloats": []float64{3.1425, -0.00000000001, 1000000},
				},
			},
		},
		{
			Description: "Simple update mutation with float array, replace with smaller size",
			Request: `mutation {
						update_Users(data: "{\"favouriteFloats\": [3.14]}") {
							name
							favouriteFloats
						}
					}`,
			Docs: map[int][]string{
				0: {
					`{
						"name": "John",
						"favouriteFloats": [3.1425, 0.00000000001, 10]
					}`,
				},
			},
			Results: []map[string]any{
				{
					"name":            "John",
					"favouriteFloats": []float64{3.14},
				},
			},
		},
		{
			Description: "Simple update mutation with float array, replace with larger size",
			Request: `mutation {
						update_Users(data: "{\"favouriteFloats\": [3.1425, 0.00000000001, -10, 6.626070]}") {
							name
							favouriteFloats
						}
					}`,
			Docs: map[int][]string{
				0: {
					`{
						"name": "John",
						"favouriteFloats": [3.1425, 0.00000000001, 10]
					}`,
				},
			},
			Results: []map[string]any{
				{
					"name":            "John",
					"favouriteFloats": []float64{3.1425, 0.00000000001, -10, 6.626070},
				},
			},
		},
	}

	for _, test := range tests {
		inlineArray.ExecuteTestCase(t, test)
	}
}

func TestMutationInlineArrayWithNillableFloats(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple inline array with no filter, nillable floats",
		Request: `mutation {
					update_Users(data: "{\"pageRatings\": [3.1425, -0.00000000001, null, 10]}") {
						name
						pageRatings
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"name": "John",
					"pageRatings": [3.1425, null, -0.00000000001, 10]
				}`,
			},
		},
		Results: []map[string]any{
			{
				"name": "John",
				"pageRatings": []immutable.Option[float64]{
					immutable.Some(3.1425),
					immutable.Some(-0.00000000001),
					immutable.None[float64](),
					immutable.Some[float64](10),
				},
			},
		},
	}

	inlineArray.ExecuteTestCase(t, test)
}

func TestMutationInlineArrayUpdateWithStrings(t *testing.T) {
	tests := []testUtils.RequestTestCase{
		{
			Description: "Simple update mutation with string array, replace with nil",
			Request: `mutation {
						update_Users(data: "{\"preferredStrings\": null}") {
							name
							preferredStrings
						}
					}`,
			Docs: map[int][]string{
				0: {
					`{
						"name": "John",
						"preferredStrings": ["", "the previous", "the first", "empty string"]
					}`,
				},
			},
			Results: []map[string]any{
				{
					"name":             "John",
					"preferredStrings": nil,
				},
			},
		},
		{
			Description: "Simple update mutation with string array, replace with empty",
			Request: `mutation {
						update_Users(data: "{\"preferredStrings\": []}") {
							name
							preferredStrings
						}
					}`,
			Docs: map[int][]string{
				0: {
					`{
						"name": "John",
						"preferredStrings": ["", "the previous", "the first", "empty string"]
					}`,
				},
			},
			Results: []map[string]any{
				{
					"name":             "John",
					"preferredStrings": []string{},
				},
			},
		},
		{
			Description: "Simple update mutation with string array, replace with same size",
			Request: `mutation {
						update_Users(data: "{\"preferredStrings\": [null, \"the previous\", \"the first\", \"null string\"]}") {
							name
							preferredStrings
						}
					}`,
			Docs: map[int][]string{
				0: {
					`{
						"name": "John",
						"preferredStrings": ["", "the previous", "the first", "empty string"]
					}`,
				},
			},
			Results: []map[string]any{
				{
					"name":             "John",
					"preferredStrings": []string{"", "the previous", "the first", "null string"},
				},
			},
		},
		{
			Description: "Simple update mutation with string array, replace with smaller size",
			Request: `mutation {
						update_Users(data: "{\"preferredStrings\": [\"\", \"the first\"]}") {
							name
							preferredStrings
						}
					}`,
			Docs: map[int][]string{
				0: {
					`{
						"name": "John",
						"preferredStrings": ["", "the previous", "the first", "empty string"]
					}`,
				},
			},
			Results: []map[string]any{
				{
					"name":             "John",
					"preferredStrings": []string{"", "the first"},
				},
			},
		},
		{
			Description: "Simple update mutation with string array, replace with larger size",
			Request: `mutation {
						update_Users(data: "{\"preferredStrings\": [\"\", \"the previous\", \"the first\", \"empty string\", \"blank string\", \"hitchi\"]}") {
							name
							preferredStrings
						}
					}`,
			Docs: map[int][]string{
				0: {
					`{
						"name": "John",
						"preferredStrings": ["", "the previous", "the first", "empty string"]
					}`,
				},
			},
			Results: []map[string]any{
				{
					"name": "John",
					"preferredStrings": []string{
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

func TestMutationInlineArrayWithNillableStrings(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple inline array with no filter, nillable strings",
		Request: `mutation {
					update_Users(data: "{\"pageHeaders\": [\"\", \"the previous\", null, \"empty string\", \"blank string\", \"hitchi\"]}") {
						name
						pageHeaders
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"name": "John",
					"pageHeaders": ["", "the previous", "the first", "empty string", null]
				}`,
			},
		},
		Results: []map[string]any{
			{
				"name": "John",
				"pageHeaders": []immutable.Option[string]{
					immutable.Some(""),
					immutable.Some("the previous"),
					immutable.None[string](),
					immutable.Some("empty string"),
					immutable.Some("blank string"),
					immutable.Some("hitchi"),
				},
			},
		},
	}

	inlineArray.ExecuteTestCase(t, test)
}

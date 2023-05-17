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

	"github.com/sourcenetwork/immutable"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryInlineArrayWithBooleans(t *testing.T) {
	tests := []testUtils.RequestTestCase{
		{
			Description: "Simple inline array with no filter, nil boolean array",
			Request: `query {
						Users {
							name
							likedIndexes
						}
					}`,
			Docs: map[int][]string{
				0: {
					`{
						"name": "John",
						"likedIndexes": null
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
			Description: "Simple inline array with no filter, empty boolean array",
			Request: `query {
						Users {
							name
							likedIndexes
						}
					}`,
			Docs: map[int][]string{
				0: {
					`{
						"name": "John",
						"likedIndexes": []
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
			Description: "Simple inline array with no filter, booleans",
			Request: `query {
						Users {
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
					"likedIndexes": []bool{true, true, false, true},
				},
			},
		},
	}

	for _, test := range tests {
		executeTestCase(t, test)
	}
}

func TestQueryInlineArrayWithNillableBooleans(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple inline array with no filter, booleans",
		Request: `query {
					Users {
						name
						indexLikesDislikes
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"name": "John",
					"indexLikesDislikes": [true, true, false, null]
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
					immutable.None[bool](),
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineArrayWithIntegers(t *testing.T) {
	tests := []testUtils.RequestTestCase{
		{
			Description: "Simple inline array with no filter, default integer array",
			Request: `query {
						Users {
							name
							favouriteIntegers
						}
					}`,
			Docs: map[int][]string{
				0: {
					`{
						"name": "John"
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
			Description: "Simple inline array with no filter, nil integer array",
			Request: `query {
						Users {
							name
							favouriteIntegers
						}
					}`,
			Docs: map[int][]string{
				0: {
					`{
						"name": "John",
						"favouriteIntegers": null
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
			Description: "Simple inline array with no filter, empty integer array",
			Request: `query {
						Users {
							name
							favouriteIntegers
						}
					}`,
			Docs: map[int][]string{
				0: {
					`{
						"name": "John",
						"favouriteIntegers": []
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
			Description: "Simple inline array with no filter, positive integers",
			Request: `query {
						Users {
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
					"favouriteIntegers": []int64{1, 2, 3, 5, 8},
				},
			},
		},
		{
			Description: "Simple inline array with no filter, negative integers",
			Request: `query {
						Users {
							name
							favouriteIntegers
						}
					}`,
			Docs: map[int][]string{
				0: {
					`{
						"name": "Andy",
						"favouriteIntegers": [-1, -2, -3, -5, -8]
					}`,
				},
			},
			Results: []map[string]any{
				{
					"name":              "Andy",
					"favouriteIntegers": []int64{-1, -2, -3, -5, -8},
				},
			},
		},
		{
			Description: "Simple inline array with no filter, mixed integers",
			Request: `query {
						Users {
							name
							favouriteIntegers
						}
					}`,
			Docs: map[int][]string{
				0: {
					`{
						"name": "Shahzad",
						"favouriteIntegers": [-1, 2, -1, 1, 0]
					}`,
				},
			},
			Results: []map[string]any{
				{
					"name":              "Shahzad",
					"favouriteIntegers": []int64{-1, 2, -1, 1, 0},
				},
			},
		},
	}

	for _, test := range tests {
		executeTestCase(t, test)
	}
}

func TestQueryInlineArrayWithNillableInts(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple inline array with no filter, nillable ints",
		Request: `query {
					Users {
						name
						testScores
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"name": "John",
					"testScores": [-1, null, -1, 2, 0]
				}`,
			},
		},
		Results: []map[string]any{
			{
				"name": "John",
				"testScores": []immutable.Option[int64]{
					immutable.Some[int64](-1),
					immutable.None[int64](),
					immutable.Some[int64](-1),
					immutable.Some[int64](2),
					immutable.Some[int64](0),
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineArrayWithFloats(t *testing.T) {
	tests := []testUtils.RequestTestCase{
		{
			Description: "Simple inline array with no filter, nil float array",
			Request: `query {
						Users {
							name
							favouriteFloats
						}
					}`,
			Docs: map[int][]string{
				0: {
					`{
						"name": "John",
						"favouriteFloats": null
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
			Description: "Simple inline array with no filter, empty float array",
			Request: `query {
						Users {
							name
							favouriteFloats
						}
					}`,
			Docs: map[int][]string{
				0: {
					`{
						"name": "John",
						"favouriteFloats": []
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
			Description: "Simple inline array with no filter, positive floats",
			Request: `query {
						Users {
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
					"favouriteFloats": []float64{3.1425, 0.00000000001, 10},
				},
			},
		},
	}

	for _, test := range tests {
		executeTestCase(t, test)
	}
}

func TestQueryInlineArrayWithNillableFloats(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple inline array with no filter, nillable floats",
		Request: `query {
					Users {
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
					immutable.None[float64](),
					immutable.Some(-0.00000000001),
					immutable.Some[float64](10),
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineArrayWithStrings(t *testing.T) {
	tests := []testUtils.RequestTestCase{
		{
			Description: "Simple inline array with no filter, nil string array",
			Request: `query {
						Users {
							name
							preferredStrings
						}
					}`,
			Docs: map[int][]string{
				0: {
					`{
						"name": "John",
						"preferredStrings": null
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
			Description: "Simple inline array with no filter, empty string array",
			Request: `query {
						Users {
							name
							preferredStrings
						}
					}`,
			Docs: map[int][]string{
				0: {
					`{
						"name": "John",
						"preferredStrings": []
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
			Description: "Simple inline array with no filter, strings",
			Request: `query {
						Users {
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
					"preferredStrings": []string{"", "the previous", "the first", "empty string"},
				},
			},
		},
	}

	for _, test := range tests {
		executeTestCase(t, test)
	}
}

func TestQueryInlineArrayWithNillableString(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple inline array with no filter, nillable strings",
		Request: `query {
					Users {
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
					immutable.Some("the first"),
					immutable.Some("empty string"),
					immutable.None[string](),
				},
			},
		},
	}

	executeTestCase(t, test)
}

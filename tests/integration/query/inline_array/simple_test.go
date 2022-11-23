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
					`{
						"Name": "John",
						"LikedIndexes": null
					}`,
				},
			},
			Results: []map[string]any{
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
					`{
						"Name": "John",
						"LikedIndexes": []
					}`,
				},
			},
			Results: []map[string]any{
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
					`{
						"Name": "John", 
						"LikedIndexes": [true, true, false, true]
					}`,
				},
			},
			Results: []map[string]any{
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

func TestQueryInlineArrayWithNillableBooleans(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple inline array with no filter, booleans",
		Query: `query {
					users {
						Name
						IndexLikesDislikes
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"IndexLikesDislikes": [true, true, false, null]
				}`,
			},
		},
		Results: []map[string]any{
			{
				"Name": "John",
				"IndexLikesDislikes": []immutable.Option[bool]{
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
	tests := []testUtils.QueryTestCase{
		{
			Description: "Simple inline array with no filter, default integer array",
			Query: `query {
						users {
							Name
							FavouriteIntegers
						}
					}`,
			Docs: map[int][]string{
				0: {
					`{
						"Name": "John"
					}`,
				},
			},
			Results: []map[string]any{
				{
					"Name":              "John",
					"FavouriteIntegers": nil,
				},
			},
		},
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
					`{
						"Name": "John",
						"FavouriteIntegers": null
					}`,
				},
			},
			Results: []map[string]any{
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
					`{
						"Name": "John",
						"FavouriteIntegers": []
					}`,
				},
			},
			Results: []map[string]any{
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
					`{
						"Name": "John",
						"FavouriteIntegers": [1, 2, 3, 5, 8]
					}`,
				},
			},
			Results: []map[string]any{
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
					`{
						"Name": "Andy",
						"FavouriteIntegers": [-1, -2, -3, -5, -8]
					}`,
				},
			},
			Results: []map[string]any{
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
					`{
						"Name": "Shahzad",
						"FavouriteIntegers": [-1, 2, -1, 1, 0]
					}`,
				},
			},
			Results: []map[string]any{
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

func TestQueryInlineArrayWithNillableInts(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple inline array with no filter, nillable ints",
		Query: `query {
					users {
						Name
						TestScores
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"TestScores": [-1, null, -1, 2, 0]
				}`,
			},
		},
		Results: []map[string]any{
			{
				"Name": "John",
				"TestScores": []immutable.Option[int64]{
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
					`{
						"Name": "John",
						"FavouriteFloats": null
					}`,
				},
			},
			Results: []map[string]any{
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
					`{
						"Name": "John",
						"FavouriteFloats": []
					}`,
				},
			},
			Results: []map[string]any{
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
					`{
						"Name": "John",
						"FavouriteFloats": [3.1425, 0.00000000001, 10]
					}`,
				},
			},
			Results: []map[string]any{
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

func TestQueryInlineArrayWithNillableFloats(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple inline array with no filter, nillable floats",
		Query: `query {
					users {
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
		Results: []map[string]any{
			{
				"Name": "John",
				"PageRatings": []immutable.Option[float64]{
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
					`{
						"Name": "John",
						"PreferredStrings": null
					}`,
				},
			},
			Results: []map[string]any{
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
					`{
						"Name": "John",
						"PreferredStrings": []
					}`,
				},
			},
			Results: []map[string]any{
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
					`{
						"Name": "John",
						"PreferredStrings": ["", "the previous", "the first", "empty string"]
					}`,
				},
			},
			Results: []map[string]any{
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

func TestQueryInlineArrayWithNillableString(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple inline array with no filter, nillable strings",
		Query: `query {
					users {
						Name
						PageHeaders
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"PageHeaders": ["", "the previous", "the first", "empty string", null]
				}`,
			},
		},
		Results: []map[string]any{
			{
				"Name": "John",
				"PageHeaders": []immutable.Option[string]{
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

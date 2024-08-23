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
	tests := []testUtils.TestCase{
		{
			Description: "Simple inline array with no filter, nil boolean array",
			Actions: []any{
				testUtils.CreateDoc{
					Doc: `{
						"name": "John",
						"likedIndexes": null
					}`,
				},
				testUtils.Request{
					Request: `query {
			 			Users {
			 				name
			 				likedIndexes
			 			}
			 		}`,
					Results: map[string]any{
						"Users": []map[string]any{
							{
								"name":         "John",
								"likedIndexes": nil,
							},
						},
					},
				},
			},
		},
		{
			Description: "Simple inline array with no filter, empty boolean array",
			Actions: []any{
				testUtils.CreateDoc{
					Doc: `{
						"name": "John",
						"likedIndexes": []
					}`,
				},
				testUtils.Request{
					Request: `query {
						Users {
							name
							likedIndexes
						}
					}`,
					Results: map[string]any{
						"Users": []map[string]any{
							{
								"name":         "John",
								"likedIndexes": []bool{},
							},
						},
					},
				},
			},
		},
		{
			Description: "Simple inline array with no filter, booleans",
			Actions: []any{
				testUtils.CreateDoc{
					Doc: `{
						"name": "John", 
						"likedIndexes": [true, true, false, true]
					}`,
				},
				testUtils.Request{
					Request: `query {
						Users {
							name
							likedIndexes
						}
					}`,
					Results: map[string]any{
						"Users": []map[string]any{
							{
								"name":         "John",
								"likedIndexes": []bool{true, true, false, true},
							},
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		executeTestCase(t, test)
	}
}

func TestQueryInlineArrayWithNillableBooleans(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple inline array with no filter, booleans",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"indexLikesDislikes": [true, true, false, null]
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						indexLikesDislikes
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
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
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineArrayWithIntegers(t *testing.T) {
	tests := []testUtils.TestCase{
		{
			Description: "Simple inline array with no filter, default integer array",
			Actions: []any{
				testUtils.CreateDoc{
					Doc: `{
						"name": "John"
					}`,
				},
				testUtils.Request{
					Request: `query {
						Users {
							name
							favouriteIntegers
						}
					}`,
					Results: map[string]any{
						"Users": []map[string]any{
							{
								"name":              "John",
								"favouriteIntegers": nil,
							},
						},
					},
				},
			},
		},
		{
			Description: "Simple inline array with no filter, nil integer array",
			Actions: []any{
				testUtils.CreateDoc{
					Doc: `{
						"name": "John",
						"favouriteIntegers": null
					}`,
				},
				testUtils.Request{
					Request: `query {
						Users {
							name
							favouriteIntegers
						}
					}`,
					Results: map[string]any{
						"Users": []map[string]any{
							{
								"name":              "John",
								"favouriteIntegers": nil,
							},
						},
					},
				},
			},
		},
		{
			Description: "Simple inline array with no filter, empty integer array",
			Actions: []any{
				testUtils.CreateDoc{
					Doc: `{
						"name": "John",
						"favouriteIntegers": []
					}`,
				},
				testUtils.Request{
					Request: `query {
						Users {
							name
							favouriteIntegers
						}
					}`,
					Results: map[string]any{
						"Users": []map[string]any{
							{
								"name":              "John",
								"favouriteIntegers": []int64{},
							},
						},
					},
				},
			},
		},
		{
			Description: "Simple inline array with no filter, positive integers",
			Actions: []any{
				testUtils.CreateDoc{
					Doc: `{
						"name": "John",
						"favouriteIntegers": [1, 2, 3, 5, 8]
					}`,
				},
				testUtils.Request{
					Request: `query {
						Users {
							name
							favouriteIntegers
						}
					}`,
					Results: map[string]any{
						"Users": []map[string]any{
							{
								"name":              "John",
								"favouriteIntegers": []int64{1, 2, 3, 5, 8},
							},
						},
					},
				},
			},
		},
		{
			Description: "Simple inline array with no filter, negative integers",
			Actions: []any{
				testUtils.CreateDoc{
					Doc: `{
						"name": "Andy",
						"favouriteIntegers": [-1, -2, -3, -5, -8]
					}`,
				},
				testUtils.Request{
					Request: `query {
						Users {
							name
							favouriteIntegers
						}
					}`,
					Results: map[string]any{
						"Users": []map[string]any{
							{
								"name":              "Andy",
								"favouriteIntegers": []int64{-1, -2, -3, -5, -8},
							},
						},
					},
				},
			},
		},
		{
			Description: "Simple inline array with no filter, mixed integers",
			Actions: []any{
				testUtils.CreateDoc{
					Doc: `{
						"name": "Shahzad",
						"favouriteIntegers": [-1, 2, -1, 1, 0]
					}`,
				},
				testUtils.Request{
					Request: `query {
						Users {
							name
							favouriteIntegers
						}
					}`,
					Results: map[string]any{
						"Users": []map[string]any{
							{
								"name":              "Shahzad",
								"favouriteIntegers": []int64{-1, 2, -1, 1, 0},
							},
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		executeTestCase(t, test)
	}
}

func TestQueryInlineArrayWithNillableInts(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple inline array with no filter, nillable ints",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"testScores": [-1, null, -1, 2, 0]
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						testScores
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
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
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineArrayWithFloats(t *testing.T) {
	tests := []testUtils.TestCase{
		{
			Description: "Simple inline array with no filter, nil float array",
			Actions: []any{
				testUtils.CreateDoc{
					Doc: `{
						"name": "John",
						"favouriteFloats": null
					}`,
				},
				testUtils.Request{
					Request: `query {
						Users {
							name
							favouriteFloats
						}
					}`,
					Results: map[string]any{
						"Users": []map[string]any{
							{
								"name":            "John",
								"favouriteFloats": nil,
							},
						},
					},
				},
			},
		},
		{
			Description: "Simple inline array with no filter, empty float array",
			Actions: []any{
				testUtils.CreateDoc{
					Doc: `{
						"name": "John",
						"favouriteFloats": []
					}`,
				},
				testUtils.Request{
					Request: `query {
						Users {
							name
							favouriteFloats
						}
					}`,
					Results: map[string]any{
						"Users": []map[string]any{
							{
								"name":            "John",
								"favouriteFloats": []float64{},
							},
						},
					},
				},
			},
		},
		{
			Description: "Simple inline array with no filter, positive floats",
			Actions: []any{
				testUtils.CreateDoc{
					Doc: `{
						"name": "John",
						"favouriteFloats": [3.1425, 0.00000000001, 10]
					}`,
				},
				testUtils.Request{
					Request: `query {
						Users {
							name
							favouriteFloats
						}
					}`,
					Results: map[string]any{
						"Users": []map[string]any{
							{
								"name":            "John",
								"favouriteFloats": []float64{3.1425, 0.00000000001, 10},
							},
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		executeTestCase(t, test)
	}
}

func TestQueryInlineArrayWithNillableFloats(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple inline array with no filter, nillable floats",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"pageRatings": [3.1425, null, -0.00000000001, 10]
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						pageRatings
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
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
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineArrayWithStrings(t *testing.T) {
	tests := []testUtils.TestCase{
		{
			Description: "Simple inline array with no filter, nil string array",
			Actions: []any{
				testUtils.CreateDoc{
					Doc: `{
						"name": "John",
						"preferredStrings": null
					}`,
				},
				testUtils.Request{
					Request: `query {
						Users {
							name
							preferredStrings
						}
					}`,
					Results: map[string]any{
						"Users": []map[string]any{
							{
								"name":             "John",
								"preferredStrings": nil,
							},
						},
					},
				},
			},
		},
		{
			Description: "Simple inline array with no filter, empty string array",
			Actions: []any{
				testUtils.CreateDoc{
					Doc: `{
						"name": "John",
						"preferredStrings": []
					}`,
				},
				testUtils.Request{
					Request: `query {
						Users {
							name
							preferredStrings
						}
					}`,
					Results: map[string]any{
						"Users": []map[string]any{
							{
								"name":             "John",
								"preferredStrings": []string{},
							},
						},
					},
				},
			},
		},
		{
			Description: "Simple inline array with no filter, strings",
			Actions: []any{
				testUtils.CreateDoc{
					Doc: `{
						"name": "John",
						"preferredStrings": ["", "the previous", "the first", "empty string"]
					}`,
				},
				testUtils.Request{
					Request: `query {
						Users {
							name
							preferredStrings
						}
					}`,
					Results: map[string]any{
						"Users": []map[string]any{
							{
								"name":             "John",
								"preferredStrings": []string{"", "the previous", "the first", "empty string"},
							},
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		executeTestCase(t, test)
	}
}

func TestQueryInlineArrayWithNillableString(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple inline array with no filter, nillable strings",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"pageHeaders": ["", "the previous", "the first", "empty string", null]
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						pageHeaders
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
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
				},
			},
		},
	}

	executeTestCase(t, test)
}

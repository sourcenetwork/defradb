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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryInlineArrayWithBooleans_Null(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
						"name": "John",
						"likedIndexes": null
					}`,
			},
			&action.Request{
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
	}

	executeTestCase(t, test)
}
func TestQueryInlineArrayWithBooleans_EmptyList(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
						"name": "John",
						"likedIndexes": []
					}`,
			},
			&action.Request{
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
	}

	executeTestCase(t, test)
}
func TestQueryInlineArrayWithBooleans_NotEmpty(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
						"name": "John", 
						"likedIndexes": [true, true, false, true]
					}`,
			},
			&action.Request{
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
	}

	executeTestCase(t, test)
}

func TestQueryInlineArrayWithNillableBooleans(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"indexLikesDislikes": [true, true, false, null]
				}`,
			},
			&action.Request{
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

func TestQueryInlineArrayWithIntegers_Missing(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
						"name": "John"
					}`,
			},
			&action.Request{
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
	}

	executeTestCase(t, test)
}

func TestQueryInlineArrayWithIntegers_Null(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
						"name": "John",
						"favouriteIntegers": null
					}`,
			},
			&action.Request{
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
	}

	executeTestCase(t, test)
}

func TestQueryInlineArrayWithIntegers_EmptyList(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
						"name": "John",
						"favouriteIntegers": []
					}`,
			},
			&action.Request{
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
	}

	executeTestCase(t, test)
}

func TestQueryInlineArrayWithIntegers_NotEmptyList(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
						"name": "John",
						"favouriteIntegers": [1, 2, 3, 5, 8]
					}`,
			},
			&action.Request{
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
	}

	executeTestCase(t, test)
}

func TestQueryInlineArrayWithNegativeIntegers_NotEmptyList(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
						"name": "Andy",
						"favouriteIntegers": [-1, -2, -3, -5, -8]
					}`,
			},
			&action.Request{
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
	}

	executeTestCase(t, test)
}

func TestQueryInlineArrayWithMixIntegers_NotEmptyList(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
						"name": "Shahzad",
						"favouriteIntegers": [-1, 2, -1, 1, 0]
					}`,
			},
			&action.Request{
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
	}

	executeTestCase(t, test)
}
func TestQueryInlineArrayWithNillableInts(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"testScores": [-1, null, -1, 2, 0]
				}`,
			},
			&action.Request{
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

func TestQueryInlineArrayWithFloats_Null(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
						"name": "John",
						"favouriteFloats": null
					}`,
			},
			&action.Request{
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
	}

	executeTestCase(t, test)
}

func TestQueryInlineArrayWithFloats_EmptyList(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
						"name": "John",
						"favouriteFloats": []
					}`,
			},
			&action.Request{
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
	}

	executeTestCase(t, test)
}

func TestQueryInlineArrayWithFloats_NotEmpty(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
						"name": "John",
						"favouriteFloats": [3.1425, 0.00000000001, 10]
					}`,
			},
			&action.Request{
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
	}

	executeTestCase(t, test)
}

func TestQueryInlineArrayWithNillableFloats(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"pageRatings": [3.1425, null, -0.00000000001, 10]
				}`,
			},
			&action.Request{
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

func TestQueryInlineArrayWithStrings_Null(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
						"name": "John",
						"preferredStrings": null
					}`,
			},
			&action.Request{
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
	}

	executeTestCase(t, test)
}

func TestQueryInlineArrayWithStrings_EmptyList(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
						"name": "John",
						"preferredStrings": []
					}`,
			},
			&action.Request{
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
	}

	executeTestCase(t, test)
}

func TestQueryInlineArrayWithStrings_NotEmpty(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
						"name": "John",
						"preferredStrings": ["", "the previous", "the first", "empty string"]
					}`,
			},
			&action.Request{
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
	}

	executeTestCase(t, test)
}

func TestQueryInlineArrayWithNillableString(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"pageHeaders": ["", "the previous", "the first", "empty string", null]
				}`,
			},
			&action.Request{
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

// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package simple

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQuerySimple_WithSimilarityOnQuery_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `type User {
					name: String
					vector: [Int!]
				}`,
			},
			testUtils.Request{
				Request: `query {
					_similarity
				}`,
				ExpectedError: "Cannot query field \"_similarity\" on type \"Query\".",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQuerySimple_WithSimilarityOnUndefinedField_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `type User {
					name: String
				}`,
			},
			testUtils.Request{
				Request: `query {
					User{
						_similarity(pointsList: {vector: [1, 2, 3]})
					}
				}`,
				ExpectedError: "Unknown argument \"pointsList\" on field \"_similarity\" of type \"User\".",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQuerySimple_WithSimilarityAndWrongVectorValueType_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `type User {
					name: String
					pointsList: [Int!]
				}`,
			},
			testUtils.Request{
				Request: `query {
					User{
						_similarity(pointsList: {vector: [1.1, 1.2, 0.9]})
					}
				}`,
				ExpectedError: "Argument \"pointsList\" has invalid value {vector: [1.1, 1.2, 0.9]}.\nIn field " +
					"\"vector\": In element #1: Expected type \"Int\", found 1.1.\nIn field \"vector\": In element #1: " +
					"Expected type \"Int\", found 1.2.\nIn field \"vector\": In element #1: Expected type \"Int\", found 0.9.",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQuerySimple_WithSimilarityAndWrongFieldType_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `type User {
					name: String
					pets: [String!]
				}`,
			},
			testUtils.Request{
				Request: `query {
					User{
						_similarity(pets: {vector: [1.1, 1.2, 0.9]})
					}
				}`,
				// Not found on _similarity because it's not a supported type.
				ExpectedError: "Unknown argument \"pets\" on field \"_similarity\" of type \"User\".",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQuerySimple_WithSimilarityOnEmptyCollection_ShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `type User {
					name: String
					pointsList: [Int!]
				}`,
			},
			testUtils.Request{
				Request: `query {
					User{
						_similarity(pointsList: {vector: [1, 2, 3]})
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQuerySimple_WithIntSimilarity_ShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `type User {
					name: String
					pointsList: [Int!]
				}`,
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name":       "John",
					"pointsList": []int64{2, 4, 1},
				},
			},
			testUtils.Request{
				Request: `query {
					User{
						name
						_similarity(pointsList: {vector: [1, 2, 0]})
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name":        "John",
							"_similarity": float64(10),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQuerySimple_WithIntSimilarityDifferentVectorLength_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `type User {
					name: String
					pointsList: [Int!]
				}`,
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name":       "John",
					"pointsList": []int64{2, 4, 1},
				},
			},
			testUtils.Request{
				Request: `query {
					User{
						name
						_similarity(pointsList: {vector: [1, 2, 0, 1]})
					}
				}`,
				ExpectedError: "source and vector must be of the same length. Source: 3, Vector: 4",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQuerySimple_WithFloat32Similarity_ShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `type User {
					name: String
					pointsList: [Float32!]
				}`,
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name":       "John",
					"pointsList": []float32{2, 4, 1},
				},
			},
			testUtils.Request{
				Request: `query {
					User{
						name
						_similarity(pointsList: {vector: [1, 2, 0]})
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name":        "John",
							"_similarity": float64(10),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQuerySimple_WithFloat64Similarity_ShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `type User {
					name: String
					pointsList: [Float64!]
				}`,
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name":       "John",
					"pointsList": []float64{2, 4, 1},
				},
			},
			testUtils.Request{
				Request: `query {
					User{
						name
						_similarity(pointsList: {vector: [1, 2, 0]})
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name":        "John",
							"_similarity": float64(10),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQuerySimple_WithJSONDocCreationSimilarity_ShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `type User {
					name: String
					pointsList: [Float64!]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"pointsList": [2, 4, 1]
				}`,
			},
			testUtils.Request{
				Request: `query {
					User{
						name
						_similarity(pointsList: {vector: [1, 2, 0]})
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name":        "John",
							"_similarity": float64(10),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQuerySimple_WithSimilarityAndFilteringOnSimilarityResult_ShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `type User {
					name: String
					pointsList: [Int!]
				}`,
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name":       "John",
					"pointsList": []int64{2, 4, 1},
				},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name":       "Bob",
					"pointsList": []int64{1, 1, 1},
				},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name":       "Alice",
					"pointsList": []int64{4, 5, 3},
				},
			},
			testUtils.Request{
				Request: `query {
					User(filter: {_alias: {sim: {_lt: 11}}}){
						name
						sim: _similarity(pointsList: {vector: [1, 2, 0]})
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Bob",
							"sim":  float64(3),
						},
						{
							"name": "John",
							"sim":  float64(10),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQuerySimple_WithSimilarityAndOrderingWithLimitOnSimilarityResult_ShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `type User {
					name: String
					pointsList: [Int!]
				}`,
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name":       "John",
					"pointsList": []int64{2, 4, 1},
				},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name":       "Bob",
					"pointsList": []int64{1, 1, 1},
				},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name":       "Alice",
					"pointsList": []int64{4, 5, 3},
				},
			},
			testUtils.Request{
				Request: `query {
					User(order: {_alias: {sim: DESC}}, limit: 2){
						name
						sim: _similarity(pointsList: {vector: [1, 2, 0]})
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Alice",
							"sim":  float64(14),
						},
						{
							"name": "John",
							"sim":  float64(10),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQuerySimple_WithTwoSimilarityAndFilteringOnSecond_ShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `type User {
					name: String
					pointsList: [Int!]
				}`,
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name":       "John",
					"pointsList": []int64{2, 4, 1},
				},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name":       "Bob",
					"pointsList": []int64{1, 1, 1},
				},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name":       "Alice",
					"pointsList": []int64{4, 5, 3},
				},
			},
			testUtils.Request{
				Request: `query {
					User(filter: {_alias: {sim2: {_gt: 20}}}){
						name
						sim: _similarity(pointsList: {vector: [1, 2, 0]})
						sim2: _similarity(pointsList: {vector: [2, 3, 0]})
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Alice",
							"sim":  float64(14),
							"sim2": float64(23),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// This test documents a bug where having two aliases in a logical _or operator
// return no results even though in the tests bellow 1 should be returned.
// https://github.com/sourcenetwork/defradb/issues/3468
func TestQuerySimple_WithTwoSimilarityAndFilteringOnBoth_ShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `type User {
					name: String
					pointsList: [Int!]
				}`,
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name":       "John",
					"pointsList": []int64{2, 4, 1},
				},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name":       "Bob",
					"pointsList": []int64{1, 1, 1},
				},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name":       "Alice",
					"pointsList": []int64{4, 5, 3},
				},
			},
			testUtils.Request{
				Request: `query {
					User(filter: {_or: [{_alias: {sim2: {_gt: 20}}}, {_alias: {sim: {_lt: 10}}}]}){
						name
						sim: _similarity(pointsList: {vector: [1, 2, 0]})
						sim2: _similarity(pointsList: {vector: [2, 3, 0]})
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

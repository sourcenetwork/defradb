// Copyright 2023 Democratized Data Foundation
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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
	"github.com/sourcenetwork/immutable"
)

func TestMutationUpdate_WithBooleanFilter_ResultNotFilteredOut(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						verified: Boolean
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"verified": true
				}`,
			},
			&action.Request{
				// The update will result in a record that no longer matches the filter
				Request: `mutation {
					update_Users(filter: {verified: {_eq: true}}, input: {verified: false}) {
						name
						verified
					}
				}`,
				Results: map[string]any{
					"update_Users": []map[string]any{
						{
							"name":     "John",
							"verified": false,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationUpdate_WithBooleanFilter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						verified: Boolean
						points: Float
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"verified": true,
					"points": 42.1
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Bob",
					"verified": false,
					"points": 66.6
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Fred",
					"verified": true,
					"points": 33
				}`,
			},
			&action.Request{
				Request: `mutation {
					update_Users(filter: {verified: {_eq: true}}, input: {points: 59}) {
						name
						points
					}
				}`,
				Results: map[string]any{
					"update_Users": []map[string]any{
						{
							"name":   "John",
							"points": float64(59),
						},
						{
							"name":   "Fred",
							"points": float64(59),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// TestMutationUpdate_WithFilterOnUpdatedField_ReturnsResult is a regression test for
// https://github.com/sourcenetwork/defradb/issues/4279
// When an update modifies the field used in the filter, the result should still be returned.
func TestMutationUpdate_WithFilterOnUpdatedField_ReturnsResult(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John"
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Fred"
				}`,
			},
			&action.Request{
				Request: `mutation {
					update_Users(filter: {name: {_eq: "John"}}, input: {name: "Jane"}) {
						name
					}
				}`,
				Results: map[string]any{
					"update_Users": []map[string]any{
						{
							"name": "Jane",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationUpdate_WithRelationSelectInResponse_ReturnsRelation(t *testing.T) {
	test := testUtils.TestCase{
		SupportedMutationTypes: immutable.Some([]state.MutationType{
			state.GQLRequestMutationType,
		}),
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Book {
						name: String
						author: Author
					}

					type Author {
						name: String
						published: Book @primary
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Painted House"
				}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":         "John Grisham",
					"_publishedID": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.Request{
				Request: `mutation($docID: [ID!]) {
					update_Author(docID: $docID, input: {name: "Jane Grisham"}) {
						name
						published {
							name
						}
					}
				}`,
				Variables: immutable.Some(map[string]any{
					"docID": testUtils.NewDocIndex(1, 0),
				}),
				Results: map[string]any{
					"update_Author": []map[string]any{
						{
							"name": "Jane Grisham",
							"published": map[string]any{
								"name": "Painted House",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationUpdate_WithRelationFilter_CorrectlyFilters(t *testing.T) {
	test := testUtils.TestCase{
		SupportedMutationTypes: immutable.Some([]state.MutationType{
			state.GQLRequestMutationType,
		}),
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Book {
						name: String
						rating: Float
						author: Author
					}

					type Author {
						name: String
						published: Book @primary
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Painted House",
					"rating": 4.9
				}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Boring Book",
					"rating": 2.0
				}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":         "John Grisham",
					"_publishedID": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":         "Bad Writer",
					"_publishedID": testUtils.NewDocIndex(0, 1),
				},
			},
			&action.Request{
				Request: `mutation {
					update_Author(
						filter: {published: {rating: {_gt: 3}}},
						input: {name: "Jane Grisham"}
					) {
						name
					}
				}`,
				Results: map[string]any{
					"update_Author": []map[string]any{
						{
							"name": "Jane Grisham",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationUpdate_WithRelationFilterAndRelationSelect_ReturnsBoth(t *testing.T) {
	test := testUtils.TestCase{
		SupportedMutationTypes: immutable.Some([]state.MutationType{
			state.GQLRequestMutationType,
		}),
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Book {
						name: String
						rating: Float
						author: Author
					}

					type Author {
						name: String
						published: Book @primary
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Painted House",
					"rating": 4.9
				}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Boring Book",
					"rating": 2.0
				}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":         "John Grisham",
					"_publishedID": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":         "Bad Writer",
					"_publishedID": testUtils.NewDocIndex(0, 1),
				},
			},
			&action.Request{
				Request: `mutation {
					update_Author(
						filter: {published: {rating: {_gt: 3}}},
						input: {name: "Jane Grisham"}
					) {
						name
						published {
							name
						}
					}
				}`,
				Results: map[string]any{
					"update_Author": []map[string]any{
						{
							"name": "Jane Grisham",
							"published": map[string]any{
								"name": "Painted House",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

package test_explain_simple

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	explainUtils "github.com/sourcenetwork/defradb/tests/integration/explain"
)

func TestSimpleExplainWithIndexOnFilter(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
						age: Int @index
					}
				`,
			},

			&action.ExplainRequest{
				Request: `query @explain(type: simple) {
					User(filter: {age: {_eq: 30}}) {
						name
					}
				}`,

				ExpectedPatterns: dataMap{
					"explain": dataMap{
						"operationNode": []dataMap{
							{
								"selectTopNode": dataMap{
									"selectNode": dataMap{
										"scanNode": dataMap{},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestSimpleExplainWithIndexOnOrder(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
						age: Int @index
					}
				`,
			},

			&action.ExplainRequest{
				Request: `query @explain(type: simple) {
					User(order: {age: ASC}) {
						name
						age
					}
				}`,

				ExpectedPatterns: dataMap{
					"explain": dataMap{
						"operationNode": []dataMap{
							{
								"selectTopNode": dataMap{
									"selectNode": dataMap{
										"scanNode": dataMap{},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestSimpleExplainWithIndexOnSubqueryNestedRelationOrder(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Author {
						name: String
						published: [Book]
					}
					type Book {
						title: String
						author: Author
						publisher: Publisher
					}
					type Publisher {
						name: String
						establishedYear: Int @index
						book: Book @primary
					}
				`,
			},

			&action.ExplainRequest{
				Request: `query @explain(type: simple) {
					Author {
						name
						published(order: {publisher: {establishedYear: DESC}}, limit: 2) {
							title
						}
					}
				}`,

				// With index, no orderNode needed - join is inverted to use index ordering
				ExpectedPatterns: dataMap{
					"explain": dataMap{
						"operationNode": []dataMap{
							{
								"selectTopNode": dataMap{
									"selectNode": dataMap{
										"typeIndexJoin": dataMap{
											"root": dataMap{
												"scanNode": dataMap{},
											},
											"subType": dataMap{
												"selectTopNode": dataMap{
													"limitNode": dataMap{
														// No orderNode - index provides ordering
														"selectNode": dataMap{
															"typeIndexJoin": dataMap{
																"root": dataMap{
																	"scanNode": dataMap{},
																},
																"subType": dataMap{
																	"selectTopNode": dataMap{
																		"selectNode": dataMap{
																			"scanNode": dataMap{},
																		},
																	},
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

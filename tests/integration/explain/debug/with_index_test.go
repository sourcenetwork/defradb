// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_explain_debug

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	explainUtils "github.com/sourcenetwork/defradb/tests/integration/explain"
)

func TestDebugExplainWithIndexOnFilter(t *testing.T) {
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
				Request: `query @explain(type: debug) {
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

func TestDebugExplainWithIndexOnOrder(t *testing.T) {
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
				Request: `query @explain(type: debug) {
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

// Tests debug explain plan structure for subquery ordering by nested relation field with index.
// The index on Publisher.establishedYear causes the Book->Publisher join to be inverted,
// so no orderNode is needed in the subquery.
func TestDebugExplainWithIndexOnSubqueryNestedRelationOrder(t *testing.T) {
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
				Request: `query @explain(type: debug) {
					Author {
						name
						published(order: {publisher: {establishedYear: DESC}}, limit: 2) {
							title
						}
					}
				}`,

				ExpectedPatterns: dataMap{
					"explain": dataMap{
						"operationNode": []dataMap{
							{
								"selectTopNode": dataMap{
									"selectNode": dataMap{
										"typeIndexJoin": dataMap{
											"typeJoinMany": dataMap{
												"root": dataMap{
													"scanNode": dataMap{},
												},
												"subType": dataMap{
													"selectTopNode": dataMap{
														"limitNode": dataMap{
															// No orderNode here - index provides ordering
															"selectNode": dataMap{
																"typeIndexJoin": dataMap{
																	"typeJoinOne": dataMap{
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
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

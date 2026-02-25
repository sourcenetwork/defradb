// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_explain_execute

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	explainUtils "github.com/sourcenetwork/defradb/tests/integration/explain"
)

func TestExecuteExplainWithIndexOnFilter(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String
						age: Int @index
					}
				`,
			},

			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"name": "John", "age": 30}`,
			},

			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"name": "Alice", "age": 25}`,
			},

			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"name": "Bob", "age": 30}`,
			},

			&action.ExplainRequest{
				Request: `query @explain(type: execute) {
					User(filter: {age: {_eq: 30}}) {
						name
					}
				}`,

				ExpectedFullGraph: dataMap{
					"explain": dataMap{
						"executionSuccess": true,
						"sizeOfResult":     1,
						"planExecutions":   uint64(2),
						"operationNode": []dataMap{
							{
								"selectTopNode": dataMap{
									"selectNode": dataMap{
										"iterations":    uint64(3),
										"filterMatches": uint64(2),
										"scanNode": dataMap{
											"iterations":   uint64(3),
											"docFetches":   uint64(2),
											"fieldFetches": uint64(4),
											// Index is used for filtering
											"indexFetches": uint64(2),
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

func TestExecuteExplainWithIndexOnOrder(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String
						age: Int @index
					}
				`,
			},

			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"name": "John", "age": 30}`,
			},

			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"name": "Alice", "age": 25}`,
			},

			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"name": "Bob", "age": 35}`,
			},

			&action.ExplainRequest{
				Request: `query @explain(type: execute) {
					User(order: {age: ASC}) {
						name
						age
					}
				}`,

				ExpectedFullGraph: dataMap{
					"explain": dataMap{
						"executionSuccess": true,
						"sizeOfResult":     1,
						"planExecutions":   uint64(2),
						"operationNode": []dataMap{
							{
								"selectTopNode": dataMap{
									"selectNode": dataMap{
										"iterations":    uint64(4),
										"filterMatches": uint64(3),
										"scanNode": dataMap{
											"iterations":   uint64(4),
											"docFetches":   uint64(3),
											"fieldFetches": uint64(6),
											// Index is used for ordering
											"indexFetches": uint64(3),
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

func TestExecuteExplainWithIndexOnSubqueryNestedRelationOrder(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			&action.AddSchema{
				Schema: `
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

			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"name": "John"}`,
			},

			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"title":  "Book2020",
					"author": testUtils.NewDocIndex(0, 0),
				},
			},

			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"title":  "Book2010",
					"author": testUtils.NewDocIndex(0, 0),
				},
			},

			&action.AddDoc{
				CollectionID: 2,
				DocMap: map[string]any{
					"name":            "Publisher2020",
					"establishedYear": 2020,
					"book":            testUtils.NewDocIndex(1, 0),
				},
			},

			&action.AddDoc{
				CollectionID: 2,
				DocMap: map[string]any{
					"name":            "Publisher2010",
					"establishedYear": 2010,
					"book":            testUtils.NewDocIndex(1, 1),
				},
			},

			&action.ExplainRequest{
				Request: `query @explain(type: execute) {
					Author {
						name
						published(order: {publisher: {establishedYear: DESC}}, limit: 2) {
							title
						}
					}
				}`,

				ExpectedFullGraph: dataMap{
					"explain": dataMap{
						"executionSuccess": true,
						"sizeOfResult":     1,
						"planExecutions":   uint64(2),
						"operationNode": []dataMap{
							{
								"selectTopNode": dataMap{
									"selectNode": dataMap{
										"iterations":    uint64(2),
										"filterMatches": uint64(1),
										"typeIndexJoin": dataMap{
											"iterations": uint64(2),
											"typeJoinMany": dataMap{
												"root": dataMap{
													"scanNode": dataMap{
														"iterations":   uint64(2),
														"docFetches":   uint64(1),
														"fieldFetches": uint64(1),
														"indexFetches": uint64(0),
													},
												},
												"subType": dataMap{
													"selectTopNode": dataMap{
														"limitNode": dataMap{
															"iterations": uint64(3),
															"selectNode": dataMap{
																"iterations":    uint64(2),
																"filterMatches": uint64(2),
																"typeIndexJoin": dataMap{
																	"iterations": uint64(2),
																	"typeJoinOne": dataMap{
																		"root": dataMap{
																			"scanNode": dataMap{
																				"iterations":   uint64(2),
																				"docFetches":   uint64(2),
																				"fieldFetches": uint64(4),
																				"indexFetches": uint64(0),
																			},
																		},
																		"subType": dataMap{
																			"selectTopNode": dataMap{
																				"selectNode": dataMap{
																					"iterations":    uint64(2),
																					"filterMatches": uint64(2),
																					"scanNode": dataMap{
																						"iterations":   uint64(2),
																						"docFetches":   uint64(2),
																						"fieldFetches": uint64(6),
																						"indexFetches": uint64(2),
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
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

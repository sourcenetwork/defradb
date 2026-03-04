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

package test_explain_execute

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	explainUtils "github.com/sourcenetwork/defradb/tests/integration/explain"
)

func TestExecuteExplainRequestWithOrderFieldOnParent(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			add2AddressDocuments(),
			add2AuthorContactDocuments(),
			add2AuthorDocuments(),

			&action.ExplainRequest{
				Request: `query @explain(type: execute) {
					Author(order: {age: ASC}) {
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
									"orderNode": dataMap{
										"iterations": uint64(3),
										"selectNode": dataMap{
											"filterMatches": uint64(2),
											"iterations":    uint64(3),
											"scanNode": dataMap{
												"iterations":   uint64(3),
												"docFetches":   uint64(2),
												"fieldFetches": uint64(8),
												"indexFetches": uint64(0),
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

func TestExecuteExplainRequestWithMultiOrderFieldsOnParent(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			// Authors
			&action.AddDoc{
				CollectionID: 2,

				Doc: `{
					"name": "Andy",
					"age": 64
				}`,
			},

			&action.AddDoc{
				CollectionID: 2,

				Doc: `{
					"name": "Another64YearOld",
					"age": 64
				}`,
			},

			&action.AddDoc{
				CollectionID: 2,

				Doc: `{
					"name": "Shahzad",
					"age": 48
				}`,
			},

			&action.AddDoc{
				CollectionID: 2,

				Doc: `{
					"name": "Another48YearOld",
					"age": 48
				}`,
			},

			&action.ExplainRequest{
				Request: `query @explain(type: execute) {
					Author(order: [{age: ASC}, {name: DESC}]) {
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
									"orderNode": dataMap{
										"iterations": uint64(5),
										"selectNode": dataMap{
											"filterMatches": uint64(4),
											"iterations":    uint64(5),
											"scanNode": dataMap{
												"iterations":   uint64(5),
												"docFetches":   uint64(4),
												"fieldFetches": uint64(8),
												"indexFetches": uint64(0),
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

func TestExecuteExplainRequestWithOrderFieldOnChild(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			add2AddressDocuments(),
			add2AuthorContactDocuments(),
			add2AuthorDocuments(),
			add3ArticleDocuments(),

			&action.ExplainRequest{
				Request: `query @explain(type: execute) {
					Author {
						name
						articles(order: {pages: DESC}) {
							pages
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
										"iterations":    uint64(3),
										"filterMatches": uint64(2),
										"typeIndexJoin": dataMap{
											"iterations": uint64(3),
											"typeJoinMany": dataMap{
												"root": dataMap{
													"scanNode": dataMap{
														"iterations":   uint64(3),
														"docFetches":   uint64(2),
														"fieldFetches": uint64(8),
														"indexFetches": uint64(0),
													},
												},
												"subType": dataMap{
													"selectTopNode": dataMap{
														"orderNode": dataMap{
															"iterations": uint64(5),
															"selectNode": dataMap{
																"iterations":    uint64(5),
																"filterMatches": uint64(3),
																"scanNode": dataMap{
																	"iterations":   uint64(5),
																	"docFetches":   uint64(6),
																	"fieldFetches": uint64(18),
																	"indexFetches": uint64(0),
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

func TestExecuteExplainRequestWithOrderFieldOnBothParentAndChild(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			add2AddressDocuments(),
			add2AuthorContactDocuments(),
			add2AuthorDocuments(),
			add3ArticleDocuments(),

			&action.ExplainRequest{
				Request: `query @explain(type: execute) {
					Author(order: {age: ASC}) {
						name
						age
						articles(order: {pages: DESC}) {
							pages
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
									"orderNode": dataMap{
										"iterations": uint64(3),
										"selectNode": dataMap{
											"iterations":    uint64(3),
											"filterMatches": uint64(2),
											"typeIndexJoin": dataMap{
												"iterations": uint64(3),
												"typeJoinMany": dataMap{
													"root": dataMap{
														"scanNode": dataMap{
															"iterations":   uint64(3),
															"docFetches":   uint64(2),
															"fieldFetches": uint64(8),
															"indexFetches": uint64(0),
														},
													},
													"subType": dataMap{
														"selectTopNode": dataMap{
															"orderNode": dataMap{
																"iterations": uint64(5),
																"selectNode": dataMap{
																	"iterations":    uint64(5),
																	"filterMatches": uint64(3),
																	"scanNode": dataMap{
																		"iterations":   uint64(5),
																		"docFetches":   uint64(6),
																		"fieldFetches": uint64(18),
																		"indexFetches": uint64(0),
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

func TestExecuteExplainRequestWhereParentFieldIsOrderedByChildField(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			add2AddressDocuments(),
			add2AuthorContactDocuments(),
			add2AuthorDocuments(),
			add3ArticleDocuments(),

			&action.ExplainRequest{
				Request: `query @explain(type: execute) {
					Author(
						order: {
							articles: {pages: ASC}
						}
					) {
						name
						articles {
						    pages
						}
					}
				}`,

				ExpectedError: "Argument \"order\" has invalid value {articles: {pages: ASC}}.\nIn field \"articles\": Unknown field.",
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestExecuteExplainRequestWithSubqueryOrderByNestedRelationField(t *testing.T) {
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
						establishedYear: Int
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
												// Author scanNode
												"root": dataMap{
													"scanNode": dataMap{
														"iterations":   uint64(2),
														"docFetches":   uint64(1),
														"fieldFetches": uint64(1),
														"indexFetches": uint64(0),
													},
												},
												// Nested Book -> Publisher join with limit/order
												"subType": dataMap{
													"selectTopNode": dataMap{
														"limitNode": dataMap{
															"iterations": uint64(3),
															"orderNode": dataMap{
																"iterations": uint64(2),
																"selectNode": dataMap{
																	"iterations":    uint64(3),
																	"filterMatches": uint64(2),
																	"typeIndexJoin": dataMap{
																		"iterations": uint64(3),
																		"typeJoinOne": dataMap{
																			"root": dataMap{
																				"scanNode": dataMap{
																					"iterations":   uint64(3),
																					"docFetches":   uint64(2),
																					"fieldFetches": uint64(4),
																					"indexFetches": uint64(0),
																				},
																			},
																			// Publisher uses relation index (indexFetches: 2)
																			"subType": dataMap{
																				"selectTopNode": dataMap{
																					"selectNode": dataMap{
																						"iterations":    uint64(4),
																						"filterMatches": uint64(2),
																						"scanNode": dataMap{
																							"iterations":   uint64(4),
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
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestExecuteExplainRequestWithSubqueryOrderByNestedRelationFieldASC(t *testing.T) {
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
						establishedYear: Int
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
						published(order: {publisher: {establishedYear: ASC}}, limit: 2) {
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
												// Author scanNode
												"root": dataMap{
													"scanNode": dataMap{
														"iterations":   uint64(2),
														"docFetches":   uint64(1),
														"fieldFetches": uint64(1),
														"indexFetches": uint64(0),
													},
												},
												// Nested Book -> Publisher join with limit/order
												"subType": dataMap{
													"selectTopNode": dataMap{
														"limitNode": dataMap{
															"iterations": uint64(3),
															"orderNode": dataMap{
																"iterations": uint64(2),
																"selectNode": dataMap{
																	"iterations":    uint64(3),
																	"filterMatches": uint64(2),
																	"typeIndexJoin": dataMap{
																		"iterations": uint64(3),
																		"typeJoinOne": dataMap{
																			"root": dataMap{
																				"scanNode": dataMap{
																					"iterations":   uint64(3),
																					"docFetches":   uint64(2),
																					"fieldFetches": uint64(4),
																					"indexFetches": uint64(0),
																				},
																			},
																			// Publisher uses relation index (indexFetches: 2)
																			"subType": dataMap{
																				"selectTopNode": dataMap{
																					"selectNode": dataMap{
																						"iterations":    uint64(4),
																						"filterMatches": uint64(2),
																						"scanNode": dataMap{
																							"iterations":   uint64(4),
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
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

// Copyright 2022 Democratized Data Foundation
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

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	explainUtils "github.com/sourcenetwork/defradb/tests/integration/explain"
)

func TestExecuteExplainRequestWithOrderFieldOnParent(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (execute) with order field on parent.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			// Authors
			create2AuthorDocuments(),

			testUtils.ExplainRequest{
				Request: `query @explain(type: execute) {
					Author(order: {age: ASC}) {
						name
						age
					}
				}`,

				ExpectedFullGraph: []dataMap{
					{
						"explain": dataMap{
							"executionSuccess": true,
							"sizeOfResult":     2,
							"planExecutions":   uint64(3),
							"selectTopNode": dataMap{
								"orderNode": dataMap{
									"iterations": uint64(3),
									"selectNode": dataMap{
										"filterMatches": uint64(2),
										"iterations":    uint64(3),
										"scanNode": dataMap{
											"iterations":   uint64(3),
											"docFetches":   uint64(2),
											"fieldFetches": uint64(4),
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
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestExecuteExplainRequestWithMultiOrderFieldsOnParent(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (execute) with multiple order fields on parent.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			// Authors
			testUtils.CreateDoc{
				CollectionID: 2,

				Doc: `{
					"name": "Andy",
					"age": 64
				}`,
			},

			testUtils.CreateDoc{
				CollectionID: 2,

				Doc: `{
					"name": "Another64YearOld",
					"age": 64
				}`,
			},

			testUtils.CreateDoc{
				CollectionID: 2,

				Doc: `{
					"name": "Shahzad",
					"age": 48
				}`,
			},

			testUtils.CreateDoc{
				CollectionID: 2,

				Doc: `{
					"name": "Another48YearOld",
					"age": 48
				}`,
			},

			testUtils.ExplainRequest{
				Request: `query @explain(type: execute) {
					Author(order: {age: ASC, name: DESC}) {
						name
						age
					}
				}`,

				ExpectedFullGraph: []dataMap{
					{
						"explain": dataMap{
							"executionSuccess": true,
							"sizeOfResult":     4,
							"planExecutions":   uint64(5),
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
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestExecuteExplainRequestWithOrderFieldOnChild(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (execute) with order field on child.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			// Articles
			create3ArticleDocuments(),

			// Authors
			create2AuthorDocuments(),

			testUtils.ExplainRequest{
				Request: `query @explain(type: execute) {
					Author {
						name
						articles(order: {pages: DESC}) {
							pages
						}
					}
				}`,

				ExpectedFullGraph: []dataMap{
					{
						"explain": dataMap{
							"executionSuccess": true,
							"sizeOfResult":     2,
							"planExecutions":   uint64(3),
							"selectTopNode": dataMap{
								"selectNode": dataMap{
									"iterations":    uint64(3),
									"filterMatches": uint64(2),
									"typeIndexJoin": dataMap{
										"iterations": uint64(3),
										"scanNode": dataMap{
											"iterations":   uint64(3),
											"docFetches":   uint64(2),
											"fieldFetches": uint64(2),
											"indexFetches": uint64(0),
										},
										"subTypeScanNode": dataMap{
											"iterations":   uint64(5),
											"docFetches":   uint64(6),
											"fieldFetches": uint64(9),
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
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestExecuteExplainRequestWithOrderFieldOnBothParentAndChild(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (execute) with order field on both parent and child.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			// Articles
			create3ArticleDocuments(),

			// Authors
			create2AuthorDocuments(),

			testUtils.ExplainRequest{
				Request: `query @explain(type: execute) {
					Author(order: {age: ASC}) {
						name
						age
						articles(order: {pages: DESC}) {
							pages
						}
					}
				}`,

				ExpectedFullGraph: []dataMap{
					{
						"explain": dataMap{
							"executionSuccess": true,
							"sizeOfResult":     2,
							"planExecutions":   uint64(3),
							"selectTopNode": dataMap{
								"orderNode": dataMap{
									"iterations": uint64(3),
									"selectNode": dataMap{
										"iterations":    uint64(3),
										"filterMatches": uint64(2),
										"typeIndexJoin": dataMap{
											"iterations": uint64(3),
											"scanNode": dataMap{
												"iterations":   uint64(3),
												"docFetches":   uint64(2),
												"fieldFetches": uint64(4),
												"indexFetches": uint64(0),
											},
											"subTypeScanNode": dataMap{
												"iterations":   uint64(5),
												"docFetches":   uint64(6),
												"fieldFetches": uint64(9),
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

func TestExecuteExplainRequestWhereParentFieldIsOrderedByChildField(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (execute) where parent field is ordered by child field.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			// Articles
			create3ArticleDocuments(),

			// Authors
			create2AuthorDocuments(),

			testUtils.ExplainRequest{
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

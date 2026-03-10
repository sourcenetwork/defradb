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

package test_explain_debug

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	explainUtils "github.com/sourcenetwork/defradb/tests/integration/explain"
)

var orderTypeJoinPattern = dataMap{
	"root": dataMap{
		"scanNode": dataMap{},
	},
	"subType": dataMap{
		"selectTopNode": dataMap{
			"orderNode": dataMap{
				"selectNode": dataMap{
					"scanNode": dataMap{},
				},
			},
		},
	},
}

func TestDebugExplainRequestWithOrderFieldOnRelatedChild(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{

				Request: `query @explain(type: debug) {
					Author {
						name
						articles(order: {name: DESC}) {
							name
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
											"typeJoinMany": orderTypeJoinPattern,
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

func TestDebugExplainRequestWithOrderFieldOnParentAndRelatedChild(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{

				Request: `query @explain(type: debug) {
					Author(order: {name: ASC}) {
						name
						articles(order: {name: DESC}) {
							name
						}
					}
				}`,

				ExpectedPatterns: dataMap{
					"explain": dataMap{
						"operationNode": []dataMap{
							{
								"selectTopNode": dataMap{
									"orderNode": dataMap{
										"selectNode": dataMap{
											"typeIndexJoin": dataMap{
												"typeJoinMany": orderTypeJoinPattern,
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

func TestDebugExplainRequestWhereParentIsOrderedByItsRelatedChild(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{

				Request: `query @explain(type: debug) {
					Author(
						order: {
							articles: {name: ASC}
						}
					) {
						articles {
							name
						}
					}
				}`,

				ExpectedError: "Argument \"order\" has invalid value {articles: {name: ASC}}.\nIn field \"articles\": Unknown field.",
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

var nestedOrderByRelationPattern = dataMap{
	"root": dataMap{
		"scanNode": dataMap{},
	},
	"subType": dataMap{
		"selectTopNode": dataMap{
			"limitNode": dataMap{
				"orderNode": dataMap{
					"selectNode": dataMap{
						// Inner join: Book -> Publisher
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
}

func TestDebugExplainRequestWithSubqueryOrderByNestedRelationField(t *testing.T) {
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
										// Outer join: Author -> Book
										"typeIndexJoin": dataMap{
											"typeJoinMany": nestedOrderByRelationPattern,
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

func TestDebugExplainRequestWithSubqueryOrderByNestedRelationFieldASC(t *testing.T) {
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

			&action.ExplainRequest{
				Request: `query @explain(type: debug) {
					Author {
						name
						published(order: {publisher: {establishedYear: ASC}}, limit: 2) {
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
											"typeJoinMany": nestedOrderByRelationPattern,
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

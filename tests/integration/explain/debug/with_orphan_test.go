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

// orphanNodePattern is the expected pattern for a query that orders by a relation
// field with an index. The orphanNode wraps the typeJoinOne to handle orphan parents.
var orphanNodePattern = dataMap{
	"orphanNode": dataMap{
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
}

func TestDebugExplainRequestWithOrderByRelationFieldWithIndex(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Book {
						title: String
						rating: Int @index
						publisher: Publisher
					}
					type Publisher {
						name: String
						book: Book @primary
					}
				`,
			},

			&action.ExplainRequest{
				// @exhaustive is required to include orphanNode in the plan
				Request: `query @explain(type: debug) @exhaustive {
					Publisher(order: {book: {rating: ASC}}) {
						name
					}
				}`,

				ExpectedPatterns: dataMap{
					"explain": dataMap{
						"operationNode": []dataMap{
							{
								"selectTopNode": dataMap{
									"selectNode": dataMap{
										"typeIndexJoin": orphanNodePattern,
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

func TestDebugExplainRequestWithOrderByRelationFieldWithIndexDESC(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Book {
						title: String
						rating: Int @index
						publisher: Publisher
					}
					type Publisher {
						name: String
						book: Book @primary
					}
				`,
			},

			&action.ExplainRequest{
				// @exhaustive is required to include orphanNode in the plan
				Request: `query @explain(type: debug) @exhaustive {
					Publisher(order: {book: {rating: DESC}}) {
						name
					}
				}`,

				ExpectedPatterns: dataMap{
					"explain": dataMap{
						"operationNode": []dataMap{
							{
								"selectTopNode": dataMap{
									"selectNode": dataMap{
										"typeIndexJoin": orphanNodePattern,
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

func TestDebugExplainRequestWithOrderByRelationFieldSecondaryParent(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Book {
						title: String
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
				// @exhaustive is required to include orphanNode in the plan
				Request: `query @explain(type: debug) @exhaustive {
					Book(order: {publisher: {establishedYear: ASC}}) {
						title
					}
				}`,

				ExpectedPatterns: dataMap{
					"explain": dataMap{
						"operationNode": []dataMap{
							{
								"selectTopNode": dataMap{
									"selectNode": dataMap{
										"typeIndexJoin": orphanNodePattern,
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

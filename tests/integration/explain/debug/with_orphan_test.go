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

var joinOnePattern = dataMap{
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
}

// primaryParentASCPattern is for primary parent FK IS NULL path, ASC order:
// orphans come first in the sequence.
var primaryParentASCPattern = dataMap{
	"sequenceNode": []dataMap{
		{"orphanNode": dataMap{}},
		joinOnePattern,
	},
}

// primaryParentDESCPattern is for primary parent FK IS NULL path, DESC order:
// join results come first, orphans last.
var primaryParentDESCPattern = dataMap{
	"sequenceNode": []dataMap{
		joinOnePattern,
		{"orphanNode": dataMap{}},
	},
}

// secondaryParentPattern is for secondary parent exclusion path:
// orphanNode wraps the join and handles ordering internally.
var secondaryParentPattern = dataMap{
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
			&action.AddCollection{
				SDL: `
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
										"typeIndexJoin": primaryParentASCPattern,
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
			&action.AddCollection{
				SDL: `
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
										"typeIndexJoin": primaryParentDESCPattern,
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
			&action.AddCollection{
				SDL: `
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
										"typeIndexJoin": secondaryParentPattern,
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

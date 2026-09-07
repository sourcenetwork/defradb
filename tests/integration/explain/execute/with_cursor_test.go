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

var userSchemaWithIndex = &action.AddCollection{
	SDL: `
		type User {
			name: String
			age: Int @index
		}
	`,
}

func createUserDocuments() []*action.AddDoc {
	return []*action.AddDoc{
		{
			CollectionID: 0,
			Doc:          `{"name": "Alice", "age": 25}`,
		},
		{
			CollectionID: 0,
			Doc:          `{"name": "Bob", "age": 30}`,
		},
		{
			CollectionID: 0,
			Doc:          `{"name": "Carol", "age": 35}`,
		},
	}
}

func TestExecuteExplainCursorQueryWithFirst(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			userSchemaWithIndex,

			createUserDocuments(),

			&action.ExplainRequest{
				Request: `query @explain(type: execute) {
					_cursor {
						User(first: 2, order: {age: ASC}) {
							name
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
									"cursorNode": dataMap{
										"iterations": uint64(3),
										"selectNode": dataMap{
											"iterations":    uint64(3),
											"filterMatches": uint64(3),
											"scanNode": dataMap{
												"iterations":   uint64(3),
												"docFetches":   uint64(3),
												"fieldFetches": uint64(6),
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
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestExecuteExplainCursorQueryWithFirstReturnsAllWhenFewerExist(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			userSchemaWithIndex,

			createUserDocuments(),

			&action.ExplainRequest{
				Request: `query @explain(type: execute) {
					_cursor {
						User(first: 10, order: {age: ASC}) {
							name
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
									"cursorNode": dataMap{
										"iterations": uint64(4),
										"selectNode": dataMap{
											"iterations":    uint64(4),
											"filterMatches": uint64(3),
											"scanNode": dataMap{
												"iterations":   uint64(4),
												"docFetches":   uint64(3),
												"fieldFetches": uint64(6),
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
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

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

var userSchemaWithIndex = &action.AddSchema{
	Schema: `
		type User {
			name: String
			age: Int @index
		}
	`,
}

func createUserDocuments() []testUtils.CreateDoc {
	return []testUtils.CreateDoc{
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

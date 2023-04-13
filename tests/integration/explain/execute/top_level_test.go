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
)

func TestExecuteExplainTopLevelAverageRequest(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (execute) request with top level average.",

		Actions: []any{
			gqlSchemaExecuteExplain(),

			testUtils.CreateDoc{
				CollectionID: 2,

				// bae-111e8e29-0530-52ae-815f-14c7ba46d277
				Doc: `{
					"name": "Andy",
					"age": 64
				}`,
			},

			testUtils.CreateDoc{
				CollectionID: 2,

				// bae-e147be24-bf9c-5d38-8c7b-ad18e4034c53
				Doc: `{
					"name": "Shahzad",
					"age": 48
				}`,
			},

			testUtils.Request{
				Request: `query @explain(type: execute) {
					_avg(
						Author: {
							field: age
						}
					)
				}`,

				Results: []dataMap{
					{
						"explain": dataMap{
							"executionSuccess": true,
							"sizeOfResult":     1,
							"planExecutions":   uint64(2),
							"topLevelNode": []dataMap{
								{
									"selectTopNode": dataMap{
										"selectNode": dataMap{
											"iterations":    uint64(3),
											"filterMatches": uint64(2),
											"scanNode": dataMap{
												"iterations":    uint64(3),
												"docFetches":    uint64(3),
												"filterMatches": uint64(2),
											},
										},
									},
								},

								{
									"sumNode": dataMap{
										"iterations": uint64(1),
									},
								},

								{
									"countNode": dataMap{
										"iterations": uint64(1),
									},
								},

								{
									"averageNode": dataMap{

										"iterations": uint64(1),
									},
								},
							},
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestExecuteExplainTopLevelCountRequest(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (execute) request with top level count.",

		Actions: []any{
			gqlSchemaExecuteExplain(),

			testUtils.CreateDoc{
				CollectionID: 2,

				// bae-111e8e29-0530-52ae-815f-14c7ba46d277
				Doc: `{
					"name": "Andy",
					"age": 64
				}`,
			},

			testUtils.CreateDoc{
				CollectionID: 2,

				// bae-e147be24-bf9c-5d38-8c7b-ad18e4034c53
				Doc: `{
					"name": "Shahzad",
					"age": 48
				}`,
			},

			testUtils.Request{
				Request: `query @explain(type: execute) {
					_count(Author: {})
				}`,

				Results: []dataMap{
					{
						"explain": dataMap{
							"executionSuccess": true,
							"sizeOfResult":     1,
							"planExecutions":   uint64(2),
							"topLevelNode": []dataMap{
								{
									"selectTopNode": dataMap{
										"selectNode": dataMap{
											"iterations":    uint64(3),
											"filterMatches": uint64(2),
											"scanNode": dataMap{
												"iterations":    uint64(3),
												"docFetches":    uint64(3),
												"filterMatches": uint64(2),
											},
										},
									},
								},

								{
									"countNode": dataMap{
										"iterations": uint64(1),
									},
								},
							},
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestExecuteExplainTopLevelSumRequest(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (execute) request with top level sum.",

		Actions: []any{
			gqlSchemaExecuteExplain(),

			testUtils.CreateDoc{
				CollectionID: 2,

				// bae-111e8e29-0530-52ae-815f-14c7ba46d277
				Doc: `{
					"name": "Andy",
					"age": 64
				}`,
			},

			testUtils.CreateDoc{
				CollectionID: 2,

				// bae-e147be24-bf9c-5d38-8c7b-ad18e4034c53
				Doc: `{
					"name": "Shahzad",
					"age": 48
				}`,
			},

			testUtils.Request{
				Request: `query @explain(type: execute) {
					_sum(
						Author: {
							field: age
						}
					)
				}`,

				Results: []dataMap{
					{
						"explain": dataMap{
							"executionSuccess": true,
							"sizeOfResult":     1,
							"planExecutions":   uint64(2),
							"topLevelNode": []dataMap{
								{
									"selectTopNode": dataMap{
										"selectNode": dataMap{
											"iterations":    uint64(3),
											"filterMatches": uint64(2),
											"scanNode": dataMap{
												"iterations":    uint64(3),
												"docFetches":    uint64(3),
												"filterMatches": uint64(2),
											},
										},
									},
								},

								{
									"sumNode": dataMap{
										"iterations": uint64(1),
									},
								},
							},
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

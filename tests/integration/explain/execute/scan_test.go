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

func TestExecuteExplainRequestWithAllDocumentsMatching(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (execute) request with all documents matching.",

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
					Author {
						name
						age
					}
				}`,

				Results: []dataMap{
					{
						"explain": dataMap{
							"executionSuccess": true,
							"sizeOfResult":     2,
							"planExecutions":   uint64(3),
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
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestExecuteExplainRequestWithNoDocuments(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (execute) request with no documents.",

		Actions: []any{
			gqlSchemaExecuteExplain(),

			testUtils.Request{
				Request: `query @explain(type: execute) {
					Author {
						name
					}
				}`,

				Results: []dataMap{
					{
						"explain": dataMap{
							"executionSuccess": true,
							"sizeOfResult":     0,
							"planExecutions":   uint64(1),
							"selectTopNode": dataMap{
								"selectNode": dataMap{
									"iterations":    uint64(1),
									"filterMatches": uint64(0),
									"scanNode": dataMap{
										"iterations":    uint64(1),
										"docFetches":    uint64(1),
										"filterMatches": uint64(0),
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

func TestExecuteExplainRequestWithSomeDocumentsMatching(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (execute) request with some documents matching.",

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
					Author(filter: {name: {_eq: "Shahzad"}}) {
						name
						age
					}
				}`,

				Results: []dataMap{
					{
						"explain": dataMap{
							"executionSuccess": true,
							"planExecutions":   uint64(2),
							"sizeOfResult":     1,
							"selectTopNode": dataMap{
								"selectNode": dataMap{
									"iterations":    uint64(2),
									"filterMatches": uint64(1),
									"scanNode": dataMap{
										"iterations":    uint64(2),
										"docFetches":    uint64(3),
										"filterMatches": uint64(1),
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

func TestExecuteExplainRequestWithDocumentsButNoMatches(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (execute) request with documents but no matches.",

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
					Author(filter: {name: {_eq: "John"}}) {
						name
						age
					}
				}`,

				Results: []dataMap{
					{
						"explain": dataMap{
							"executionSuccess": true,
							"planExecutions":   uint64(1),
							"sizeOfResult":     0,
							"selectTopNode": dataMap{
								"selectNode": dataMap{
									"iterations":    uint64(1),
									"filterMatches": uint64(0),
									"scanNode": dataMap{
										"iterations":    uint64(1),
										"docFetches":    uint64(3),
										"filterMatches": uint64(0),
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

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

func TestExecuteExplainRequestWithAOneToOneJoin(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain a one-to-one join relation query, with alias.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			// Authors
			create2AuthorDocuments(),

			// Contacts
			create2AuthorContactDocuments(),

			testUtils.ExplainRequest{
				Request: `query @explain(type: execute) {
					Author {
						OnlyEmail: contact {
							email
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
											"iterations":   uint64(2),
											"docFetches":   uint64(2),
											"fieldFetches": uint64(2),
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

func TestExecuteExplainWithMultipleOneToOneJoins(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (execute) with two one-to-one join relation.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			// Authors
			create2AuthorDocuments(),

			// Contacts
			create2AuthorContactDocuments(),

			testUtils.ExplainRequest{
				Request: `query @explain(type: execute) {
					Author {
						OnlyEmail: contact {
							email
						}
						contact {
							cell
							email
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
									"parallelNode": []dataMap{
										{
											"typeIndexJoin": dataMap{
												"iterations": uint64(3),
												"scanNode": dataMap{
													"iterations":   uint64(3),
													"docFetches":   uint64(2),
													"fieldFetches": uint64(2),
													"indexFetches": uint64(0),
												},
												"subTypeScanNode": dataMap{
													"iterations":   uint64(2),
													"docFetches":   uint64(2),
													"fieldFetches": uint64(2),
													"indexFetches": uint64(0),
												},
											},
										},
										{
											"typeIndexJoin": dataMap{
												"iterations": uint64(3),
												"scanNode": dataMap{
													"iterations":   uint64(3),
													"docFetches":   uint64(2),
													"fieldFetches": uint64(2),
													"indexFetches": uint64(0),
												},
												"subTypeScanNode": dataMap{
													"iterations":   uint64(2),
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
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestExecuteExplainWithTwoLevelDeepNestedJoins(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (execute) with two nested level deep one to one join.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			// Authors
			create2AuthorDocuments(),

			// Contacts
			create2AuthorContactDocuments(),

			// Addresses
			create2AddressDocuments(),

			testUtils.ExplainRequest{
				Request: `query @explain(type: execute) {
					Author {
						name
						contact {
							email
							address {
								city
							}
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
											"fieldFetches": uint64(4),
											"indexFetches": uint64(0),
										},
										"subTypeScanNode": dataMap{
											"iterations":   uint64(2),
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

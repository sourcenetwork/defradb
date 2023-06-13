// Copyright 2023 Democratized Data Foundation
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

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	explainUtils "github.com/sourcenetwork/defradb/tests/integration/explain"
)

var normalTypeJoinPattern = dataMap{
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
}

var debugTypeJoinPattern = dataMap{
	"root": dataMap{
		"multiScanNode": dataMap{
			"scanNode": dataMap{},
		},
	},
	"subType": dataMap{
		"selectTopNode": dataMap{
			"selectNode": dataMap{
				"scanNode": dataMap{},
			},
		},
	},
}

func TestDebugExplainRequestWith2SingleJoinsAnd1ManyJoin(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (debug) request with 2 single joins and 1 many join.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain(type: debug) {
					Author {
						OnlyEmail: contact {
							email
						}
						articles {
							name
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
							"selectTopNode": dataMap{
								"selectNode": dataMap{
									"parallelNode": []dataMap{
										{
											"typeIndexJoin": dataMap{
												"typeJoinOne": debugTypeJoinPattern,
											},
										},
										{
											"typeIndexJoin": dataMap{
												"typeJoinMany": debugTypeJoinPattern,
											},
										},
										{
											"typeIndexJoin": dataMap{
												"typeJoinOne": debugTypeJoinPattern,
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

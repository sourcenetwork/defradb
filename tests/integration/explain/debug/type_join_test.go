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

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{

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

				ExpectedFullGraph: dataMap{
					"explain": dataMap{
						"operationNode": []dataMap{
							{
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
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

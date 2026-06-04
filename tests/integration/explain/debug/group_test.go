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

var groupPattern = dataMap{
	"explain": dataMap{
		"operationNode": []dataMap{
			{
				"selectTopNode": dataMap{
					"groupNode": dataMap{
						"selectNode": dataMap{
							"pipeNode": dataMap{
								"scanNode": dataMap{},
							},
						},
					},
				},
			},
		},
	},
}

func TestDebugExplainRequestWithGroupByOnParent(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{

				Request: `query @explain(type: debug) {
					Author (groupBy: [age]) {
						age
						GROUP {
							name
						}
					}
				}`,

				ExpectedFullGraph: groupPattern,
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDebugExplainRequestWithGroupByTwoFieldsOnParent(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{

				Request: `query @explain(type: debug) {
					Author (groupBy: [age, name]) {
						age
						GROUP {
							name
						}
					}
				}`,

				ExpectedFullGraph: groupPattern,
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

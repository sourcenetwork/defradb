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

var debugGroupLimitPattern = dataMap{
	"explain": dataMap{
		"operationNode": []dataMap{
			{
				"selectTopNode": dataMap{
					"limitNode": dataMap{
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
	},
}

func TestDebugExplainRequestWithLimitAndOffsetOnParentGroupBy(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{

				Request: `query @explain(type: debug) {
					Author(
						groupBy: [name],
						limit: 1,
						offset: 1
					) {
						name
						GROUP {
							age
						}
					}
				}`,

				ExpectedPatterns: debugGroupLimitPattern,
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDebugExplainRequestWithLimitOnParentGroupByAndInnerGroupSelection(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{

				Request: `query @explain(type: debug) {
					Author(
						groupBy: [name],
						limit: 1
					) {
						name
						GROUP(limit: 2) {
							age
						}
					}
				}`,

				ExpectedPatterns: debugGroupLimitPattern,
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

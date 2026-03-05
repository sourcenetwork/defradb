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

package test_explain_simple

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	explainUtils "github.com/sourcenetwork/defradb/tests/integration/explain"
)

type dataMap = map[string]any

func TestSimpleExplainRequest(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{
				Request: `query @explain(type: simple) {
					Author {
						_docID
						name
						age
					}
				}`,

				ExpectedFullGraph: dataMap{
					"explain": dataMap{
						"operationNode": []dataMap{
							{
								"selectTopNode": dataMap{
									"selectNode": dataMap{
										"docID":  nil,
										"filter": nil,
										"scanNode": dataMap{
											"filter":         nil,
											"collectionID":   "bafyreid73sgzodav5hxhrsypjapj6r2uzo7mhm3vqykjhfehj7i5hhksuu",
											"collectionName": "Author",
											"prefixes": []string{
												"/3",
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

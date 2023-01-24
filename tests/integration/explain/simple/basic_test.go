// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_explain_simple

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestSimpleExplainRequest(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Explain (simple) a basic request.",

		Request: `query @explain(type: simple) {
			author {
				_key
				name
				age
			}
		}`,

		Docs: map[int][]string{
			2: {
				`{
					"name": "John",
					"age": 21
				}`,
			},
		},

		Results: []dataMap{
			{
				"explain": dataMap{
					"selectTopNode": dataMap{
						"selectNode": dataMap{
							"filter": nil,
							"scanNode": dataMap{
								"filter":         nil,
								"collectionID":   "3",
								"collectionName": "author",
								"spans": []dataMap{
									{
										"start": "/3",
										"end":   "/4",
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

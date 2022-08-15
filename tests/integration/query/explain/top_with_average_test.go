// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_explain

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestExplainTopLevelAverageQuery(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Explain top-level average query.",

		Query: `query @explain {
					_avg(
						author: {
							field: age
						}
					)
				}`,

		Docs: map[int][]string{
			//authors
			2: {
				`{
					"name": "John",
					"verified": false,
					"age": 28
				}`,
				`{
					"name": "Bob",
					"verified": true,
					"age": 30
				}`,
			},
		},

		Results: []dataMap{
			{
				"explain": dataMap{
					"topLevelNode": dataMap{},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestExplainTopLevelAverageQueryWithFilter(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Explain top-level average query with filter.",
		Query: `query @explain {
					_avg(
						author: {
							field: age,
							filter: {
								age: {
									_gt: 26
								}
							}
						}
					)
				}`,

		Docs: map[int][]string{
			//authors
			2: {
				`{
					"name": "John",
					"verified": false,
					"age": 21
				}`,
				`{
					"name": "Bob",
					"verified": false,
					"age": 30
				}`,
				`{
					"name": "Alice",
					"verified": false,
					"age": 32
				}`,
			},
		},

		Results: []dataMap{
			{
				"explain": dataMap{
					"topLevelNode": dataMap{},
				},
			},
		},
	}

	executeTestCase(t, test)
}

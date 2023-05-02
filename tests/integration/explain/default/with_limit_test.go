// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_explain_default

import (
	"testing"

	explainUtils "github.com/sourcenetwork/defradb/tests/integration/explain"
)

var limitPattern = dataMap{
	"explain": dataMap{
		"selectTopNode": dataMap{
			"limitNode": dataMap{
				"selectNode": dataMap{
					"scanNode": dataMap{},
				},
			},
		},
	},
}

func TestDefaultExplainRequestWithOnlyLimit(t *testing.T) {
	test := explainUtils.ExplainRequestTestCase{

		Description: "Explain (default) request with only limit.",

		Request: `query @explain {
			author(limit: 2) {
				name
			}
		}`,

		Docs: map[int][]string{
			// authors
			2: {
				// _key: bae-41598f0c-19bc-5da6-813b-e80f14a10df3
				`{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,

				// _key: bae-aa839756-588e-5b57-887d-33689a06e375
				`{
					"name": "Shahzad Sisley",
					"age": 26,
					"verified": true
				}`,

				// _key: bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04
				`{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`,

				// _key: bae-e7e87bbb-1079-59db-b4b9-0e14b24d5b69
				`{
					"name": "Andrew Lone",
					"age": 28,
					"verified": true
				}`,
			},
		},

		ExpectedPatterns: []dataMap{limitPattern},

		ExpectedTargets: []explainUtils.PlanNodeTargetCase{
			{
				TargetNodeName:    "limitNode",
				IncludeChildNodes: false,
				ExpectedAttributes: dataMap{
					"limit":  uint64(2),
					"offset": uint64(0),
				},
			},
		},
	}

	runExplainTest(t, test)
}

func TestDefaultExplainRequestWithOnlyOffset(t *testing.T) {
	test := explainUtils.ExplainRequestTestCase{

		Description: "Explain (default) request with only offset.",

		Request: `query @explain {
			author(offset: 2) {
				name
			}
		}`,

		Docs: map[int][]string{
			// authors
			2: {
				// _key: bae-41598f0c-19bc-5da6-813b-e80f14a10df3
				`{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,

				// _key: bae-aa839756-588e-5b57-887d-33689a06e375
				`{
					"name": "Shahzad Sisley",
					"age": 26,
					"verified": true
				}`,

				// _key: bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04
				`{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`,

				// _key: bae-e7e87bbb-1079-59db-b4b9-0e14b24d5b69
				`{
					"name": "Andrew Lone",
					"age": 28,
					"verified": true
				}`,
			},
		},

		ExpectedPatterns: []dataMap{limitPattern},

		ExpectedTargets: []explainUtils.PlanNodeTargetCase{
			{
				TargetNodeName:    "limitNode",
				IncludeChildNodes: false,
				ExpectedAttributes: dataMap{
					"limit":  nil,
					"offset": uint64(2),
				},
			},
		},
	}

	runExplainTest(t, test)
}

func TestDefaultExplainRequestWithLimitAndOffset(t *testing.T) {
	test := explainUtils.ExplainRequestTestCase{

		Description: "Explain (default) request with limit and offset.",

		Request: `query @explain {
			author(limit: 3, offset: 1) {
				name
			}
		}`,

		Docs: map[int][]string{
			// authors
			2: {
				// _key: bae-41598f0c-19bc-5da6-813b-e80f14a10df3
				`{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,

				// _key: bae-aa839756-588e-5b57-887d-33689a06e375
				`{
					"name": "Shahzad Sisley",
					"age": 26,
					"verified": true
				}`,

				// _key: bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04
				`{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`,

				// _key: bae-e7e87bbb-1079-59db-b4b9-0e14b24d5b69
				`{
					"name": "Andrew Lone",
					"age": 28,
					"verified": true
				}`,
			},
		},

		ExpectedPatterns: []dataMap{limitPattern},

		ExpectedTargets: []explainUtils.PlanNodeTargetCase{
			{
				TargetNodeName:    "limitNode",
				IncludeChildNodes: false,
				ExpectedAttributes: dataMap{
					"limit":  uint64(3),
					"offset": uint64(1),
				},
			},
		},
	}

	runExplainTest(t, test)
}

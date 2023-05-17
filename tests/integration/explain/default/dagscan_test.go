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

var dagScanPattern = dataMap{
	"explain": dataMap{
		"selectTopNode": dataMap{
			"selectNode": dataMap{
				"dagScanNode": dataMap{},
			},
		},
	},
}

func TestDefaultExplainCommitsDagScanQueryOp(t *testing.T) {
	test := explainUtils.ExplainRequestTestCase{

		Description: "Explain (default) commits query-op.",

		Request: `query @explain {
			commits (dockey: "bae-41598f0c-19bc-5da6-813b-e80f14a10df3", fieldId: "1") {
				links {
					cid
				}
			}
		}`,

		Docs: map[int][]string{
			//authors
			2: {
				// bae-41598f0c-19bc-5da6-813b-e80f14a10df3
				`{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
				// bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04
				`{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`,
			},
		},

		ExpectedPatterns: []dataMap{dagScanPattern},

		ExpectedTargets: []explainUtils.PlanNodeTargetCase{
			{
				TargetNodeName:    "dagScanNode",
				IncludeChildNodes: true, // Shouldn't have any as this is the last node in the chain.
				ExpectedAttributes: dataMap{
					"cid":     nil,
					"fieldId": "1",
					"spans": []dataMap{
						{
							"start": "/bae-41598f0c-19bc-5da6-813b-e80f14a10df3/1",
							"end":   "/bae-41598f0c-19bc-5da6-813b-e80f14a10df3/2",
						},
					},
				},
			},
		},
	}

	runExplainTest(t, test)
}

func TestDefaultExplainCommitsDagScanQueryOpWithoutField(t *testing.T) {
	test := explainUtils.ExplainRequestTestCase{

		Description: "Explain (default) commits query-op with only dockey (no field).",

		Request: `query @explain {
			commits (dockey: "bae-41598f0c-19bc-5da6-813b-e80f14a10df3") {
				links {
					cid
				}
			}
		}`,

		Docs: map[int][]string{
			//authors
			2: {
				// bae-41598f0c-19bc-5da6-813b-e80f14a10df3
				`{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
				// bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04
				`{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`,
			},
		},

		ExpectedPatterns: []dataMap{dagScanPattern},

		ExpectedTargets: []explainUtils.PlanNodeTargetCase{
			{
				TargetNodeName:    "dagScanNode",
				IncludeChildNodes: true, // Shouldn't have any as this is the last node in the chain.
				ExpectedAttributes: dataMap{
					"cid":     nil,
					"fieldId": nil,
					"spans": []dataMap{
						{
							"start": "/bae-41598f0c-19bc-5da6-813b-e80f14a10df3",
							"end":   "/bae-41598f0c-19bc-5da6-813b-e80f14a10df4",
						},
					},
				},
			},
		},
	}

	runExplainTest(t, test)
}

func TestDefaultExplainLatestCommitsDagScanQueryOp(t *testing.T) {
	test := explainUtils.ExplainRequestTestCase{

		Description: "Explain (default) latestCommits query-op.",

		Request: `query @explain {
			latestCommits(dockey: "bae-41598f0c-19bc-5da6-813b-e80f14a10df3", fieldId: "1") {
				cid
				links {
					cid
				}
			}
		}`,

		Docs: map[int][]string{
			//authors
			2: {
				// bae-41598f0c-19bc-5da6-813b-e80f14a10df3
				`{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
				// bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04
				`{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`,
			},
		},

		ExpectedPatterns: []dataMap{dagScanPattern},

		ExpectedTargets: []explainUtils.PlanNodeTargetCase{
			{
				TargetNodeName:    "dagScanNode",
				IncludeChildNodes: true, // Shouldn't have any as this is the last node in the chain.
				ExpectedAttributes: dataMap{
					"cid":     nil,
					"fieldId": "1",
					"spans": []dataMap{
						{
							"start": "/bae-41598f0c-19bc-5da6-813b-e80f14a10df3/1",
							"end":   "/bae-41598f0c-19bc-5da6-813b-e80f14a10df3/2",
						},
					},
				},
			},
		},
	}

	runExplainTest(t, test)
}

func TestDefaultExplainLatestCommitsDagScanQueryOpWithoutField(t *testing.T) {
	test := explainUtils.ExplainRequestTestCase{

		Description: "Explain (default) latestCommits query-op with only dockey (no field).",

		Request: `query @explain {
			latestCommits(dockey: "bae-41598f0c-19bc-5da6-813b-e80f14a10df3") {
				cid
				links {
					cid
				}
			}
		}`,

		Docs: map[int][]string{
			//authors
			2: {
				// bae-41598f0c-19bc-5da6-813b-e80f14a10df3
				`{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
				// bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04
				`{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`,
			},
		},

		ExpectedPatterns: []dataMap{dagScanPattern},

		ExpectedTargets: []explainUtils.PlanNodeTargetCase{
			{
				TargetNodeName:    "dagScanNode",
				IncludeChildNodes: true, // Shouldn't have any as this is the last node in the chain.
				ExpectedAttributes: dataMap{
					"cid":     nil,
					"fieldId": "C",
					"spans": []dataMap{
						{
							"start": "/bae-41598f0c-19bc-5da6-813b-e80f14a10df3/C",
							"end":   "/bae-41598f0c-19bc-5da6-813b-e80f14a10df3/D",
						},
					},
				},
			},
		},
	}

	runExplainTest(t, test)
}

func TestDefaultExplainLatestCommitsDagScanWithoutDocKey_Failure(t *testing.T) {
	test := explainUtils.ExplainRequestTestCase{

		Description: "Explain (default) latestCommits query without DocKey.",

		Request: `query @explain {
			latestCommits(fieldId: "1") {
				cid
				links {
					cid
				}
			}
		}`,

		ExpectedError: "Field \"latestCommits\" argument \"dockey\" of type \"ID!\" is required but not provided.",
	}

	runExplainTest(t, test)
}

func TestDefaultExplainLatestCommitsDagScanWithoutAnyArguments_Failure(t *testing.T) {
	test := explainUtils.ExplainRequestTestCase{

		Description: "Explain (default) latestCommits query without any arguments.",

		Request: `query @explain {
			latestCommits {
				cid
				links {
					cid
				}
			}
		}`,

		ExpectedError: "Field \"latestCommits\" argument \"dockey\" of type \"ID!\" is required but not provided.",
	}

	runExplainTest(t, test)
}

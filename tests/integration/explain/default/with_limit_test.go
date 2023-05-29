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
			Author(limit: 2) {
				name
			}
		}`,

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

	explainUtils.RunExplainTest(t, test)
}

func TestDefaultExplainRequestWithOnlyOffset(t *testing.T) {
	test := explainUtils.ExplainRequestTestCase{

		Description: "Explain (default) request with only offset.",

		Request: `query @explain {
			Author(offset: 2) {
				name
			}
		}`,

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

	explainUtils.RunExplainTest(t, test)
}

func TestDefaultExplainRequestWithLimitAndOffset(t *testing.T) {
	test := explainUtils.ExplainRequestTestCase{

		Description: "Explain (default) request with limit and offset.",

		Request: `query @explain {
			Author(limit: 3, offset: 1) {
				name
			}
		}`,

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

	explainUtils.RunExplainTest(t, test)
}

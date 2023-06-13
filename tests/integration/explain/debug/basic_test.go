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

func TestDebugExplainRequest(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Explain (debug) a basic request, assert full graph.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{
				Request: `query @explain(type: debug) {
					Author {
						name
						age
					}
				}`,

				ExpectedFullGraph: []dataMap{basicPattern},
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

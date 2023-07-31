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

func TestDebugExplainRequestWithLimitAndOffsetOnInnerGroupSelection(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (debug) request with limit and offset on inner _group selection.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain(type: debug) {
					Author(groupBy: [name]) {
						name
						_group(limit: 2, offset: 1) {
							age
						}
					}
				}`,

				ExpectedPatterns: []dataMap{groupPattern},
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDebugExplainRequestWithLimitAndOffsetOnMultipleInnerGroupSelections(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (debug) request with limit and offset on multiple inner _group selections.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain(type: debug) {
					Author(groupBy: [name]) {
						name
						innerFirstGroup: _group(limit: 1, offset: 2) {
							age
						}
						innerSecondGroup: _group(limit: 2) {
							age
						}
					}
				}`,

				ExpectedPatterns: []dataMap{groupPattern},
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	explainUtils "github.com/sourcenetwork/defradb/tests/integration/explain"
)

func TestDebugExplainRequestWithLimitAndOffsetOnInnerGroupSelection(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{

				Request: `query @explain(type: debug) {
					Author(groupBy: [name]) {
						name
						GROUP(limit: 2, offset: 1) {
							age
						}
					}
				}`,

				ExpectedPatterns: groupPattern,
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDebugExplainRequestWithLimitAndOffsetOnMultipleInnerGroupSelections(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{

				Request: `query @explain(type: debug) {
					Author(groupBy: [name]) {
						name
						innerFirstGroup: GROUP(limit: 1, offset: 2) {
							age
						}
						innerSecondGroup: GROUP(limit: 2) {
							age
						}
					}
				}`,

				ExpectedPatterns: groupPattern,
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

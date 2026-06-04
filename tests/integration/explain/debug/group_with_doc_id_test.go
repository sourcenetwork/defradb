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

func TestDebugExplainRequestWithDocIDOnParentGroupBy(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{

				Request: `query @explain(type: debug) {
					Author(
						groupBy: [age],
						docID: "bae-6a4c5bc5-b044-5a03-a868-8260af6f2254"
					) {
						age
						GROUP {
							name
						}
					}
				}`,

				ExpectedPatterns: groupPattern,
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDebugExplainRequestWithDocIDsAndFilterOnParentGroupBy(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{

				Request: `query @explain(type: debug) {
					Author(
						groupBy: [age],
						filter: {age: {_eq: 20}},
						docID: [
							"bae-6a4c5bc5-b044-5a03-a868-8260af6f2254",
							"bae-4ea9d148-13f3-5a48-a0ef-9ffd344caeed"
						]
					) {
						age
						GROUP {
							name
						}
					}
				}`,

				ExpectedPatterns: groupPattern,
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

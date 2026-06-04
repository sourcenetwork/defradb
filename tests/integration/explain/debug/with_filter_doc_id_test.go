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

func TestDebugExplainRequestWithDocIDFilter(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{

				Request: `query @explain(type: debug) {
					Author(docID: "bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d") {
						name
						age
					}
				}`,

				ExpectedPatterns: basicPattern,
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDebugExplainRequestWithDocIDsFilterUsingOneID(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{

				Request: `query @explain(type: debug) {
					Author(docID: ["bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d"]) {
						name
						age
					}
				}`,

				ExpectedPatterns: basicPattern,
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDebugExplainRequestWithDocIDsFilterUsingMultipleButDuplicateIDs(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{

				Request: `query @explain(type: debug) {
					Author(
						docID: [
							"bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d",
							"bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d"
						]
					) {
						name
						age
					}
				}`,

				ExpectedPatterns: basicPattern,
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDebugExplainRequestWithDocIDsFilterUsingMultipleUniqueIDs(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{

				Request: `query @explain(type: debug) {
					Author(
						docID: [
							"bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d",
							"bae-bfbfc89c-0d63-5ea4-81a3-3ebd295be67f"
						]
					) {
						name
						age
					}
				}`,

				ExpectedPatterns: basicPattern,
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDebugExplainRequestWithMatchingIDFilter(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{

				Request: `query @explain(type: debug) {
					Author(
						filter: {
							_docID: {
								_eq: "bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d"
							}
						}
					) {
						name
						age
					}
				}`,

				ExpectedPatterns: basicPattern,
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

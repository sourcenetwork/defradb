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

func TestDebugExplainRequestWithDocKeyFilter(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (debug) request with dockey filter.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain(type: debug) {
					Author(dockey: "bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d") {
						name
						age
					}
				}`,

				ExpectedPatterns: []dataMap{basicPattern},
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDebugExplainRequestWithDocKeysFilterUsingOneKey(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (debug) request with dockeys filter using one key.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain(type: debug) {
					Author(dockeys: ["bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d"]) {
						name
						age
					}
				}`,

				ExpectedPatterns: []dataMap{basicPattern},
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDebugExplainRequestWithDocKeysFilterUsingMultipleButDuplicateKeys(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (debug) request with dockeys filter using multiple but duplicate keys.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain(type: debug) {
					Author(
						dockeys: [
							"bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d",
							"bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d"
						]
					) {
						name
						age
					}
				}`,

				ExpectedPatterns: []dataMap{basicPattern},
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDebugExplainRequestWithDocKeysFilterUsingMultipleUniqueKeys(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (debug) request with dockeys filter using multiple unique keys.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain(type: debug) {
					Author(
						dockeys: [
							"bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d",
							"bae-bfbfc89c-0d63-5ea4-81a3-3ebd295be67f"
						]
					) {
						name
						age
					}
				}`,

				ExpectedPatterns: []dataMap{basicPattern},
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDebugExplainRequestWithMatchingKeyFilter(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (debug) request with a filter to match key.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

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

				ExpectedPatterns: []dataMap{basicPattern},
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

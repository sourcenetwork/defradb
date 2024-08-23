// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package simple

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQuerySimpleWithDocIDsFilter(t *testing.T) {
	tests := []testUtils.TestCase{
		{
			Description: "Simple query with basic filter (single ID by docIDs arg)",
			Actions: []any{
				testUtils.CreateDoc{
					Doc: `{
						"Name": "John",
						"Age": 21
					}`,
				},
				testUtils.Request{
					Request: `query {
						Users(docIDs: ["bae-d4303725-7db9-53d2-b324-f3ee44020e52"]) {
							Name
							Age
						}
					}`,
					Results: map[string]any{
						"Users": []map[string]any{
							{
								"Name": "John",
								"Age":  int64(21),
							},
						},
					},
				},
			},
		},
		{
			Description: "Simple query with basic filter (single ID by docIDs arg), no results",
			Actions: []any{
				testUtils.CreateDoc{
					Doc: `{
						"Name": "John",
						"Age": 21
					}`,
				},
				testUtils.Request{
					Request: `query {
						Users(docIDs: ["bae-52b9170d-b77a-5887-b877-cbdbb99b009g"]) {
							Name
							Age
						}
					}`,
					Results: map[string]any{
						"Users": []map[string]any{},
					},
				},
			},
		},
		{
			Description: "Simple query with basic filter (duplicate ID by docIDs arg), partial results",
			Actions: []any{
				testUtils.CreateDoc{
					Doc: `{
						"Name": "John",
						"Age": 21
					}`,
				},
				testUtils.CreateDoc{
					Doc: `{
						"Name": "Bob",
						"Age": 32
					}`,
				},
				testUtils.Request{
					Request: `query {
						Users(docIDs: ["bae-d4303725-7db9-53d2-b324-f3ee44020e52", "bae-d4303725-7db9-53d2-b324-f3ee44020e52"]) {
							Name
							Age
						}
					}`,
					Results: map[string]any{
						"Users": []map[string]any{
							{
								"Name": "John",
								"Age":  int64(21),
							},
						},
					},
				},
			},
		},
		{
			Description: "Simple query with basic filter (multiple ID by docIDs arg), partial results",
			Actions: []any{
				testUtils.CreateDoc{
					Doc: `{
						"Name": "John",
						"Age": 21
					}`,
				},
				testUtils.CreateDoc{
					Doc: `{
						"Name": "Bob",
						"Age": 32
					}`,
				},
				testUtils.CreateDoc{
					Doc: `{
						"Name": "Jim",
						"Age": 27
					}`,
				},
				testUtils.Request{
					Request: `query {
						Users(docIDs: ["bae-d4303725-7db9-53d2-b324-f3ee44020e52", "bae-428c6d76-3491-520b-ad1f-a218f4dad787"]) {
							Name
							Age
						}
					}`,
					Results: map[string]any{
						"Users": []map[string]any{
							{
								"Name": "Jim",
								"Age":  int64(27),
							},
							{
								"Name": "John",
								"Age":  int64(21),
							},
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		executeTestCase(t, test)
	}
}

func TestQuerySimpleReturnsNothinGivenEmptyDocIDsFilter(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with empty docIDs arg",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(docIDs: []) {
						Name
						Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{},
				},
			},
		},
	}

	executeTestCase(t, test)
}

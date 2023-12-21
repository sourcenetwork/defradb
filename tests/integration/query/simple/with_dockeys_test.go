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
	tests := []testUtils.RequestTestCase{
		{
			Description: "Simple query with basic filter (single ID by docIDs arg)",
			Request: `query {
						Users(docIDs: ["bae-52b9170d-b77a-5887-b877-cbdbb99b009f"]) {
							Name
							Age
						}
					}`,
			Docs: map[int][]string{
				0: {
					`{
						"Name": "John",
						"Age": 21
					}`,
				},
			},
			Results: []map[string]any{
				{
					"Name": "John",
					"Age":  int64(21),
				},
			},
		},
		{
			Description: "Simple query with basic filter (single ID by docIDs arg), no results",
			Request: `query {
						Users(docIDs: ["bae-52b9170d-b77a-5887-b877-cbdbb99b009g"]) {
							Name
							Age
						}
					}`,
			Docs: map[int][]string{
				0: {
					`{
						"Name": "John",
						"Age": 21
					}`,
				},
			},
			Results: []map[string]any{},
		},
		{
			Description: "Simple query with basic filter (duplicate ID by docIDs arg), partial results",
			Request: `query {
						Users(docIDs: ["bae-52b9170d-b77a-5887-b877-cbdbb99b009f", "bae-52b9170d-b77a-5887-b877-cbdbb99b009f"]) {
							Name
							Age
						}
					}`,
			Docs: map[int][]string{
				0: {
					`{
						"Name": "John",
						"Age": 21
					}`,
					`{
						"Name": "Bob",
						"Age": 32
					}`,
				},
			},
			Results: []map[string]any{
				{
					"Name": "John",
					"Age":  int64(21),
				},
			},
		},
		{
			Description: "Simple query with basic filter (multiple ID by docIDs arg), partial results",
			Request: `query {
						Users(docIDs: ["bae-52b9170d-b77a-5887-b877-cbdbb99b009f", "bae-1378ab62-e064-5af4-9ea6-49941c8d8f94"]) {
							Name
							Age
						}
					}`,
			Docs: map[int][]string{
				0: {
					`{
						"Name": "John",
						"Age": 21
					}`,
					`{
						"Name": "Bob",
						"Age": 32
					}`,
					`{
						"Name": "Jim",
						"Age": 27
					}`,
				},
			},
			Results: []map[string]any{
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
	}

	for _, test := range tests {
		executeTestCase(t, test)
	}
}

func TestQuerySimpleReturnsNothinGivenEmptyDocIDsFilter(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with empty docIDs arg",
		Request: `query {
					Users(docIDs: []) {
						Name
						Age
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"Age": 21
				}`,
			},
		},
		Results: []map[string]any{},
	}

	executeTestCase(t, test)
}

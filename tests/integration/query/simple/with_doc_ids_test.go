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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryWithDocIDsFilter_SingleTargetNotFound(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
						"Name": "John",
						"Age": 21
					}`,
			},
			&action.Request{
				Request: `query {
						Users(docID: ["bae-52b9170d-b77a-5887-b877-cbdbb99b009g"]) {
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

func TestQueryWithDocIDsFilter_SingleTargetFound(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
						"Name": "John",
						"Age": 21
					}`,
			},
			&action.Request{
				Request: `query {
						Users(docID: ["bae-619ea0d2-35ba-5e8c-ac4d-2b769937213b"]) {
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
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithDocIDsFilter_OneFoundFromMultipleTargets(t *testing.T) {
	test := testUtils.TestCase{
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
			&action.Request{
				Request: `query {
						Users(docID: ["bae-619ea0d2-35ba-5e8c-ac4d-2b769937213b", "bae-619ea0d2-35ba-5e8c-ac4d-2b769937213b"]) {
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
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithDocIDsFilter_AllFoundFromMultipleTargets(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
						"Name": "Jim",
						"Age": 27
					}`,
			},
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
			&action.Request{
				Request: `query {
						Users(docID: ["bae-619ea0d2-35ba-5e8c-ac4d-2b769937213b", "bae-0000ef46-9bf6-5a83-9bbf-da288687c830"]) {
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
				NonOrderedResults: true,
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleReturnsNothinGivenEmptyDocIDsFilter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			&action.Request{
				Request: `query {
					Users(docID: []) {
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

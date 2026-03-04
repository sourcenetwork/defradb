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

package simple

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryWithDocIDsFilter_SingleTargetNotFound(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
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
			&action.AddDoc{
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
			&action.AddDoc{
				Doc: `{
						"Name": "John",
						"Age": 21
					}`,
			},
			&action.AddDoc{
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
			&action.AddDoc{
				Doc: `{
						"Name": "Jim",
						"Age": 27
					}`,
			},
			&action.AddDoc{
				Doc: `{
						"Name": "John",
						"Age": 21
					}`,
			},
			&action.AddDoc{
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
			&action.AddDoc{
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

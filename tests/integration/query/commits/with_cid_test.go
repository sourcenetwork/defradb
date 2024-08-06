// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package commits

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryCommitsWithCid(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with cid",
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: `{
					"age":	22
				}`,
			},
			testUtils.Request{
				Request: `query {
						commits(
							cid: "bafyreihv7jqe32wsuff5vwzlp7izoo6pqg6kgqf5edknp3mqm3344gu35q"
						) {
							cid
						}
					}`,
				Results: map[string]any{
					"commits": []map[string]any{
						{
							"cid": "bafyreihv7jqe32wsuff5vwzlp7izoo6pqg6kgqf5edknp3mqm3344gu35q",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithCidForFieldCommit(t *testing.T) {
	// cid is for a field commit, see TestQueryCommitsWithDocIDAndFieldId
	test := testUtils.TestCase{
		Description: "Simple all commits query with cid",
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			testUtils.Request{
				Request: `query {
						commits(
							cid: "bafyreihv7jqe32wsuff5vwzlp7izoo6pqg6kgqf5edknp3mqm3344gu35q"
						) {
							cid
						}
					}`,
				Results: map[string]any{
					"commits": []map[string]any{
						{
							"cid": "bafyreihv7jqe32wsuff5vwzlp7izoo6pqg6kgqf5edknp3mqm3344gu35q",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithInvalidCid(t *testing.T) {
	test := testUtils.TestCase{
		Description: "query for a single block by invalid CID",
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			testUtils.Request{
				Request: `query {
						commits(cid: "fhbnjfahfhfhanfhga") {
							cid
							height
							delta
						}
					}`,
				Results: map[string]any{
					"commits": []map[string]any{},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithInvalidShortCid(t *testing.T) {
	test := testUtils.TestCase{
		Description: "query for a single block by invalid, short CID",
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			testUtils.Request{
				Request: `query {
						commits(cid: "bafybeidfhbnjfahfhfhanfhga") {
							cid
							height
							delta
						}
					}`,
				Results: map[string]any{
					"commits": []map[string]any{},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithUnknownCid(t *testing.T) {
	test := testUtils.TestCase{
		Description: "query for a single block by unknown CID",
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			testUtils.Request{
				Request: `query {
						commits(cid: "bafybeid57gpbwi4i6bg7g35hhhhhhhhhhhhhhhhhhhhhhhdoesnotexist") {
							cid
							height
							delta
						}
					}`,
				Results: map[string]any{
					"commits": []map[string]any{},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

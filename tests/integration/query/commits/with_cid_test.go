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
							cid: "bafybeigvpf62j7j2wbpid5iavzxielbhbsbbirmgzqkw3wpptdvysuztwi"
						) {
							cid
						}
					}`,
				Results: []map[string]any{
					{
						"cid": "bafybeigvpf62j7j2wbpid5iavzxielbhbsbbirmgzqkw3wpptdvysuztwi",
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
							cid: "bafybeigvpf62j7j2wbpid5iavzxielbhbsbbirmgzqkw3wpptdvysuztwi"
						) {
							cid
						}
					}`,
				Results: []map[string]any{
					{
						"cid": "bafybeigvpf62j7j2wbpid5iavzxielbhbsbbirmgzqkw3wpptdvysuztwi",
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
				Results: []map[string]any{},
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
				Results: []map[string]any{},
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
				Results: []map[string]any{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

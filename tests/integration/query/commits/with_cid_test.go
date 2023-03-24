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
						"Name":	"John",
						"Age":	21
					}`,
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: `{
					"Age":	22
				}`,
			},
			testUtils.Request{
				Request: `query {
						commits(
							cid: "bafybeih7athlviqvv255ujgk6mikxcwxp46n7yyaekxldrgsvgrkrgajtq"
						) {
							cid
						}
					}`,
				Results: []map[string]any{
					{
						"cid": "bafybeih7athlviqvv255ujgk6mikxcwxp46n7yyaekxldrgsvgrkrgajtq",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"users"}, test)
}

func TestQueryCommitsWithCidForFieldCommit(t *testing.T) {
	// cid is for a field commit, see TestQueryCommitsWithDockeyAndFieldId
	test := testUtils.TestCase{
		Description: "Simple all commits query with cid",
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"Name":	"John",
						"Age":	21
					}`,
			},
			testUtils.Request{
				Request: `query {
						commits(
							cid: "bafybeih7athlviqvv255ujgk6mikxcwxp46n7yyaekxldrgsvgrkrgajtq"
						) {
							cid
						}
					}`,
				Results: []map[string]any{
					{
						"cid": "bafybeih7athlviqvv255ujgk6mikxcwxp46n7yyaekxldrgsvgrkrgajtq",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"users"}, test)
}

func TestQueryCommitsWithInvalidCid(t *testing.T) {
	test := testUtils.TestCase{
		Description: "query for a single block by invalid CID",
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"Name":	"John",
						"Age":	21
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

	testUtils.ExecuteTestCase(t, []string{"users"}, test)
}

func TestQueryCommitsWithInvalidShortCid(t *testing.T) {
	test := testUtils.TestCase{
		Description: "query for a single block by invalid, short CID",
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"Name":	"John",
						"Age":	21
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

	testUtils.ExecuteTestCase(t, []string{"users"}, test)
}

func TestQueryCommitsWithUnknownCid(t *testing.T) {
	test := testUtils.TestCase{
		Description: "query for a single block by unknown CID",
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"Name":	"John",
						"Age":	21
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

	testUtils.ExecuteTestCase(t, []string{"users"}, test)
}

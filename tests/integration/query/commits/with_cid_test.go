// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package all_commits

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryAllCommitsWithCid(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple all commits query with cid",
		Query: `query {
					allCommits(
						cid: "bafybeibrbfg35mwggcj4vnskak4qn45hp7fy5a4zp2n34sbq5vt5utr6pq"
					) {
						cid
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
		Updates: map[int]map[int][]string{
			0: {
				0: {
					`{
						"Age": 22
					}`,
				},
			},
		},
		Results: []map[string]any{
			{
				"cid": "bafybeibrbfg35mwggcj4vnskak4qn45hp7fy5a4zp2n34sbq5vt5utr6pq",
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryAllCommitsWithCidForFieldCommit(t *testing.T) {
	// cid is for a field commit, see TestQueryAllCommitsWithDockeyAndFieldId
	test := testUtils.QueryTestCase{
		Description: "Simple all commits query with cid",
		Query: `query {
					allCommits(
						cid: "bafybeidst2mzxhdoh4ayjdjoh4vibo7vwnuoxk3xgyk5mzmep55jklni2a"
					) {
						cid
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
				"cid": "bafybeidst2mzxhdoh4ayjdjoh4vibo7vwnuoxk3xgyk5mzmep55jklni2a",
			},
		},
	}

	executeTestCase(t, test)
}

// This test is for documentation reasons only. This is not
// desired behaviour (error message could be better, or empty result).
func TestQueryAllCommitsWithInvalidCid(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "query for a single block by invalid CID",
		Query: `query {
					allCommits(cid: "fhbnjfahfhfhanfhga") {
						cid
						height
						delta
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
		ExpectedError: "encoding/hex: invalid byte:",
	}

	executeTestCase(t, test)
}

// This test is for documentation reasons only. This is not
// desired behaviour (error message could be better, or empty result).
func TestQueryAllCommitsWithInvalidShortCid(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "query for a single block by invalid, short CID",
		Query: `query {
					allCommits(cid: "bafybeidfhbnjfahfhfhanfhga") {
						cid
						height
						delta
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
		ExpectedError: "length greater than remaining number of bytes in buffer",
	}

	executeTestCase(t, test)
}

func TestQueryAllCommitsWithUnknownCid(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "query for a single block by unknown CID",
		Query: `query {
					allCommits(cid: "bafybeid57gpbwi4i6bg7g35hhhhhhhhhhhhhhhhhhhhhhhdoesnotexist") {
						cid
						height
						delta
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

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
	test := testUtils.QueryTestCase{
		Description: "Simple all commits query with cid",
		Query: `query {
					commits(
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

func TestQueryCommitsWithCidForFieldCommit(t *testing.T) {
	// cid is for a field commit, see TestQueryCommitsWithDockeyAndFieldId
	test := testUtils.QueryTestCase{
		Description: "Simple all commits query with cid",
		Query: `query {
					commits(
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

func TestQueryCommitsWithInvalidCid(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "query for a single block by invalid CID",
		Query: `query {
					commits(cid: "fhbnjfahfhfhanfhga") {
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

func TestQueryCommitsWithInvalidShortCid(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "query for a single block by invalid, short CID",
		Query: `query {
					commits(cid: "bafybeidfhbnjfahfhfhanfhga") {
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

func TestQueryCommitsWithUnknownCid(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "query for a single block by unknown CID",
		Query: `query {
					commits(cid: "bafybeid57gpbwi4i6bg7g35hhhhhhhhhhhhhhhhhhhhhhhdoesnotexist") {
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

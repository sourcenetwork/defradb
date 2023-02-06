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
	test := testUtils.RequestTestCase{
		Description: "Simple all commits query with cid",
		Request: `query {
					commits(
						cid: "bafybeignirnk6wuxtg2fzwfbvs26wmrldlhpqj243kwnxb4ewafbae23m4"
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
				"cid": "bafybeignirnk6wuxtg2fzwfbvs26wmrldlhpqj243kwnxb4ewafbae23m4",
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryCommitsWithCidForFieldCommit(t *testing.T) {
	// cid is for a field commit, see TestQueryCommitsWithDockeyAndFieldId
	test := testUtils.RequestTestCase{
		Description: "Simple all commits query with cid",
		Request: `query {
					commits(
						cid: "bafybeicovjpmtwu544e7hzgg7mcwabstmugesi3n62ju6kbimcsjqp23gu"
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
				"cid": "bafybeicovjpmtwu544e7hzgg7mcwabstmugesi3n62ju6kbimcsjqp23gu",
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryCommitsWithInvalidCid(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "query for a single block by invalid CID",
		Request: `query {
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
	test := testUtils.RequestTestCase{
		Description: "query for a single block by invalid, short CID",
		Request: `query {
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
	test := testUtils.RequestTestCase{
		Description: "query for a single block by unknown CID",
		Request: `query {
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

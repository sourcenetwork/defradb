// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package commit

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// This test is for documentation reasons only. This is not
// desired behaviour (error message could be better, or empty result).
func TestQueryOneCommitWithInvalidCid(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "query for a single block by invalid CID",
		Query: `query {
					commit(cid: "fhbnjfahfhfhanfhga") {
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
func TestQueryOneCommitWithInvalidShortCid(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "query for a single block by invalid, short CID",
		Query: `query {
					commit(cid: "bafybeidfhbnjfahfhfhanfhga") {
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

// This test is for documentation reasons only. This is not
// desired behaviour (should be empty result).
func TestQueryOneCommitWithUnknownCid(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "query for a single block by unknown CID",
		Query: `query {
					commit(cid: "bafybeid57gpbwi4i6bg7g35hhhhhhhhhhhhhhhhhhhhhhhdoesnotexist") {
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
		ExpectedError: "ipld: could not find",
	}

	executeTestCase(t, test)
}

func TestQueryOneCommitWithCid(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "query for a single block by CID",
		Query: `query {
					commit(cid: "bafybeid57gpbwi4i6bg7g357vwwyzsmr4bjo22rmhoxrwqvdxlqxcgaqvu") {
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
		Results: []map[string]any{
			{
				"cid":    "bafybeid57gpbwi4i6bg7g357vwwyzsmr4bjo22rmhoxrwqvdxlqxcgaqvu",
				"height": int64(1),
				// cbor encoded delta
				"delta": []uint8{
					0xa2,
					0x63,
					0x41,
					0x67,
					0x65,
					0x15,
					0x64,
					0x4e,
					0x61,
					0x6d,
					0x65,
					0x64,
					0x4a,
					0x6f,
					0x68,
					0x6e,
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryOneCommitWithCidAndLinks(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "query for a single block by CID, with links",
		Query: `query {
					commit(cid: "bafybeid57gpbwi4i6bg7g357vwwyzsmr4bjo22rmhoxrwqvdxlqxcgaqvu") {
						cid
						height
						delta
						links {
							cid
							name
						}
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
				"cid":    "bafybeid57gpbwi4i6bg7g357vwwyzsmr4bjo22rmhoxrwqvdxlqxcgaqvu",
				"height": int64(1),
				// cbor encoded delta
				"delta": []uint8{
					0xa2,
					0x63,
					0x41,
					0x67,
					0x65,
					0x15,
					0x64,
					0x4e,
					0x61,
					0x6d,
					0x65,
					0x64,
					0x4a,
					0x6f,
					0x68,
					0x6e,
				},
				"links": []map[string]any{
					{
						"cid":  "bafybeidst2mzxhdoh4ayjdjoh4vibo7vwnuoxk3xgyk5mzmep55jklni2a",
						"name": "Age",
					},
					{
						"cid":  "bafybeihhypcsqt7blkrqtcmpl43eo3yunrog5pchox5naji6hisdme4swm",
						"name": "Name",
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

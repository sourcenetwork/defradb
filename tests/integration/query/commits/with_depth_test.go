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

func TestQueryCommitsWithDepth1(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple all commits query with depth 1",
		Query: `query {
					commits(depth: 1) {
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
			{
				"cid": "bafybeihhypcsqt7blkrqtcmpl43eo3yunrog5pchox5naji6hisdme4swm",
			},
			{
				"cid": "bafybeid57gpbwi4i6bg7g357vwwyzsmr4bjo22rmhoxrwqvdxlqxcgaqvu",
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryCommitsWithDepth1WithUpdate(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple all commits query with depth 1, and doc updates",
		Query: `query {
					commits(depth: 1) {
						cid
						height
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
				"cid":    "bafybeicvef4ugls2dl7j4hibt2ahxss2i2i4bbgps7tkjiaoybp6q73mca",
				"height": int64(2),
			},
			{
				// "Name" field head (unchanged from create)
				"cid":    "bafybeihhypcsqt7blkrqtcmpl43eo3yunrog5pchox5naji6hisdme4swm",
				"height": int64(1),
			},
			{
				// "Age" field head
				"cid":    "bafybeibrbfg35mwggcj4vnskak4qn45hp7fy5a4zp2n34sbq5vt5utr6pq",
				"height": int64(2),
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryCommitsWithDepth2WithUpdate(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple all commits query with depth 2, and doc updates",
		Query: `query {
					commits(depth: 2) {
						cid
						height
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
					`{
						"Age": 23
					}`,
				},
			},
		},
		Results: []map[string]any{
			{
				// Composite head
				"cid":    "bafybeiaxjhz6dna7fyf7tqo5hooilwvaezswd5xfsmb2lfgcy7tpzklikm",
				"height": int64(3),
			},
			{
				// Composite head -1
				"cid":    "bafybeicvef4ugls2dl7j4hibt2ahxss2i2i4bbgps7tkjiaoybp6q73mca",
				"height": int64(2),
			},
			{
				// "Name" field head (unchanged from create)
				"cid":    "bafybeihhypcsqt7blkrqtcmpl43eo3yunrog5pchox5naji6hisdme4swm",
				"height": int64(1),
			},
			{
				// "Age" field head
				"cid":    "bafybeid2tudsm4go5boq7yvz6pprtgaiddkazq2dip6c4fsqt3afhzexbq",
				"height": int64(3),
			},
			{
				// "Age" field head -1
				"cid":    "bafybeibrbfg35mwggcj4vnskak4qn45hp7fy5a4zp2n34sbq5vt5utr6pq",
				"height": int64(2),
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryCommitsWithDepth1AndMultipleDocs(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple all commits query with depth 1",
		Query: `query {
					commits(depth: 1) {
						cid
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"Age": 21
				}`,
				`{
					"Name": "Fred",
					"Age": 25
				}`,
			},
		},
		Results: []map[string]any{
			{
				"cid": "bafybeidst2mzxhdoh4ayjdjoh4vibo7vwnuoxk3xgyk5mzmep55jklni2a",
			},
			{
				"cid": "bafybeihhypcsqt7blkrqtcmpl43eo3yunrog5pchox5naji6hisdme4swm",
			},
			{
				"cid": "bafybeid57gpbwi4i6bg7g357vwwyzsmr4bjo22rmhoxrwqvdxlqxcgaqvu",
			},
			{
				"cid": "bafybeiajhlicqju3thdnyemvparx35kg6vfb6sr3vuemhw7zjrulx2tkom",
			},
			{
				"cid": "bafybeifl4q2htt4sozl5dnxjqkpstpqbpkurgqc56dnn2bvtsora3srl2q",
			},
			{
				"cid": "bafybeigbxr6aavljpcfruccoxi43kxjb46locbb7mwupwzjxbz5whplydu",
			},
		},
	}

	executeTestCase(t, test)
}

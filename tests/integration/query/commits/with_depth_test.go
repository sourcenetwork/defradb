// Copyright 2023 Democratized Data Foundation
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
	test := testUtils.RequestTestCase{
		Description: "Simple all commits query with depth 1",
		Request: `query {
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
				"cid": "bafybeigju7dgicfq3fxvtlxtjao7won4xc7kusykkvumngjfx5i2c7ibny",
			},
			{
				"cid": "bafybeiaqarrcayyoly2gdiam6mhh72ls4azwa7brozxxc3q2srnggkkqkq",
			},
			{
				"cid": "bafybeid5l577igkgcn6wjqjeqxlta4dcc3a3iykwkborf4fklaenjuctoq",
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryCommitsWithDepth1WithUpdate(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple all commits query with depth 1, and doc updates",
		Request: `query {
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
				"cid":    "bafybeiacqac6scm7pmtlvqptvtljmoroevnoedku42qi5bmfdpaelcu5fm",
				"height": int64(2),
			},
			{
				// "Name" field head (unchanged from create)
				"cid":    "bafybeiaqarrcayyoly2gdiam6mhh72ls4azwa7brozxxc3q2srnggkkqkq",
				"height": int64(1),
			},
			{
				// "Age" field head
				"cid":    "bafybeibz3vbkt75siz3zogke6tlzvpcxttpiy4xivjvgyeaorjz6wsbguq",
				"height": int64(2),
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryCommitsWithDepth2WithUpdate(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple all commits query with depth 2, and doc updates",
		Request: `query {
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
				"cid":    "bafybeiho5z6seahwxgbdyobylzyarrdschgzmood7rkdtp4qpd2uxebaxy",
				"height": int64(3),
			},
			{
				// Composite head -1
				"cid":    "bafybeiacqac6scm7pmtlvqptvtljmoroevnoedku42qi5bmfdpaelcu5fm",
				"height": int64(2),
			},
			{
				// "Name" field head (unchanged from create)
				"cid":    "bafybeiaqarrcayyoly2gdiam6mhh72ls4azwa7brozxxc3q2srnggkkqkq",
				"height": int64(1),
			},
			{
				// "Age" field head
				"cid":    "bafybeib6yxcmbg2gz5ss6d67u5mu6wcatfjtdp2rv44difyznp3rqlyu4m",
				"height": int64(3),
			},
			{
				// "Age" field head -1
				"cid":    "bafybeibz3vbkt75siz3zogke6tlzvpcxttpiy4xivjvgyeaorjz6wsbguq",
				"height": int64(2),
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryCommitsWithDepth1AndMultipleDocs(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple all commits query with depth 1",
		Request: `query {
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
				"cid": "bafybeigju7dgicfq3fxvtlxtjao7won4xc7kusykkvumngjfx5i2c7ibny",
			},
			{
				"cid": "bafybeiaqarrcayyoly2gdiam6mhh72ls4azwa7brozxxc3q2srnggkkqkq",
			},
			{
				"cid": "bafybeid5l577igkgcn6wjqjeqxlta4dcc3a3iykwkborf4fklaenjuctoq",
			},
			{
				"cid": "bafybeifzqn3n6unmfd4kabhxermcbrp564nu2ms3uh6m73i26b3zwvjrku",
			},
			{
				"cid": "bafybeieicf7so27bdrrarhwxi4wzzs5yyxku2wtea555gcrgz4kmpjgdvu",
			},
			{
				"cid": "bafybeig2efgh5jy5kbknnvxkgbtz66kn75ixr5pcrrreve6t46e7ba3l3y",
			},
		},
	}

	executeTestCase(t, test)
}

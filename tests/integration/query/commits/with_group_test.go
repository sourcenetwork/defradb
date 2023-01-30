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

func TestQueryCommitsWithGroupBy(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple all commits query, group by height",
		Request: `query {
					commits(groupBy: [height]) {
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
				"height": int64(2),
			},
			{
				"height": int64(1),
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryCommitsWithGroupByHeightWithChild(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple all commits query, group by height",
		Request: `query {
					commits(groupBy: [height]) {
						height
						_group {
							cid
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
				"height": int64(2),
				"_group": []map[string]any{
					{
						"cid": "bafybeiacqac6scm7pmtlvqptvtljmoroevnoedku42qi5bmfdpaelcu5fm",
					},
					{
						"cid": "bafybeibz3vbkt75siz3zogke6tlzvpcxttpiy4xivjvgyeaorjz6wsbguq",
					},
				},
			},
			{
				"height": int64(1),
				"_group": []map[string]any{
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
			},
		},
	}

	executeTestCase(t, test)
}

// This is an odd test, but we need to make sure it works
func TestQueryCommitsWithGroupByCidWithChild(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple all commits query, group by cid",
		Request: `query {
					commits(groupBy: [cid]) {
						cid
						_group {
							height
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
				"cid": "bafybeigju7dgicfq3fxvtlxtjao7won4xc7kusykkvumngjfx5i2c7ibny",
				"_group": []map[string]any{
					{
						"height": int64(1),
					},
				},
			},
			{
				"cid": "bafybeiaqarrcayyoly2gdiam6mhh72ls4azwa7brozxxc3q2srnggkkqkq",
				"_group": []map[string]any{
					{
						"height": int64(1),
					},
				},
			},
			{
				"cid": "bafybeid5l577igkgcn6wjqjeqxlta4dcc3a3iykwkborf4fklaenjuctoq",
				"_group": []map[string]any{
					{
						"height": int64(1),
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

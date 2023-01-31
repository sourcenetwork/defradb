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

func TestQueryCommitsWithDockeyAndOrderHeightDesc(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple all commits query with dockey, order height desc",
		Request: `query {
					commits(dockey: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f", order: {height: DESC}) {
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
				"cid":    "bafybeibz3vbkt75siz3zogke6tlzvpcxttpiy4xivjvgyeaorjz6wsbguq",
				"height": int64(2),
			},
			{
				"cid":    "bafybeiacqac6scm7pmtlvqptvtljmoroevnoedku42qi5bmfdpaelcu5fm",
				"height": int64(2),
			},
			{
				"cid":    "bafybeid5l577igkgcn6wjqjeqxlta4dcc3a3iykwkborf4fklaenjuctoq",
				"height": int64(1),
			},
			{
				"cid":    "bafybeiaqarrcayyoly2gdiam6mhh72ls4azwa7brozxxc3q2srnggkkqkq",
				"height": int64(1),
			},
			{
				"cid":    "bafybeigju7dgicfq3fxvtlxtjao7won4xc7kusykkvumngjfx5i2c7ibny",
				"height": int64(1),
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryCommitsWithDockeyAndOrderHeightAsc(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple all commits query with dockey, order height asc",
		Request: `query {
					commits(dockey: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f", order: {height: ASC}) {
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
				"cid":    "bafybeid5l577igkgcn6wjqjeqxlta4dcc3a3iykwkborf4fklaenjuctoq",
				"height": int64(1),
			},
			{
				"cid":    "bafybeiaqarrcayyoly2gdiam6mhh72ls4azwa7brozxxc3q2srnggkkqkq",
				"height": int64(1),
			},
			{
				"cid":    "bafybeigju7dgicfq3fxvtlxtjao7won4xc7kusykkvumngjfx5i2c7ibny",
				"height": int64(1),
			},
			{
				"cid":    "bafybeibz3vbkt75siz3zogke6tlzvpcxttpiy4xivjvgyeaorjz6wsbguq",
				"height": int64(2),
			},
			{
				"cid":    "bafybeiacqac6scm7pmtlvqptvtljmoroevnoedku42qi5bmfdpaelcu5fm",
				"height": int64(2),
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryCommitsWithDockeyAndOrderCidDesc(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple all commits query with dockey, order cid desc",
		Request: `query {
					commits(dockey: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f", order: {cid: DESC}) {
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
				"cid":    "bafybeigju7dgicfq3fxvtlxtjao7won4xc7kusykkvumngjfx5i2c7ibny",
				"height": int64(1),
			},
			{
				"cid":    "bafybeid5l577igkgcn6wjqjeqxlta4dcc3a3iykwkborf4fklaenjuctoq",
				"height": int64(1),
			},
			{
				"cid":    "bafybeibz3vbkt75siz3zogke6tlzvpcxttpiy4xivjvgyeaorjz6wsbguq",
				"height": int64(2),
			},
			{
				"cid":    "bafybeiaqarrcayyoly2gdiam6mhh72ls4azwa7brozxxc3q2srnggkkqkq",
				"height": int64(1),
			},
			{
				"cid":    "bafybeiacqac6scm7pmtlvqptvtljmoroevnoedku42qi5bmfdpaelcu5fm",
				"height": int64(2),
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryCommitsWithDockeyAndOrderCidAsc(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple all commits query with dockey, order cid asc",
		Request: `query {
					commits(dockey: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f", order: {cid: ASC}) {
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
				"cid":    "bafybeiaqarrcayyoly2gdiam6mhh72ls4azwa7brozxxc3q2srnggkkqkq",
				"height": int64(1),
			},
			{
				"cid":    "bafybeibz3vbkt75siz3zogke6tlzvpcxttpiy4xivjvgyeaorjz6wsbguq",
				"height": int64(2),
			},
			{
				"cid":    "bafybeid5l577igkgcn6wjqjeqxlta4dcc3a3iykwkborf4fklaenjuctoq",
				"height": int64(1),
			},
			{
				"cid":    "bafybeigju7dgicfq3fxvtlxtjao7won4xc7kusykkvumngjfx5i2c7ibny",
				"height": int64(1),
			},
		},
	}

	executeTestCase(t, test)
}

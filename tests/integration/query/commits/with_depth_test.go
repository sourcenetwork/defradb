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
	test := testUtils.TestCase{
		Description: "Simple all commits query with depth 1",
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
						commits(depth: 1) {
							cid
						}
					}`,
				Results: []map[string]any{
					{
						"cid": "bafybeihfw5lufgs7ygv45to5rqvt3xkecjgikoccjyx6y2i7lnaclmrcjm",
					},
					{
						"cid": "bafybeigmez6gtszsqx6aevzlanvpazhhezw5va4wizhqtqz5k4s2dqjb24",
					},
					{
						"cid": "bafybeiedu23doqe2nagdbmkvfyuouajnfxo7ezy57vbv34dqewhwbfg45u",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithDepth1WithUpdate(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with depth 1, and doc updates",
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
						commits(depth: 1) {
							cid
							height
						}
					}`,
				Results: []map[string]any{
					{
						// "Age" field head
						"cid":    "bafybeihqgrwnhc4w7e5cbhycxvqrpzgi2ei4xrcsre2plceclptgn4tc3i",
						"height": int64(2),
					},
					{
						// "Name" field head (unchanged from create)
						"cid":    "bafybeigmez6gtszsqx6aevzlanvpazhhezw5va4wizhqtqz5k4s2dqjb24",
						"height": int64(1),
					},
					{
						"cid":    "bafybeibfwqf5szatmlyl3alru4nq3gnxaiyyb3ggqung2jwb4qnm6mejyu",
						"height": int64(2),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithDepth2WithUpdate(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with depth 2, and doc updates",
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
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: `{
					"age":	23
				}`,
			},
			testUtils.Request{
				Request: `query {
						commits(depth: 2) {
							cid
							height
						}
					}`,
				Results: []map[string]any{
					{
						// Composite head
						"cid":    "bafybeihvhr7ke7bjgjixce262544tlo7mdlyuswtgl66zsrxcfc5targjy",
						"height": int64(3),
					},
					{
						// Composite head -1
						"cid":    "bafybeihqgrwnhc4w7e5cbhycxvqrpzgi2ei4xrcsre2plceclptgn4tc3i",
						"height": int64(2),
					},
					{
						// "Name" field head (unchanged from create)
						"cid":    "bafybeigmez6gtszsqx6aevzlanvpazhhezw5va4wizhqtqz5k4s2dqjb24",
						"height": int64(1),
					},
					{
						// "Age" field head
						"cid":    "bafybeicacrvck5qf37pk3pdsiavvxy2jk67dbdpww5pvoun2k52lw2ftqi",
						"height": int64(3),
					},
					{
						// "Age" field head -1
						"cid":    "bafybeibfwqf5szatmlyl3alru4nq3gnxaiyyb3ggqung2jwb4qnm6mejyu",
						"height": int64(2),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithDepth1AndMultipleDocs(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with depth 1",
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"Fred",
						"age":	25
					}`,
			},
			testUtils.Request{
				Request: `query {
						commits(depth: 1) {
							cid
						}
					}`,
				Results: []map[string]any{
					{
						"cid": "bafybeigvcksw7ck2o7rqfyxncn2h5u6bbwj5ejjfvsihsjibxvrqrxbtui",
					},
					{
						"cid": "bafybeibcdmghhshx4v3xamoktw3n6blv7courh6x2d5cttwuzlodml74ny",
					},
					{
						"cid": "bafybeig6rwkq6hlf5rcjq64jodl3gtfv5svnmsjlkwrnmbcjui7t3vy3qi",
					},
					{
						"cid": "bafybeihfw5lufgs7ygv45to5rqvt3xkecjgikoccjyx6y2i7lnaclmrcjm",
					},
					{
						"cid": "bafybeigmez6gtszsqx6aevzlanvpazhhezw5va4wizhqtqz5k4s2dqjb24",
					},
					{
						"cid": "bafybeiedu23doqe2nagdbmkvfyuouajnfxo7ezy57vbv34dqewhwbfg45u",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

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
						"cid": "bafybeiagew5ervkxkyz6sezgkrcrhv6towawpbirg6ziouah7lj5eg7lwy",
					},
					{
						"cid": "bafybeibsfjpv57libozcmiihs2ixgsoseu752ni6nkeqqv2nhgxgpe7l5m",
					},
					{
						"cid": "bafybeiaitgizimh6ttm3stfabpwqvg5tvty3j46p6hqbat4w7ippcbw4ky",
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
						"cid":    "bafybeiblr52xw3xplsat66bwwyfg4pgbh3gei7njmdvomodf2qygt627r4",
						"height": int64(2),
					},
					{
						// "Name" field head (unchanged from create)
						"cid":    "bafybeibsfjpv57libozcmiihs2ixgsoseu752ni6nkeqqv2nhgxgpe7l5m",
						"height": int64(1),
					},
					{
						"cid":    "bafybeihu5onq2trqv25zieyejrnj6kyh2533i3rchvq64kahbjz7uzrquu",
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
						"cid":    "bafybeigzqhl3ckoduvhl7slck4mlgqo6lif5w2nedjlwz7rq3pacodc6t4",
						"height": int64(3),
					},
					{
						// Composite head -1
						"cid":    "bafybeiblr52xw3xplsat66bwwyfg4pgbh3gei7njmdvomodf2qygt627r4",
						"height": int64(2),
					},
					{
						// "Name" field head (unchanged from create)
						"cid":    "bafybeibsfjpv57libozcmiihs2ixgsoseu752ni6nkeqqv2nhgxgpe7l5m",
						"height": int64(1),
					},
					{
						// "Age" field head
						"cid":    "bafybeibbigbsvqdf2smkd33c34psvcljxfkexvbpismeumiy7r3mkuhtai",
						"height": int64(3),
					},
					{
						// "Age" field head -1
						"cid":    "bafybeihu5onq2trqv25zieyejrnj6kyh2533i3rchvq64kahbjz7uzrquu",
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
						"cid": "bafybeihzfcx7als6w6qxy76lcxdxs6yepdxt7fmhcqmkgm3s2vjjokcdaq",
					},
					{
						"cid": "bafybeiffaln23wx47mcfhvzgiuuyqfjto2r32hjjrd2kq5hzce72vwdixy",
					},
					{
						"cid": "bafybeie3y57zvbpw4pnw7okbj55uneh42ihspp4vjqsqxvr6if4xgwffqq",
					},
					{
						"cid": "bafybeiagew5ervkxkyz6sezgkrcrhv6towawpbirg6ziouah7lj5eg7lwy",
					},
					{
						"cid": "bafybeibsfjpv57libozcmiihs2ixgsoseu752ni6nkeqqv2nhgxgpe7l5m",
					},
					{
						"cid": "bafybeiaitgizimh6ttm3stfabpwqvg5tvty3j46p6hqbat4w7ippcbw4ky",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

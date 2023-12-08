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

func TestQueryCommitsWithDocIDAndOrderHeightDesc(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with docID, order height desc",
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
						commits(docID: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7", order: {height: DESC}) {
							cid
							height
						}
					}`,
				Results: []map[string]any{
					{
						"cid":    "bafybeiblr52xw3xplsat66bwwyfg4pgbh3gei7njmdvomodf2qygt627r4",
						"height": int64(2),
					},
					{
						"cid":    "bafybeihu5onq2trqv25zieyejrnj6kyh2533i3rchvq64kahbjz7uzrquu",
						"height": int64(2),
					},
					{
						"cid":    "bafybeiagew5ervkxkyz6sezgkrcrhv6towawpbirg6ziouah7lj5eg7lwy",
						"height": int64(1),
					},
					{
						"cid":    "bafybeibsfjpv57libozcmiihs2ixgsoseu752ni6nkeqqv2nhgxgpe7l5m",
						"height": int64(1),
					},
					{
						"cid":    "bafybeiaitgizimh6ttm3stfabpwqvg5tvty3j46p6hqbat4w7ippcbw4ky",
						"height": int64(1),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithDocIDAndOrderHeightAsc(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with docID, order height asc",
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
						commits(docID: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7", order: {height: ASC}) {
							cid
							height
						}
					}`,
				Results: []map[string]any{
					{
						"cid":    "bafybeiagew5ervkxkyz6sezgkrcrhv6towawpbirg6ziouah7lj5eg7lwy",
						"height": int64(1),
					},
					{
						"cid":    "bafybeibsfjpv57libozcmiihs2ixgsoseu752ni6nkeqqv2nhgxgpe7l5m",
						"height": int64(1),
					},
					{
						"cid":    "bafybeiaitgizimh6ttm3stfabpwqvg5tvty3j46p6hqbat4w7ippcbw4ky",
						"height": int64(1),
					},
					{
						"cid":    "bafybeiblr52xw3xplsat66bwwyfg4pgbh3gei7njmdvomodf2qygt627r4",
						"height": int64(2),
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

func TestQueryCommitsWithDocIDAndOrderCidDesc(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with docID, order cid desc",
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
						commits(docID: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7", order: {cid: DESC}) {
							cid
							height
						}
					}`,
				Results: []map[string]any{
					{
						"cid":    "bafybeihu5onq2trqv25zieyejrnj6kyh2533i3rchvq64kahbjz7uzrquu",
						"height": int64(2),
					},
					{
						"cid":    "bafybeibsfjpv57libozcmiihs2ixgsoseu752ni6nkeqqv2nhgxgpe7l5m",
						"height": int64(1),
					},
					{
						"cid":    "bafybeiblr52xw3xplsat66bwwyfg4pgbh3gei7njmdvomodf2qygt627r4",
						"height": int64(2),
					},
					{
						"cid":    "bafybeiaitgizimh6ttm3stfabpwqvg5tvty3j46p6hqbat4w7ippcbw4ky",
						"height": int64(1),
					},
					{
						"cid":    "bafybeiagew5ervkxkyz6sezgkrcrhv6towawpbirg6ziouah7lj5eg7lwy",
						"height": int64(1),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithDocIDAndOrderCidAsc(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with docID, order cid asc",
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
						commits(docID: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7", order: {cid: ASC}) {
							cid
							height
						}
					}`,
				Results: []map[string]any{
					{
						"cid":    "bafybeiagew5ervkxkyz6sezgkrcrhv6towawpbirg6ziouah7lj5eg7lwy",
						"height": int64(1),
					},
					{
						"cid":    "bafybeiaitgizimh6ttm3stfabpwqvg5tvty3j46p6hqbat4w7ippcbw4ky",
						"height": int64(1),
					},
					{
						"cid":    "bafybeiblr52xw3xplsat66bwwyfg4pgbh3gei7njmdvomodf2qygt627r4",
						"height": int64(2),
					},
					{
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

func TestQueryCommitsWithDocIDAndOrderAndMultiUpdatesCidAsc(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with docID, multiple updates with order cid asc",
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
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: `{
					"age":	24
				}`,
			},
			testUtils.Request{
				Request: `query {
						 commits(docID: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7", order: {height: ASC}) {
							 cid
							 height
						 }
					 }`,
				Results: []map[string]any{
					{
						"cid":    "bafybeiagew5ervkxkyz6sezgkrcrhv6towawpbirg6ziouah7lj5eg7lwy",
						"height": int64(1),
					},
					{
						"cid":    "bafybeibsfjpv57libozcmiihs2ixgsoseu752ni6nkeqqv2nhgxgpe7l5m",
						"height": int64(1),
					},
					{
						"cid":    "bafybeiaitgizimh6ttm3stfabpwqvg5tvty3j46p6hqbat4w7ippcbw4ky",
						"height": int64(1),
					},
					{
						"cid":    "bafybeiblr52xw3xplsat66bwwyfg4pgbh3gei7njmdvomodf2qygt627r4",
						"height": int64(2),
					},
					{
						"cid":    "bafybeihu5onq2trqv25zieyejrnj6kyh2533i3rchvq64kahbjz7uzrquu",
						"height": int64(2),
					},
					{
						"cid":    "bafybeigzqhl3ckoduvhl7slck4mlgqo6lif5w2nedjlwz7rq3pacodc6t4",
						"height": int64(3),
					},
					{
						"cid":    "bafybeibbigbsvqdf2smkd33c34psvcljxfkexvbpismeumiy7r3mkuhtai",
						"height": int64(3),
					},
					{
						"cid":    "bafybeifxmvyhbhn5r7bndk3ihddi3f57c5ud46lhdsdiiampczkp63wrt4",
						"height": int64(4),
					},
					{
						"cid":    "bafybeigraptlxmhowk7we2mx552baukqlyijlovj6oqqczf5jb473vjbla",
						"height": int64(4),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

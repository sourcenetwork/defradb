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

func TestQueryCommitsWithUnknownDocID(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with unknown document ID",
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
						commits(docID: "unknown document ID") {
							cid
						}
					}`,
				Results: []map[string]any{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithDocID(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with docID",
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
						commits(docID: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7") {
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

func TestQueryCommitsWithDocIDAndLinks(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with docID, with links",
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
						commits(docID: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7") {
							cid
							links {
								cid
								name
							}
						}
					}`,
				Results: []map[string]any{
					{
						"cid":   "bafybeiagew5ervkxkyz6sezgkrcrhv6towawpbirg6ziouah7lj5eg7lwy",
						"links": []map[string]any{},
					},
					{
						"cid":   "bafybeibsfjpv57libozcmiihs2ixgsoseu752ni6nkeqqv2nhgxgpe7l5m",
						"links": []map[string]any{},
					},
					{
						"cid": "bafybeiaitgizimh6ttm3stfabpwqvg5tvty3j46p6hqbat4w7ippcbw4ky",
						"links": []map[string]any{
							{
								"cid":  "bafybeiagew5ervkxkyz6sezgkrcrhv6towawpbirg6ziouah7lj5eg7lwy",
								"name": "age",
							},
							{
								"cid":  "bafybeibsfjpv57libozcmiihs2ixgsoseu752ni6nkeqqv2nhgxgpe7l5m",
								"name": "name",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithDocIDAndUpdate(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with docID, multiple results",
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
						commits(docID: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7") {
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
						"cid":    "bafybeiagew5ervkxkyz6sezgkrcrhv6towawpbirg6ziouah7lj5eg7lwy",
						"height": int64(1),
					},
					{
						"cid":    "bafybeibsfjpv57libozcmiihs2ixgsoseu752ni6nkeqqv2nhgxgpe7l5m",
						"height": int64(1),
					},
					{
						"cid":    "bafybeihu5onq2trqv25zieyejrnj6kyh2533i3rchvq64kahbjz7uzrquu",
						"height": int64(2),
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

// This test is for documentation reasons only. This is not
// desired behaviour (first results includes link._head, second
// includes link._Name).
func TestQueryCommitsWithDocIDAndUpdateAndLinks(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with docID, multiple results and links",
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
						commits(docID: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7") {
							cid
							links {
								cid
								name
							}
						}
					}`,
				Results: []map[string]any{
					{
						"cid": "bafybeiblr52xw3xplsat66bwwyfg4pgbh3gei7njmdvomodf2qygt627r4",
						"links": []map[string]any{
							{
								"cid":  "bafybeiagew5ervkxkyz6sezgkrcrhv6towawpbirg6ziouah7lj5eg7lwy",
								"name": "_head",
							},
						},
					},
					{
						"cid":   "bafybeiagew5ervkxkyz6sezgkrcrhv6towawpbirg6ziouah7lj5eg7lwy",
						"links": []map[string]any{},
					},
					{
						"cid":   "bafybeibsfjpv57libozcmiihs2ixgsoseu752ni6nkeqqv2nhgxgpe7l5m",
						"links": []map[string]any{},
					},
					{
						"cid": "bafybeihu5onq2trqv25zieyejrnj6kyh2533i3rchvq64kahbjz7uzrquu",
						"links": []map[string]any{
							{
								"cid":  "bafybeiaitgizimh6ttm3stfabpwqvg5tvty3j46p6hqbat4w7ippcbw4ky",
								"name": "_head",
							},
							{
								"cid":  "bafybeiblr52xw3xplsat66bwwyfg4pgbh3gei7njmdvomodf2qygt627r4",
								"name": "age",
							},
						},
					},
					{
						"cid": "bafybeiaitgizimh6ttm3stfabpwqvg5tvty3j46p6hqbat4w7ippcbw4ky",
						"links": []map[string]any{
							{
								"cid":  "bafybeiagew5ervkxkyz6sezgkrcrhv6towawpbirg6ziouah7lj5eg7lwy",
								"name": "age",
							},
							{
								"cid":  "bafybeibsfjpv57libozcmiihs2ixgsoseu752ni6nkeqqv2nhgxgpe7l5m",
								"name": "name",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

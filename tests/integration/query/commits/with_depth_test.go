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
						"cid": "bafybeiazsz3twea2uxpen6452qqa7qnzp2xildfxliidhqk632jpvbixkm",
					},
					{
						"cid": "bafybeidzukbs36cwwhab4rkpi6jfhhxse2vjtc5tf767qda5valcinilmy",
					},
					{
						"cid": "bafybeihbcl2ijavd6vdcj4vgunw4q5qt5itmumxw7iy7fhoqfsuvkpkqeq",
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
						"cid":    "bafybeift3qzwhklfpkgszvrmbfzb6zp3g3cqryhjkuaoz3kp2yrj763jce",
						"height": int64(2),
					},
					{
						// "Name" field head (unchanged from create)
						"cid":    "bafybeidzukbs36cwwhab4rkpi6jfhhxse2vjtc5tf767qda5valcinilmy",
						"height": int64(1),
					},
					{
						"cid":    "bafybeige35bkafoez4cf4v6hgdkm5iaqcuqfq4bkyt7fxeycbdnqtbr7g4",
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
						"cid":    "bafybeietneua73vfrkvfefw5rmju7yhee6rywychdbls5xqmqtqmzfckzq",
						"height": int64(3),
					},
					{
						// Composite head -1
						"cid":    "bafybeift3qzwhklfpkgszvrmbfzb6zp3g3cqryhjkuaoz3kp2yrj763jce",
						"height": int64(2),
					},
					{
						// "Name" field head (unchanged from create)
						"cid":    "bafybeidzukbs36cwwhab4rkpi6jfhhxse2vjtc5tf767qda5valcinilmy",
						"height": int64(1),
					},
					{
						// "Age" field head
						"cid":    "bafybeidwgrk2xyu25pmwvpkfs4hnswtgej6gopkf26jrgm6lpbofa3rs3e",
						"height": int64(3),
					},
					{
						// "Age" field head -1
						"cid":    "bafybeige35bkafoez4cf4v6hgdkm5iaqcuqfq4bkyt7fxeycbdnqtbr7g4",
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
						"cid": "bafybeihayvvwwsjvd3yefenc4ubebriluyg4rdzxmizrhefk4agotcqlp4",
					},
					{
						"cid": "bafybeiezcqlaqvozdw3ogdf2dxukwrf5m3xydd7lyy6ylcqycx5uqqepfm",
					},
					{
						"cid": "bafybeicr2lalkqj6weqcafm32posw22hjmybwohau57eswg5a442qilc2q",
					},
					{
						"cid": "bafybeiazsz3twea2uxpen6452qqa7qnzp2xildfxliidhqk632jpvbixkm",
					},
					{
						"cid": "bafybeidzukbs36cwwhab4rkpi6jfhhxse2vjtc5tf767qda5valcinilmy",
					},
					{
						"cid": "bafybeihbcl2ijavd6vdcj4vgunw4q5qt5itmumxw7iy7fhoqfsuvkpkqeq",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

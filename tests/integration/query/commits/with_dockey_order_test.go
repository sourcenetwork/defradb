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
	test := testUtils.TestCase{
		Description: "Simple all commits query with dockey, order height desc",
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
						commits(dockey: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7", order: {height: DESC}) {
							cid
							height
						}
					}`,
				Results: []map[string]any{
					{
						"cid":    "bafybeift3qzwhklfpkgszvrmbfzb6zp3g3cqryhjkuaoz3kp2yrj763jce",
						"height": int64(2),
					},
					{
						"cid":    "bafybeige35bkafoez4cf4v6hgdkm5iaqcuqfq4bkyt7fxeycbdnqtbr7g4",
						"height": int64(2),
					},
					{
						"cid":    "bafybeiazsz3twea2uxpen6452qqa7qnzp2xildfxliidhqk632jpvbixkm",
						"height": int64(1),
					},
					{
						"cid":    "bafybeidzukbs36cwwhab4rkpi6jfhhxse2vjtc5tf767qda5valcinilmy",
						"height": int64(1),
					},
					{
						"cid":    "bafybeihbcl2ijavd6vdcj4vgunw4q5qt5itmumxw7iy7fhoqfsuvkpkqeq",
						"height": int64(1),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithDockeyAndOrderHeightAsc(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with dockey, order height asc",
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
						commits(dockey: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7", order: {height: ASC}) {
							cid
							height
						}
					}`,
				Results: []map[string]any{
					{
						"cid":    "bafybeiazsz3twea2uxpen6452qqa7qnzp2xildfxliidhqk632jpvbixkm",
						"height": int64(1),
					},
					{
						"cid":    "bafybeidzukbs36cwwhab4rkpi6jfhhxse2vjtc5tf767qda5valcinilmy",
						"height": int64(1),
					},
					{
						"cid":    "bafybeihbcl2ijavd6vdcj4vgunw4q5qt5itmumxw7iy7fhoqfsuvkpkqeq",
						"height": int64(1),
					},
					{
						"cid":    "bafybeift3qzwhklfpkgszvrmbfzb6zp3g3cqryhjkuaoz3kp2yrj763jce",
						"height": int64(2),
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

func TestQueryCommitsWithDockeyAndOrderCidDesc(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with dockey, order cid desc",
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
						commits(dockey: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7", order: {cid: DESC}) {
							cid
							height
						}
					}`,
				Results: []map[string]any{
					{
						"cid":    "bafybeihbcl2ijavd6vdcj4vgunw4q5qt5itmumxw7iy7fhoqfsuvkpkqeq",
						"height": int64(1),
					},
					{
						"cid":    "bafybeige35bkafoez4cf4v6hgdkm5iaqcuqfq4bkyt7fxeycbdnqtbr7g4",
						"height": int64(2),
					},
					{
						"cid":    "bafybeift3qzwhklfpkgszvrmbfzb6zp3g3cqryhjkuaoz3kp2yrj763jce",
						"height": int64(2),
					},
					{
						"cid":    "bafybeidzukbs36cwwhab4rkpi6jfhhxse2vjtc5tf767qda5valcinilmy",
						"height": int64(1),
					},
					{
						"cid":    "bafybeiazsz3twea2uxpen6452qqa7qnzp2xildfxliidhqk632jpvbixkm",
						"height": int64(1),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithDockeyAndOrderCidAsc(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with dockey, order cid asc",
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
						commits(dockey: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7", order: {cid: ASC}) {
							cid
							height
						}
					}`,
				Results: []map[string]any{
					{
						"cid":    "bafybeiazsz3twea2uxpen6452qqa7qnzp2xildfxliidhqk632jpvbixkm",
						"height": int64(1),
					},
					{
						"cid":    "bafybeidzukbs36cwwhab4rkpi6jfhhxse2vjtc5tf767qda5valcinilmy",
						"height": int64(1),
					},
					{
						"cid":    "bafybeift3qzwhklfpkgszvrmbfzb6zp3g3cqryhjkuaoz3kp2yrj763jce",
						"height": int64(2),
					},
					{
						"cid":    "bafybeige35bkafoez4cf4v6hgdkm5iaqcuqfq4bkyt7fxeycbdnqtbr7g4",
						"height": int64(2),
					},
					{
						"cid":    "bafybeihbcl2ijavd6vdcj4vgunw4q5qt5itmumxw7iy7fhoqfsuvkpkqeq",
						"height": int64(1),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithDockeyAndOrderAndMultiUpdatesCidAsc(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with dockey, multiple updates with order cid asc",
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
						 commits(dockey: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7", order: {height: ASC}) {
							 cid
							 height
						 }
					 }`,
				Results: []map[string]any{
					{
						"cid":    "bafybeiazsz3twea2uxpen6452qqa7qnzp2xildfxliidhqk632jpvbixkm",
						"height": int64(1),
					},
					{
						"cid":    "bafybeidzukbs36cwwhab4rkpi6jfhhxse2vjtc5tf767qda5valcinilmy",
						"height": int64(1),
					},
					{
						"cid":    "bafybeihbcl2ijavd6vdcj4vgunw4q5qt5itmumxw7iy7fhoqfsuvkpkqeq",
						"height": int64(1),
					},
					{
						"cid":    "bafybeift3qzwhklfpkgszvrmbfzb6zp3g3cqryhjkuaoz3kp2yrj763jce",
						"height": int64(2),
					},
					{
						"cid":    "bafybeige35bkafoez4cf4v6hgdkm5iaqcuqfq4bkyt7fxeycbdnqtbr7g4",
						"height": int64(2),
					},
					{
						"cid":    "bafybeietneua73vfrkvfefw5rmju7yhee6rywychdbls5xqmqtqmzfckzq",
						"height": int64(3),
					},
					{
						"cid":    "bafybeidwgrk2xyu25pmwvpkfs4hnswtgej6gopkf26jrgm6lpbofa3rs3e",
						"height": int64(3),
					},
					{
						"cid":    "bafybeiahvakoy5joy563em7hlzvqcarxqdp2nin4gnxythoj4fqjh7umzu",
						"height": int64(4),
					},
					{
						"cid":    "bafybeighft6vokgntjvpirwdt233xizmnhxtawiqeahwypxv7u26dwseoe",
						"height": int64(4),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
